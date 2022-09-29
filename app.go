package grpcapp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/caarlos0/env/v6"
	"github.com/golang-jwt/jwt"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcCtxTags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// App interface.
type App interface {

	// Start the application.
	Start()

	// Tools returns Tools.
	Tools() Tools

	// GrpcServer returns gRPC server.
	GrpcServer() *grpc.Server

	// HttpServer return HTTP server.
	HttpServer() *http.Server
}

// Option interface.
type Option interface {
	option(*app)
}

// Implementation interface.
type Implementation interface {

	// UseTools within service implementation.
	UseTools(Tools)
}

// Tools interface.
type Tools interface {

	// Config provided on application init.
	Config() *Config

	// DB connection if initialized or nil.
	DB() *pgx.Conn

	// Logger if provided of application init or nil.
	Logger() *zap.Logger

	// JwtToken from context or nil.
	JwtToken(ctx context.Context) *jwt.Token

	// JwtClaims from context or nil.
	JwtClaims(ctx context.Context) jwt.MapClaims
}

// Config of the application.
type Config struct {

	// DatabaseDSN from env.
	DatabaseDSN string `env:"DATABASE_DSN"`

	// LogLevel from env (default "info").
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	// GrpcListenPort from environment (default 9000).
	GrpcListenPort int `env:"GRPC_LISTEN_PORT" envDefault:"9000"`

	// HttpListenPort from environment (default 8080).
	HttpListenPort int `env:"HTTP_LISTEN_PORT" envDefault:"8080"`

	// TLSCertificate file from environment.
	TLSCertificate string `env:"TLS_CERTIFICATE"`

	// TLSKey file from environment.
	TLSKey string `env:"TLS_KEY"`
}

type StartHook func(App) error

// Start shortcut to New().Start().
func Start(options ...Option) {
	New(options...).Start()
}

// New app generator.
func New(options ...Option) App {
	a := &app{
		tools: &tools{},
		done:  make(chan struct{}),
	}
	for _, o := range options {
		o.option(a)
	}
	return a
}

const (
	// TokenContextKey defined value key of JWT token within context.
	TokenContextKey = "jwt_token"
)

type app struct {
	serviceImplementations []serviceImplementation
	serverOptions          []grpc.ServerOption
	unaryInterceptors      []grpc.UnaryServerInterceptor
	streamInterceptors     []grpc.StreamServerInterceptor
	tools                  *tools
	serveHttp              bool
	grpcServer             *grpc.Server
	httpServer             *http.Server
	startHooks             []StartHook
	done                   chan struct{}
	shutdownCh             chan os.Signal
}

func (a *app) Start() {
	// read environment configuration
	a.initConfig()

	// initialize zap logger
	a.initLogger()

	// note: any error prior this line will panic,
	//       logger.Fatal will be called downwards

	// initialize database
	a.initDatabase()

	// initialize servers
	a.initServers()

	// initialize service implementations
	a.initServiceImplementations()

	// start servers
	a.listen()

	// listen to "shutdown" signals
	a.shutdown()

	// wait
	<-a.done
}

func (a *app) Tools() Tools {
	return a.tools
}

func (a *app) GrpcServer() *grpc.Server {
	return a.grpcServer
}

func (a *app) HttpServer() *http.Server {
	return a.httpServer
}

func (a *app) initConfig() {
	if a.tools.cfg == nil {
		a.tools.cfg = new(Config)
		if err := env.Parse(a.tools.cfg); err != nil {
			panic("failed to parse config:" + err.Error())
		}
	}
}

func (a *app) initLogger() {
	if a.tools.log == nil {
		lvl, err := zapcore.ParseLevel(a.tools.cfg.LogLevel)
		if err != nil {
			panic("invalid log level: " + a.tools.cfg.LogLevel)
		}
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(lvl)
		if lvl != zapcore.DebugLevel {
			cfg.DisableCaller = true
			cfg.DisableStacktrace = true
		}

		a.tools.log, err = cfg.Build()
		if err != nil {
			panic("failed to build logger config: " + err.Error())
		}
	}
}

func (a *app) initServiceImplementations() {
	if len(a.serviceImplementations) > 0 {
		for _, si := range a.serviceImplementations {
			desc, impl := si.desc, si.impl
			impl.UseTools(a.tools)
			a.grpcServer.RegisterService(desc, impl)
			a.tools.log.Info("service registration",
				zap.String("serviceName", desc.ServiceName))
		}
	}
}

func (a *app) initDatabase() {
	if a.tools.db == nil && a.tools.cfg.DatabaseDSN != "" {
		var err error
		a.tools.db, err = pgx.Connect(context.Background(), a.tools.cfg.DatabaseDSN)
		if err != nil {
			a.tools.log.Fatal("failed to connect to database",
				zap.Error(err))
		}
	}
}

func (a *app) initServers() {
	if a.grpcServer == nil {
		opts := []grpcZap.Option{
			grpcZap.WithLevels(func(code codes.Code) zapcore.Level {
				switch code {
				case codes.OK:
					return zapcore.DebugLevel
				case codes.Canceled:
					return zapcore.ErrorLevel
				case codes.Unknown:
					return zapcore.ErrorLevel
				case codes.InvalidArgument:
					return zapcore.ErrorLevel
				case codes.DeadlineExceeded:
					return zapcore.ErrorLevel
				case codes.NotFound:
					return zapcore.ErrorLevel
				case codes.AlreadyExists:
					return zapcore.ErrorLevel
				case codes.PermissionDenied:
					return zapcore.ErrorLevel
				case codes.ResourceExhausted:
					return zapcore.ErrorLevel
				case codes.FailedPrecondition:
					return zapcore.ErrorLevel
				case codes.Aborted:
					return zapcore.ErrorLevel
				case codes.OutOfRange:
					return zapcore.ErrorLevel
				case codes.Unimplemented:
					return zapcore.ErrorLevel
				case codes.Internal:
					return zapcore.ErrorLevel
				case codes.Unavailable:
					return zapcore.ErrorLevel
				case codes.DataLoss:
					return zapcore.ErrorLevel
				case codes.Unauthenticated:
					return zapcore.ErrorLevel
				}
				return zapcore.ErrorLevel
			}),
		}
		unaryInterceptors := append([]grpc.UnaryServerInterceptor{
			grpcCtxTags.UnaryServerInterceptor(grpcCtxTags.WithFieldExtractor(grpcCtxTags.CodeGenRequestFieldExtractor)),
			grpcZap.UnaryServerInterceptor(a.tools.log, opts...),
		}, a.unaryInterceptors...)
		streamInterceptors := append([]grpc.StreamServerInterceptor{
			grpcCtxTags.StreamServerInterceptor(grpcCtxTags.WithFieldExtractor(grpcCtxTags.CodeGenRequestFieldExtractor)),
			grpcZap.StreamServerInterceptor(a.tools.log, opts...),
		}, a.streamInterceptors...)
		if a.tools.jwt != nil && a.tools.jwt.keyFunc != nil {
			ui, si := makeJwtInterceptors(a.tools)
			unaryInterceptors = append(unaryInterceptors, ui)
			streamInterceptors = append(streamInterceptors, si)
		}
		a.serverOptions = append(a.serverOptions,
			grpcMiddleware.WithUnaryServerChain(unaryInterceptors...),
			grpcMiddleware.WithStreamServerChain(streamInterceptors...),
		)
		a.grpcServer = grpc.NewServer(a.serverOptions...)
	}
	if a.httpServer == nil {
		a.httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", a.tools.cfg.HttpListenPort),
			Handler: a.grpcServer,
		}
	} else {
		a.httpServer.Handler = a.grpcServer
	}
}

func (a *app) listen() {
	go a.listenGrpc()
	if a.serveHttp {
		go a.listenHttp()
	}
}

func (a *app) listenGrpc() {
	addr := fmt.Sprintf(":%d", a.tools.cfg.GrpcListenPort)
	a.tools.log.Info("starting grpc server",
		zap.String("address", addr))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		a.tools.log.Fatal("failed to listen grpc",
			zap.String("address", addr),
			zap.Error(err))
	}
	if err = a.grpcServer.Serve(lis); err != nil {
		a.tools.log.Fatal("failed to serve grpc",
			zap.Error(err))
	}
}

func (a *app) listenHttp() {
	a.tools.log.Info("starting http server",
		zap.String("address", a.httpServer.Addr),
		zap.String("tlsCertificate", a.tools.cfg.TLSCertificate),
		zap.String("tlsKey", a.tools.cfg.TLSKey))
	if a.tools.cfg.TLSCertificate == "" {
		a.tools.log.Fatal("cannot start http server without tls certificate")
	}
	if a.tools.cfg.TLSKey == "" {
		a.tools.log.Fatal("cannot start http server without tls key")
	}
	if err := a.httpServer.ListenAndServeTLS(
		a.tools.cfg.TLSCertificate,
		a.tools.cfg.TLSKey,
	); err != nil && err != http.ErrServerClosed {
		a.tools.log.Fatal("failed to serve http",
			zap.Error(err))
	}
}

func (a *app) shutdown() {
	go func() {
		a.shutdownCh = make(chan os.Signal)
		signal.Notify(a.shutdownCh, os.Interrupt, os.Kill)
		sig := <-a.shutdownCh
		a.tools.log.Info("graceful shutdown",
			zap.String("signal", sig.String()))

		// stop grpc server
		a.grpcServer.GracefulStop()
		a.tools.log.Info("stopped grpc server")

		// stop http server (optionally)
		if a.serveHttp {
			if err := a.httpServer.Shutdown(context.Background()); err != nil {
				a.tools.log.Error("failed to shutdown http server gracefully",
					zap.Error(err))
			}
			a.tools.log.Info("stopped http server")
		}
		a.done <- struct{}{}
		close(a.done)
	}()
}

type tools struct {
	cfg *Config
	log *zap.Logger
	db  *pgx.Conn
	jwt *jwtData
}

// Config provided on application init.
func (t *tools) Config() *Config {
	return t.cfg
}

// DB connection if initialized or nil.
func (t *tools) DB() *pgx.Conn {
	return t.db
}

// Logger if provided of application init or nil.
func (t *tools) Logger() *zap.Logger {
	return t.log
}

// JwtToken from context or nil.
func (t *tools) JwtToken(ctx context.Context) *jwt.Token {
	token := ctx.Value(TokenContextKey)
	if token != nil {
		if v, ok := token.(*jwt.Token); ok {
			return v
		}
	}
	return nil
}

// JwtClaims from context or nil.
func (t *tools) JwtClaims(ctx context.Context) jwt.MapClaims {
	if token := t.JwtToken(ctx); token != nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			return claims
		}
	}
	return nil
}

// WithConfig replaces the default Config. Environment variables will not be parsed.
func WithConfig(cfg *Config) Option {
	return &configOption{cfg}
}

type configOption struct {
	cfg *Config
}

func (opt *configOption) option(a *app) {
	a.tools.cfg = opt.cfg
}

// WithLogger replaces the default logger.
func WithLogger(log *zap.Logger) Option {
	return &loggerOption{log}
}

type loggerOption struct{ log *zap.Logger }

func (opt *loggerOption) option(a *app) {
	a.tools.log = opt.log
}

// WithServiceImplementation appends service implementation to grpc.Server, regardless
// if WithGrpcServer options was used or not.
func WithServiceImplementation(desc *grpc.ServiceDesc, impl Implementation) Option {
	return &serviceImplementationOption{desc, impl}
}

type serviceImplementationOption struct {
	desc *grpc.ServiceDesc
	impl Implementation
}

type serviceImplementation struct {
	desc *grpc.ServiceDesc
	impl Implementation
}

func (opt *serviceImplementationOption) option(a *app) {
	a.serviceImplementations = append(a.serviceImplementations, serviceImplementation{
		desc: opt.desc,
		impl: opt.impl,
	})
}

// WithGrpcServerOptions appends grpc.ServerOption to default gRPC server.
func WithGrpcServerOptions(options ...grpc.ServerOption) Option {
	return &grpcServerOptionsOption{options}
}

type grpcServerOptionsOption struct {
	options []grpc.ServerOption
}

func (opt *grpcServerOptionsOption) option(a *app) {
	a.serverOptions = opt.options
}

// WithDatabase replaces the default database connection. If used - the app will not
// try to connect using DatabaseDSN provided in Config.
func WithDatabase(db *pgx.Conn) Option {
	return &databaseOption{db}
}

type databaseOption struct {
	db *pgx.Conn
}

func (opt *databaseOption) option(a *app) {
	a.tools.db = opt.db
}

// WithHTTP option enables http server to listen in addition to gRPC server.
// When enabled TLSCertificate and TLSKey in Config must be provided.
func WithHTTP() Option {
	return new(httpOption)
}

type httpOption struct{}

func (*httpOption) option(a *app) {
	a.serveHttp = true
}

// WithGrpcServer replaces the default grpc.Server in app by the one provided
// in arguments. When used, all gRPC server related options will be ignored,
// like WithJwtAuthentication, WithGrpcServerOptions, WithUnaryInterceptor etc.
func WithGrpcServer(srv *grpc.Server) Option {
	return &grpcServerOption{srv}
}

type grpcServerOption struct {
	srv *grpc.Server
}

func (opt *grpcServerOption) option(a *app) {
	a.grpcServer = opt.srv
}

// WithHttpServer replaces the default http.Server. Must be used together
// with WithHTTP option, otherwise will be ignored.
func WithHttpServer(srv *http.Server) Option {
	return &httpServerOption{srv}
}

type httpServerOption struct {
	srv *http.Server
}

func (opt *httpServerOption) option(a *app) {
	a.httpServer = opt.srv
}

// WithJwtAuthentication enables JWT authentication for provided methods. If not methods
// provided, the authentication will be enabled for all requests.
func WithJwtAuthentication(keyFunc jwt.Keyfunc, methods ...string) Option {
	return &jwtAuthOption{keyFunc, methods}
}

type jwtAuthOption struct {
	keyFunc jwt.Keyfunc
	methods []string
}

func (opt *jwtAuthOption) option(a *app) {
	a.tools.jwt = &jwtData{opt.keyFunc, opt.methods}
}

type jwtData struct {
	keyFunc jwt.Keyfunc
	methods []string
}

func makeJwtInterceptors(t *tools) (
	grpc.UnaryServerInterceptor,
	grpc.StreamServerInterceptor,
) {
	const (
		authHeader  = "authorization"
		tokenHeader = TokenContextKey
	)

	errUnauthorized := status.Error(codes.Unauthenticated, "unauthenticated")
	checkMethods := len(t.jwt.methods) > 0
	useMethods := make(map[string]struct{}, len(t.jwt.methods))

	for _, method := range t.jwt.methods {
		useMethods[method] = struct{}{}
	}

	getToken := func(ctx context.Context) (*jwt.Token, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			t.log.Debug("failed to extract metadata from context")
			return nil, errUnauthorized
		}
		values := md[authHeader]
		if len(values) == 0 {
			t.log.Debug("authorization token not found in metadata",
				zap.Any("md", md))
			return nil, errUnauthorized
		}
		token, err := jwt.Parse(values[0], t.jwt.keyFunc)
		if err != nil {
			t.log.Warn("failed to validate token",
				zap.Error(err),
				zap.Any("md", md))
			return nil, errUnauthorized
		}
		return token, nil
	}

	unaryInterceptor := func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if checkMethods {
			if _, ok := useMethods[info.FullMethod]; !ok {
				return handler(ctx, req)
			}
		}
		token, err := getToken(ctx)
		if err != nil {
			return nil, err
		}
		return handler(context.WithValue(ctx, tokenHeader, token), req)
	}

	streamInterceptor := func(
		srv any,
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if checkMethods {
			if _, ok := useMethods[info.FullMethod]; !ok {
				return handler(srv, stream)
			}
		}
		token, err := getToken(stream.Context())
		if err != nil {
			return err
		}
		return handler(srv, &grpcStreamWrapper{
			ctx:    context.WithValue(stream.Context(), tokenHeader, token),
			stream: stream,
		})
	}

	return unaryInterceptor, streamInterceptor
}

type grpcStreamWrapper struct {
	ctx    context.Context
	stream grpc.ServerStream
}

func (sw *grpcStreamWrapper) SetHeader(md metadata.MD) error {
	return sw.stream.SetHeader(md)
}

func (sw *grpcStreamWrapper) SendHeader(md metadata.MD) error {
	return sw.stream.SendHeader(md)
}

func (sw *grpcStreamWrapper) SetTrailer(md metadata.MD) {
	sw.stream.SetTrailer(md)
}

func (sw *grpcStreamWrapper) Context() context.Context {
	return sw.ctx
}

func (sw *grpcStreamWrapper) SendMsg(m any) error {
	return sw.stream.SendMsg(m)
}

func (sw *grpcStreamWrapper) RecvMsg(m any) error {
	return sw.stream.RecvMsg(m)
}

// WithUnaryInterceptor appends grpc.UnaryServerInterceptor using WithUnaryServerChain
// middleware after Logger and JWT interceptors. This option is ignored if
// WithGrpcServer option is used.
func WithUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) Option {
	return &unaryInterceptorOption{interceptor}
}

type unaryInterceptorOption struct {
	interceptor grpc.UnaryServerInterceptor
}

func (opt *unaryInterceptorOption) option(a *app) {
	a.unaryInterceptors = append(a.unaryInterceptors, opt.interceptor)
}

// WithStreamInterceptor appends grpc.StreamServerInterceptor using WithStreamServerChain
// middleware after Logger and JWT interceptors. This option is ignored if
// WithGrpcServer option is used.
func WithStreamInterceptor(interceptor grpc.StreamServerInterceptor) Option {
	return &streamInterceptorOption{interceptor}
}

type streamInterceptorOption struct {
	interceptor grpc.StreamServerInterceptor
}

func (opt *streamInterceptorOption) option(a *app) {
	a.streamInterceptors = append(a.streamInterceptors, opt.interceptor)
}

// WithStartHook appends StartHook to run before application starts listening.
func WithStartHook(hook StartHook) Option {
	return &startHookOption{hook}
}

type startHookOption struct {
	hook StartHook
}

func (opt *startHookOption) option(a *app) {
	if opt.hook != nil {
		a.startHooks = append(a.startHooks, opt.hook)
	}
}

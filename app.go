package grpcapp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/caarlos0/env/v6"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// App interface.
type App interface {
	Start()
}

// Option interface.
type Option interface {
	option(*app)
}

// Implementation interface.
type Implementation interface {
	UseTools(Tools)
}

// Tools interface.
type Tools interface {
	Config() *Config
	DB() *pgx.Conn
	Log() *zap.Logger
}

// Config of the application.
type Config struct {
	DatabaseDSN    string `env:"DATABASE_DSN"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	GrpcListenPort int    `env:"GRPC_LISTEN_PORT" envDefault:"9000"`
	HttpListenPort int    `env:"HTTP_LISTEN_PORT" envDefault:"8080"`
	TLSCertificate string `env:"TLS_CERTIFICATE"`
	TLSKey         string `env:"TLS_KEY"`
}

// Start shortcut to New().Listen().
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

type app struct {
	serviceImplementations []serviceImplementation
	serverOptions          []grpc.ServerOption
	tools                  *tools
	serveHttp              bool
	grpcServer             *grpc.Server
	httpServer             *http.Server
	done                   chan struct{}
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
	if a.tools.cfg.DatabaseDSN != "" {
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
		opts := []grpc_zap.Option{
			grpc_zap.WithLevels(func(code codes.Code) zapcore.Level {
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
			//grpc_zap.WithLevels(customFunc),
		}
		a.serverOptions = append(a.serverOptions,
			grpc_middleware.WithUnaryServerChain(
				grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_zap.UnaryServerInterceptor(a.tools.log, opts...),
			),
			grpc_middleware.WithStreamServerChain(
				grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_zap.StreamServerInterceptor(a.tools.log, opts...),
			),
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
		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt, os.Kill)
		sig := <-ch
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
	}()
}

type tools struct {
	cfg *Config
	log *zap.Logger
	db  *pgx.Conn
}

func (t *tools) Config() *Config {
	return t.cfg
}

func (t *tools) DB() *pgx.Conn {
	return t.db
}

func (t *tools) Log() *zap.Logger {
	return t.log
}

func WithConfig(cfg *Config) Option {
	return &configOption{cfg}
}

type configOption struct {
	cfg *Config
}

func (opt *configOption) option(a *app) {
	a.tools.cfg = opt.cfg
}

func WithLogger(log *zap.Logger) Option {
	return &loggerOption{log}
}

type loggerOption struct{ log *zap.Logger }

func (opt *loggerOption) option(a *app) {
	a.tools.log = opt.log
}

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

func WithGrpcServerOptions(options ...grpc.ServerOption) Option {
	return &grpcServerOptionsOption{options}
}

type grpcServerOptionsOption struct {
	options []grpc.ServerOption
}

func (opt *grpcServerOptionsOption) option(a *app) {
	a.serverOptions = opt.options
}

func WithDatabase(db *pgx.Conn) Option {
	return &databaseOption{db}
}

type databaseOption struct {
	db *pgx.Conn
}

func (opt *databaseOption) option(a *app) {
	a.tools.db = opt.db
}

func WithHTTP() Option {
	return new(httpOption)
}

type httpOption struct{}

func (*httpOption) option(a *app) {
	a.serveHttp = true
}

func WithGrpcServer(srv *grpc.Server) Option {
	return &grpcServerOption{srv}
}

type grpcServerOption struct {
	srv *grpc.Server
}

func (opt *grpcServerOption) option(a *app) {
	a.grpcServer = opt.srv
}

func WithHttpServer(srv *http.Server) Option {
	return &httpServerOption{srv}
}

type httpServerOption struct {
	srv *http.Server
}

func (opt *httpServerOption) option(a *app) {
	a.httpServer = opt.srv
}

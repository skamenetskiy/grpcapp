package grpcapp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App interface {
	Listen()
}

type Option interface {
	option(*app)
}

type Implementation interface {
	UseTools(Tools)
}

type Tools interface {
	DB() *pgx.Conn
	Log() *zap.Logger
}

func New(options ...Option) App {
	a := new(app)
	for _, o := range options {
		o.option(a)
	}
	return a
}

type app struct {
	dsn                   string
	dbConn                *pgx.Conn
	log                   *zap.Logger
	serviceDescription    *grpc.ServiceDesc
	serviceImplementation Implementation
	listenPort            int
	grpcServerOptions     []grpc.ServerOption
}

func (a *app) Listen() {
	var err error
	if a.log == nil {
		a.log, err = zap.NewProduction()
		if err != nil {
			panic(err)
		}
	}
	t := &tools{
		log: a.log,
	}
	if a.dbConn != nil {
		t.db = a.dbConn
	} else if a.dsn != "" {
		t.db, err = pgx.Connect(context.Background(), a.dsn)
		if err != nil {
			a.log.Fatal("failed to connect to database",
				zap.Error(err))
		}
	}
	port := a.listenPort
	if port == 0 {
		const defaultPort = 9000
		if v := os.Getenv("PORT"); v != "" {
			port, err = strconv.Atoi(v)
			if err != nil {
				a.log.Warn("failed to parse PORT variable",
					zap.Error(err))
				port = defaultPort
			}
		} else {
			port = defaultPort
		}
	}
	addr := fmt.Sprintf(":%d", port)
	a.log.Debug("start listening",
		zap.String("address", addr))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		a.log.Fatal("failed to listen", zap.Error(err))
	}
	srv := grpc.NewServer(a.grpcServerOptions...)
	if a.serviceDescription != nil {
		a.serviceImplementation.UseTools(t)
		srv.RegisterService(a.serviceDescription, a.serviceImplementation)
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go func() {
		sig := <-ch
		a.log.Info("graceful shutdown",
			zap.String("signal", sig.String()))
		srv.GracefulStop()
	}()
	a.log.Debug("start serving")
	if err = srv.Serve(lis); err != nil {
		a.log.Fatal("failed to serve", zap.Error(err))
	}
}

type tools struct {
	db  *pgx.Conn
	log *zap.Logger
}

func (t *tools) DB() *pgx.Conn {
	return t.db
}

func (t *tools) Log() *zap.Logger {
	return t.log
}

func WithLogger(log *zap.Logger) Option {
	return &loggerOption{log}
}

type loggerOption struct{ log *zap.Logger }

func (opt *loggerOption) option(a *app) {
	a.log = opt.log
}

func WithListenPort(port int) Option {
	return &listenPortOption{port}
}

type listenPortOption struct{ port int }

func (opt *listenPortOption) option(a *app) {
	a.listenPort = opt.port
}

func WithServiceImplementation(desc *grpc.ServiceDesc, impl Implementation) Option {
	return &serviceImplementationOption{desc, impl}
}

type serviceImplementationOption struct {
	desc *grpc.ServiceDesc
	impl Implementation
}

func (opt *serviceImplementationOption) option(a *app) {
	a.serviceDescription = opt.desc
	a.serviceImplementation = opt.impl
}

func WithGrpcServerOptions(options ...grpc.ServerOption) Option {
	return &grpcServerOptionsOption{options}
}

type grpcServerOptionsOption struct {
	options []grpc.ServerOption
}

func (opt *grpcServerOptionsOption) option(a *app) {
	a.grpcServerOptions = opt.options
}

func WithDatabase[T string | *pgx.Conn](db T) Option {
	opt := new(databaseOption)
	switch v := any(db).(type) {
	case string:
		opt.dsn = v
	case *pgx.Conn:
		opt.conn = v
	}
	return opt
}

type databaseOption struct {
	dsn  string
	conn *pgx.Conn
}

func (opt *databaseOption) option(a *app) {
	if opt.dsn != "" {
		a.dsn = opt.dsn
	}
	if opt.conn != nil {
		a.dbConn = opt.conn
	}
}

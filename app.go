package grpcapp

import (
	"fmt"
	"net"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App interface {
	Listen()
}

func New(options ...Option) App {
	a := new(app)
	for _, o := range options {
		o.option(a)
	}
	return a
}

type app struct {
	log                   *zap.Logger
	serviceDescription    *grpc.ServiceDesc
	serviceImplementation any
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
	port := a.listenPort
	if port == 0 {
		port = 9000
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

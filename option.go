package grpcapp

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Option interface {
	option(*app)
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

func WithServiceImplementation(desc *grpc.ServiceDesc, impl any) Option {
	return &serviceImplementationOption{desc, impl}
}

type serviceImplementationOption struct {
	desc *grpc.ServiceDesc
	impl any
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

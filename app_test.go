package grpcapp

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestNew(t *testing.T) {
	type args struct {
		options []Option
	}
	c := make(chan struct{})
	s := &http.Server{}
	o := WithHttpServer(s)
	tests := []struct {
		name string
		args args
		want App
	}{
		{"new", args{}, &app{tools: &tools{}, done: c}},
		{"with options", args{options: []Option{o}}, &app{tools: &tools{}, httpServer: s, done: c}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.options...).(*app)
			got.done = c // a hack to verify channel
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStart(t *testing.T) {
	type args struct {
		options []Option
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Start(tt.args.options...)
		})
	}
}

func TestWithConfig(t *testing.T) {
	type args struct {
		cfg *Config
	}
	c := &Config{}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"basic", args{c}, &configOption{c}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithConfig(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithDatabase(t *testing.T) {
	type args struct {
		db *pgx.Conn
	}
	c := new(pgx.Conn)
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"basic", args{c}, &databaseOption{c}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithDatabase(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDatabase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithGrpcServer(t *testing.T) {
	type args struct {
		srv *grpc.Server
	}
	s := grpc.NewServer()
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"basic", args{s}, &grpcServerOption{s}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithGrpcServer(tt.args.srv); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithGrpcServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithGrpcServerOptions(t *testing.T) {
	type args struct {
		options []grpc.ServerOption
	}
	o := &grpc.EmptyServerOption{}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"basic", args{[]grpc.ServerOption{o}}, &grpcServerOptionsOption{[]grpc.ServerOption{o}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithGrpcServerOptions(tt.args.options...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithGrpcServerOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithHTTP(t *testing.T) {
	tests := []struct {
		name string
		want Option
	}{
		{"basic", &httpOption{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithHTTP(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithHTTP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithHttpServer(t *testing.T) {
	type args struct {
		srv *http.Server
	}
	s := &http.Server{}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"basic", args{s}, &httpServerOption{s}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithHttpServer(tt.args.srv); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithHttpServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithLogger(t *testing.T) {
	type args struct {
		log *zap.Logger
	}
	l := zap.NewNop()
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"basic", args{l}, &loggerOption{l}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithLogger(tt.args.log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

type sampleImplementation struct{}

func (sampleImplementation) UseTools(_ Tools) {}

func TestWithServiceImplementation(t *testing.T) {
	d := &grpc.ServiceDesc{}
	i := &sampleImplementation{}
	type args struct {
		desc *grpc.ServiceDesc
		impl Implementation
	}
	tests := []struct {
		name string
		args args
		want Option
	}{
		{"basic", args{d, i}, &serviceImplementationOption{d, i}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithServiceImplementation(tt.args.desc, tt.args.impl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithServiceImplementation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_app_Start(t *testing.T) {
	type fields struct {
		serviceImplementations []serviceImplementation
		serverOptions          []grpc.ServerOption
		tools                  *tools
		serveHttp              bool
		grpcServer             *grpc.Server
		httpServer             *http.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				serviceImplementations: tt.fields.serviceImplementations,
				serverOptions:          tt.fields.serverOptions,
				tools:                  tt.fields.tools,
				serveHttp:              tt.fields.serveHttp,
				grpcServer:             tt.fields.grpcServer,
				httpServer:             tt.fields.httpServer,
			}
			a.Start()
		})
	}
}

func Test_app_initConfig(t *testing.T) {
	c := &Config{}
	cd := &Config{
		LogLevel:       "info",
		GrpcListenPort: 9000,
		HttpListenPort: 8080,
	}
	type fields struct {
		tools *tools
	}
	tests := []struct {
		name     string
		fields   fields
		injected bool
	}{
		{"with config", fields{&tools{cfg: c}}, true},
		{"without config", fields{&tools{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				tools: tt.fields.tools,
			}
			a.initConfig()
			if a.tools.cfg == nil {
				t.Error("config is nil")
			}
			if tt.injected && !reflect.DeepEqual(a.tools.cfg, c) {
				t.Errorf("expected %v, got %v", c, a.tools.cfg)
			} else if !tt.injected && !reflect.DeepEqual(a.tools.cfg, cd) {
				t.Errorf("expected %v, got %v", cd, a.tools.cfg)
			}
		})
	}
}

func Test_app_initDatabase(t *testing.T) {
	type fields struct {
		serviceImplementations []serviceImplementation
		serverOptions          []grpc.ServerOption
		tools                  *tools
		serveHttp              bool
		grpcServer             *grpc.Server
		httpServer             *http.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				serviceImplementations: tt.fields.serviceImplementations,
				serverOptions:          tt.fields.serverOptions,
				tools:                  tt.fields.tools,
				serveHttp:              tt.fields.serveHttp,
				grpcServer:             tt.fields.grpcServer,
				httpServer:             tt.fields.httpServer,
			}
			a.initDatabase()
		})
	}
}

func Test_app_initLogger(t *testing.T) {
	type fields struct {
		serviceImplementations []serviceImplementation
		serverOptions          []grpc.ServerOption
		tools                  *tools
		serveHttp              bool
		grpcServer             *grpc.Server
		httpServer             *http.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				serviceImplementations: tt.fields.serviceImplementations,
				serverOptions:          tt.fields.serverOptions,
				tools:                  tt.fields.tools,
				serveHttp:              tt.fields.serveHttp,
				grpcServer:             tt.fields.grpcServer,
				httpServer:             tt.fields.httpServer,
			}
			a.initLogger()
		})
	}
}

func Test_app_initServers(t *testing.T) {
	type fields struct {
		serviceImplementations []serviceImplementation
		serverOptions          []grpc.ServerOption
		tools                  *tools
		serveHttp              bool
		grpcServer             *grpc.Server
		httpServer             *http.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				serviceImplementations: tt.fields.serviceImplementations,
				serverOptions:          tt.fields.serverOptions,
				tools:                  tt.fields.tools,
				serveHttp:              tt.fields.serveHttp,
				grpcServer:             tt.fields.grpcServer,
				httpServer:             tt.fields.httpServer,
			}
			a.initServers()
		})
	}
}

func Test_app_initServiceImplementations(t *testing.T) {
	type fields struct {
		serviceImplementations []serviceImplementation
		serverOptions          []grpc.ServerOption
		tools                  *tools
		serveHttp              bool
		grpcServer             *grpc.Server
		httpServer             *http.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				serviceImplementations: tt.fields.serviceImplementations,
				serverOptions:          tt.fields.serverOptions,
				tools:                  tt.fields.tools,
				serveHttp:              tt.fields.serveHttp,
				grpcServer:             tt.fields.grpcServer,
				httpServer:             tt.fields.httpServer,
			}
			a.initServiceImplementations()
		})
	}
}

func Test_app_listen(t *testing.T) {
	type fields struct {
		serviceImplementations []serviceImplementation
		serverOptions          []grpc.ServerOption
		tools                  *tools
		serveHttp              bool
		grpcServer             *grpc.Server
		httpServer             *http.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				serviceImplementations: tt.fields.serviceImplementations,
				serverOptions:          tt.fields.serverOptions,
				tools:                  tt.fields.tools,
				serveHttp:              tt.fields.serveHttp,
				grpcServer:             tt.fields.grpcServer,
				httpServer:             tt.fields.httpServer,
			}
			a.listen()
		})
	}
}

func Test_app_listenGrpc(t *testing.T) {
	type fields struct {
		serviceImplementations []serviceImplementation
		serverOptions          []grpc.ServerOption
		tools                  *tools
		serveHttp              bool
		grpcServer             *grpc.Server
		httpServer             *http.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				serviceImplementations: tt.fields.serviceImplementations,
				serverOptions:          tt.fields.serverOptions,
				tools:                  tt.fields.tools,
				serveHttp:              tt.fields.serveHttp,
				grpcServer:             tt.fields.grpcServer,
				httpServer:             tt.fields.httpServer,
			}
			a.listenGrpc()
		})
	}
}

func Test_app_listenHttp(t *testing.T) {
	type fields struct {
		serviceImplementations []serviceImplementation
		serverOptions          []grpc.ServerOption
		tools                  *tools
		serveHttp              bool
		grpcServer             *grpc.Server
		httpServer             *http.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				serviceImplementations: tt.fields.serviceImplementations,
				serverOptions:          tt.fields.serverOptions,
				tools:                  tt.fields.tools,
				serveHttp:              tt.fields.serveHttp,
				grpcServer:             tt.fields.grpcServer,
				httpServer:             tt.fields.httpServer,
			}
			a.listenHttp()
		})
	}
}

func Test_app_shutdown(t *testing.T) {
	type fields struct {
		serviceImplementations []serviceImplementation
		serverOptions          []grpc.ServerOption
		tools                  *tools
		serveHttp              bool
		grpcServer             *grpc.Server
		httpServer             *http.Server
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				serviceImplementations: tt.fields.serviceImplementations,
				serverOptions:          tt.fields.serverOptions,
				tools:                  tt.fields.tools,
				serveHttp:              tt.fields.serveHttp,
				grpcServer:             tt.fields.grpcServer,
				httpServer:             tt.fields.httpServer,
			}
			a.shutdown()
		})
	}
}

func Test_configOption_option(t *testing.T) {
	type fields struct {
		cfg *Config
	}
	type args struct {
		a *app
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &configOption{
				cfg: tt.fields.cfg,
			}
			opt.option(tt.args.a)
		})
	}
}

func Test_databaseOption_option(t *testing.T) {
	type fields struct {
		db *pgx.Conn
	}
	type args struct {
		a *app
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &databaseOption{
				db: tt.fields.db,
			}
			opt.option(tt.args.a)
		})
	}
}

func Test_grpcServerOption_option(t *testing.T) {
	type fields struct {
		srv *grpc.Server
	}
	type args struct {
		a *app
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &grpcServerOption{
				srv: tt.fields.srv,
			}
			opt.option(tt.args.a)
		})
	}
}

func Test_grpcServerOptionsOption_option(t *testing.T) {
	type fields struct {
		options []grpc.ServerOption
	}
	type args struct {
		a *app
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &grpcServerOptionsOption{
				options: tt.fields.options,
			}
			opt.option(tt.args.a)
		})
	}
}

func Test_httpOption_option(t *testing.T) {
	type args struct {
		a *app
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ht := &httpOption{}
			ht.option(tt.args.a)
		})
	}
}

func Test_httpServerOption_option(t *testing.T) {
	type fields struct {
		srv *http.Server
	}
	type args struct {
		a *app
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &httpServerOption{
				srv: tt.fields.srv,
			}
			opt.option(tt.args.a)
		})
	}
}

func Test_loggerOption_option(t *testing.T) {
	type fields struct {
		log *zap.Logger
	}
	type args struct {
		a *app
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &loggerOption{
				log: tt.fields.log,
			}
			opt.option(tt.args.a)
		})
	}
}

func Test_serviceImplementationOption_option(t *testing.T) {
	type fields struct {
		desc *grpc.ServiceDesc
		impl Implementation
	}
	type args struct {
		a *app
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &serviceImplementationOption{
				desc: tt.fields.desc,
				impl: tt.fields.impl,
			}
			opt.option(tt.args.a)
		})
	}
}

func Test_tools_Config(t1 *testing.T) {
	type fields struct {
		cfg *Config
		log *zap.Logger
		db  *pgx.Conn
	}
	tests := []struct {
		name   string
		fields fields
		want   *Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &tools{
				cfg: tt.fields.cfg,
				log: tt.fields.log,
				db:  tt.fields.db,
			}
			if got := t.Config(); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Config() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tools_DB(t1 *testing.T) {
	type fields struct {
		cfg *Config
		log *zap.Logger
		db  *pgx.Conn
	}
	tests := []struct {
		name   string
		fields fields
		want   *pgx.Conn
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &tools{
				cfg: tt.fields.cfg,
				log: tt.fields.log,
				db:  tt.fields.db,
			}
			if got := t.DB(); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("DB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tools_Log(t1 *testing.T) {
	type fields struct {
		cfg *Config
		log *zap.Logger
		db  *pgx.Conn
	}
	tests := []struct {
		name   string
		fields fields
		want   *zap.Logger
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &tools{
				cfg: tt.fields.cfg,
				log: tt.fields.log,
				db:  tt.fields.db,
			}
			if got := t.Log(); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Log() = %v, want %v", got, tt.want)
			}
		})
	}
}

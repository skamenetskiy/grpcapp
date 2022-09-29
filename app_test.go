package grpcapp

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	cert := []byte(`-----BEGIN CERTIFICATE-----
MIIEGDCCAoCgAwIBAgIRAMlqvQCoNvt6729Qa1kGP9wwDQYJKoZIhvcNAQELBQAw
YTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMRswGQYDVQQLDBJzaW1v
bkBkZWxsIChTaW1vbikxIjAgBgNVBAMMGW1rY2VydCBzaW1vbkBkZWxsIChTaW1v
bikwHhcNMjIwOTI4MTA0NjMwWhcNMjQxMjI4MTA0NjMwWjBGMScwJQYDVQQKEx5t
a2NlcnQgZGV2ZWxvcG1lbnQgY2VydGlmaWNhdGUxGzAZBgNVBAsMEnNpbW9uQGRl
bGwgKFNpbW9uKTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALSXDm4j
1l/roOiEyxqgFkqHDsPM5lCfEqYAgeblfyJju154Vzwb8XuBiPSbgZciTI51UgBD
LmeurAfrvR1xcjDVZ3F60uNtVLpvCNUykXpoaizKsOMP2utkhpYQEU/ssgRiKoAe
YpDIJ6+hhx9N4dT6SGn7IC7im9ffDsjrBzvR4y18BlszTMnUaWXZlKn5zaw0/pTM
+nqptMoaZugXkgNPK8CYiRsVuzacek6MCHwjlI2n+zb9/MrIpzFRCA+r0nLxEV0+
a5wqqAsk9m5IeqaropKxe36Lo9//ciHo4UN/GEY0h87ijbfEXaZnnUv8pRCTmScb
c4Vm+04VhZZHLfsCAwEAAaNmMGQwDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoG
CCsGAQUFBwMBMB8GA1UdIwQYMBaAFOLCY4DBNHfmhlzRdKZYFcHQw2OsMBwGA1Ud
EQQVMBOCEWxrLmthbWVuZXRza2lpLnJ1MA0GCSqGSIb3DQEBCwUAA4IBgQAVPk3j
NsmdmYyjXpUxxT4BTc+A+qynhrEMwgVjnJlNzsToz9R3+JL5RNIx9UIAtnJNVY5d
SbJcCwYFQBws3Gof+hfVsY3nxLo9yrFBBBm2Tj7dPHx9AVjsQ8QM5eQDv65x/7Zr
j/LYi96QEMuwFNyI7jew4CCG0VVoWdp1AkdBB35u0wyw0bYJEC22Crgs12goDXVq
snA3PszKy9O34AMCmkBlSIf4zDrbCdvwZtcZhT/Q5AkxQ07NBb6ud1G7c49n9kln
nAWhGAsw/63yWJctd//D9347NXw/lXJmDSq9Bhp1y49vf82/IMD/sidsJTvqHY63
rVlw0jC9lHnOz7BB71uVDQ4xM6yabFheLVxqRdUxU32ShNDPHfHhWsu7rSvtBPJE
UusFxCTdW29wS44pDnP2eRT9e2f9FeDgJRPAGbT46DLvpguKa7rvUwGCsJS4GQsx
/mZ8WE+doy2t3yQFXPrIaXaKYaSszX7jNXizcf6Hqq3sTbe8ntd9Siv1N5s=
-----END CERTIFICATE-----`)
	key := []byte(`-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC0lw5uI9Zf66Do
hMsaoBZKhw7DzOZQnxKmAIHm5X8iY7teeFc8G/F7gYj0m4GXIkyOdVIAQy5nrqwH
670dcXIw1WdxetLjbVS6bwjVMpF6aGosyrDjD9rrZIaWEBFP7LIEYiqAHmKQyCev
oYcfTeHU+khp+yAu4pvX3w7I6wc70eMtfAZbM0zJ1Gll2ZSp+c2sNP6UzPp6qbTK
GmboF5IDTyvAmIkbFbs2nHpOjAh8I5SNp/s2/fzKyKcxUQgPq9Jy8RFdPmucKqgL
JPZuSHqmq6KSsXt+i6Pf/3Ih6OFDfxhGNIfO4o23xF2mZ51L/KUQk5knG3OFZvtO
FYWWRy37AgMBAAECggEAJRCX4p0yY6+N8AtJUGapDJTZv/AvGT786dtSzhwuUtWb
YFFcvjaSAqJchK/iEi30/owvz2P6g0dDgcCtqPxezo0OVSk7XXhUGoutiWx+lVwW
5qiXU5MugH+6a9RSaVAQXXv0cyVJX6PlGVQS0qb+geL4t9/WBCl4iP78Htq2Ol2K
SuvVfK6nFicNRsVFm9d/OmkWZEGRgDMnJwic37UpG49FlOPjuKeM/WG5c6hiMj5g
g3Gu+4CrRqrDEyQTdgjaCwVO+cYJibkWNZXFV7cC7AXB/0P2DGAy5mggGsC4Q+UG
XWN1ecWHg73OwBDDfoJ8e3ZAuFEMZ0buFX1W9g4McQKBgQDlhSP6EybxFLjdVjyA
2IsC4wvWGyNUazoJ1WWKmq43vW1zybgEzV+U2JYQKnoh0G9QNYaxC5Byyf8jPL0V
RRlyXS/p60N9P7vYlrEDSNYwRf7f7+wtFADdZHcSBYcQ7Z8rN+U9HZJ0aIYDbLwr
sDQYrmNYGK4dl2sLKSkx4O89qQKBgQDJbMVvtssGj9I6PuD1HgQFYx+9su5NR71U
svlqbXX5cXoWIb9LVI7RQdzybA1fzDgRiDhi8iEI04JcDAqjyYThUKIJaT6G6E51
CjdpBYECXBgeouK+L1kFM1zbw61VJfwM6Fz6d9rk4+CvuDkWr/BOdEsCfLjhERKn
y6LCcsTtAwKBgQCdjCjH7gGbFrhW5m0dnIa/co6bZ4F23yu1uE+9NrRD+rl484xn
b3oeuBU5/45aS7M9AaD1QpTi1plV3MmGIip3gFP1Y9Kt1OPipn3XXVX5SPLNUOlz
f/mf/uhk7HpsOlA54GJw8y2mzmC/VRJNguQf1QTIYhiSo2+M97IZVOekEQKBgDZe
j+SZuK+qvppOQranRXqWyQiRddWSWb61GLHrnf6Y7NVwgow45NwDJTqig/Gp1DCX
TnEW3mfdf8CM14piaOXQuAxGRkRwDE13VoGYpLwYU8JhQUcIzMSkmpoPdYgYWrK5
Pe+1znYeNJX56h7/mqPyrBSdyeGmlByK0QIfrJw1AoGAG6uedjIidyBarhAk7blf
j7wIWmYKuSaE9hYYLoEJtJfybag+HJfeYnZ005d4hVukDJR1p2v25dbey+YB9dKn
YQ1Haf0Ff5MgdzAoQF2TdYqQ7GlwnHQYN3skLRt7+0ornICsSow9xqL6dSg4E6Qg
meCGtVbjncVVuYP072iINtA=
-----END PRIVATE KEY-----`)
	tmpDir := os.TempDir()
	certFile := filepath.Join(tmpDir, "tls.crt")
	keyFile := filepath.Join(tmpDir, "tls.key")

	_ = os.WriteFile(certFile, cert, 0644)
	_ = os.WriteFile(keyFile, key, 0644)
	defer func() {
		_ = os.Remove(certFile)
		_ = os.Remove(keyFile)
	}()

	type fields struct {
		tools      *tools
		serveHttp  bool
		httpServer *http.Server
	}
	tests := []struct {
		name        string
		fields      fields
		fatalHook   *fatalHook
		expectFatal bool
	}{
		{
			"one",
			fields{
				&tools{
					cfg: &Config{
						TLSKey:         keyFile,
						TLSCertificate: certFile,
					},
				},
				true,
				&http.Server{
					Addr: ":20808",
				},
			},
			new(fatalHook),
			false,
		},
		{
			"one",
			fields{
				&tools{
					cfg: &Config{},
				},
				true,
				&http.Server{
					Addr: ":20808",
				},
			},
			new(fatalHook),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log("if no messages after this line, fatal was called")
			v := false
			a := &app{
				tools:      tt.fields.tools,
				serveHttp:  tt.fields.serveHttp,
				httpServer: tt.fields.httpServer,
				grpcServer: grpc.NewServer(),
			}
			a.tools.log = zap.NewNop().WithOptions(zap.WithFatalHook(tt.fatalHook))
			a.httpServer.RegisterOnShutdown(func() {
				v = true
			})
			a.shutdown()
			go func() {
				<-time.After(time.Second)
				a.shutdownCh <- os.Kill
			}()
			a.listenHttp()
			<-time.After(time.Second)
			if tt.expectFatal != tt.fatalHook.called {
				t.Errorf("expected fatalHook to be called")
			}
			t.Log(*tt.fatalHook)
			t.Log(v)
		})
	}
}

type fatalHook struct {
	called bool
}

func (f *fatalHook) OnWrite(entry *zapcore.CheckedEntry, fields []zapcore.Field) {
	f.called = true
}

func Test_app_shutdown(t *testing.T) {
	type fields struct {
		serveHttp  bool
		grpcServer *grpc.Server
		httpServer *http.Server
	}
	tests := []struct {
		name   string
		fields fields
		signal os.Signal
	}{
		{
			"basic",
			fields{
				true,
				grpc.NewServer(),
				&http.Server{},
			},
			os.Interrupt,
		},
		{
			"basic",
			fields{
				true,
				grpc.NewServer(),
				&http.Server{},
			},
			os.Kill,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				done: make(chan struct{}),
				tools: &tools{
					log: zap.NewNop(),
				},
				serveHttp:  tt.fields.serveHttp,
				grpcServer: tt.fields.grpcServer,
				httpServer: tt.fields.httpServer,
			}
			a.shutdown()
			<-time.After(time.Second)
			a.shutdownCh <- os.Interrupt
			<-a.done
		})
	}
}

func Test_configOption_option(t *testing.T) {
	type fields struct {
		cfg *Config
	}
	tests := []struct {
		name   string
		fields fields
		want   *Config
	}{
		{"with database", fields{&Config{}}, &Config{}},
		{"without database", fields{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{tools: &tools{}}
			opt := &configOption{
				tt.fields.cfg,
			}
			opt.option(a)
			if !reflect.DeepEqual(a.tools.cfg, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, a.tools.cfg)
			}
		})
	}
}

func Test_databaseOption_option(t *testing.T) {
	type fields struct {
		db *pgx.Conn
	}
	tests := []struct {
		name   string
		fields fields
		want   *pgx.Conn
	}{
		{"with database", fields{&pgx.Conn{}}, &pgx.Conn{}},
		{"without database", fields{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{tools: &tools{}}
			opt := &databaseOption{
				tt.fields.db,
			}
			opt.option(a)
			if !reflect.DeepEqual(a.tools.db, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, a.tools.db)
			}
		})
	}
}

func Test_grpcServerOption_option(t *testing.T) {
	type fields struct {
		srv *grpc.Server
	}
	s := grpc.NewServer()
	tests := []struct {
		name   string
		fields fields
		want   *grpc.Server
	}{
		{"with grpcServer", fields{s}, s},
		{"without grpcServer", fields{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{tools: &tools{}}
			opt := &grpcServerOption{
				tt.fields.srv,
			}
			opt.option(a)
			if !reflect.DeepEqual(a.grpcServer, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, a.grpcServer)
			}
		})
	}
}

func Test_grpcServerOptionsOption_option(t *testing.T) {
	type fields struct {
		options []grpc.ServerOption
	}
	tests := []struct {
		name   string
		fields fields
		want   []grpc.ServerOption
	}{
		{"none", fields{}, nil},
		{"none", fields{[]grpc.ServerOption{grpc.EmptyServerOption{}}}, []grpc.ServerOption{grpc.EmptyServerOption{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{}
			opt := &grpcServerOptionsOption{
				options: tt.fields.options,
			}
			opt.option(a)
			if !reflect.DeepEqual(a.serverOptions, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, a.serverOptions)
			}
		})
	}
}

func Test_httpOption_option(t *testing.T) {
	a := &app{}
	if a.serveHttp == true {
		t.Error("expected false, got true")
	}
	(&httpOption{}).option(a)
	if a.serveHttp == false {
		t.Error("expected true, got false")
	}
}

func Test_httpServerOption_option(t *testing.T) {
	type fields struct {
		srv *http.Server
	}
	tests := []struct {
		name   string
		fields fields
		want   *http.Server
	}{
		{"with httpServer", fields{&http.Server{}}, &http.Server{}},
		{"without httpServer", fields{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{tools: &tools{}}
			opt := &httpServerOption{
				tt.fields.srv,
			}
			opt.option(a)
			if !reflect.DeepEqual(a.httpServer, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, a.tools.log)
			}
		})
	}
}

func Test_loggerOption_option(t *testing.T) {
	type fields struct {
		log *zap.Logger
	}
	tests := []struct {
		name   string
		fields fields
		want   *zap.Logger
	}{
		{"with logger", fields{zap.NewNop()}, zap.NewNop()},
		{"with logger", fields{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{tools: &tools{}}
			opt := &loggerOption{
				log: tt.fields.log,
			}
			opt.option(a)
			if !reflect.DeepEqual(a.tools.log, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, a.tools.log)
			}
		})
	}
}

func Test_serviceImplementationOption_option(t *testing.T) {
	s := &sampleImplementation{}
	type fields struct {
		desc *grpc.ServiceDesc
		impl Implementation
	}
	d := &grpc.ServiceDesc{}
	tests := []struct {
		name   string
		fields fields
		want   *app
	}{
		{"with implementation", fields{d, s}, &app{serviceImplementations: []serviceImplementation{{d, s}}}},
		{"without implementation", fields{}, &app{serviceImplementations: []serviceImplementation{{}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{}
			opt := &serviceImplementationOption{
				desc: tt.fields.desc,
				impl: tt.fields.impl,
			}
			opt.option(a)
			if !reflect.DeepEqual(a, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, a)
			}
		})
	}
}

func Test_tools_Config(t1 *testing.T) {
	type fields struct {
		cfg *Config
	}
	tests := []struct {
		name   string
		fields fields
		want   *Config
	}{
		{"with config", fields{&Config{}}, &Config{}},
		{"without config", fields{}, nil},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &tools{
				cfg: tt.fields.cfg,
			}
			if got := t.Config(); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Config() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tools_DB(t1 *testing.T) {
	type fields struct {
		db *pgx.Conn
	}
	tests := []struct {
		name   string
		fields fields
		want   *pgx.Conn
	}{
		{"with db", fields{&pgx.Conn{}}, &pgx.Conn{}},
		{"without db", fields{}, nil},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &tools{
				db: tt.fields.db,
			}
			if got := t.DB(); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("DB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tools_Log(t1 *testing.T) {
	type fields struct {
		log *zap.Logger
	}
	tests := []struct {
		name   string
		fields fields
		want   *zap.Logger
	}{
		{"has logger", fields{zap.NewNop()}, zap.NewNop()},
		{"without logger", fields{}, nil},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &tools{
				log: tt.fields.log,
			}
			if got := t.Logger(); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Logger() = %v, want %v", got, tt.want)
			}
		})
	}
}

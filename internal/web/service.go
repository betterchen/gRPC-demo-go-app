package web

import (
	"context"
	"net/http"

	web_v1 "github.com/betterchen/go-project-tmpl/api/proto/demo/v1"
	"github.com/betterchen/go-project-tmpl/pkg/multiservices"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

// RunServices ...
func RunServices(ctx context.Context) error {

	// 创建gRPC服务实例
	grpcTracer := NewGRPCUnaryServerInterceptor(generateID)
	s := grpc.NewServer(grpc.UnaryInterceptor(grpcTracer))

	// 注册服务Handler
	webSrv := DemoServer{}
	web_v1.RegisterDemoServer(s, webSrv)

	// gRPC反射API
	reflection.Register(s)

	// 创建gateway路由
	gw := runtime.NewServeMux(
		// 完全自定义income metadata
		runtime.WithMetadata(
			func(ctx context.Context, r *http.Request) metadata.MD {
				kv := map[string]string{}
				for k := range r.Header {
					kv[k] = r.Header.Get(k)
				}

				var md metadata.MD
				if c, ok := metadata.FromIncomingContext(ctx); ok {
					md = metadata.Join(c, metadata.New(kv))
				} else {
					md = metadata.New(kv)
				}

				return md
			},
		),
		// 也可以通过修改HeaderMatcher来指定需要映射的Header
		runtime.WithIncomingHeaderMatcher(
			runtime.DefaultHeaderMatcher,
		),
	)

	// 注册gRPC服务Handler
	if err := web_v1.RegisterDemoHandlerFromEndpoint(
		ctx,
		gw,
		"localhost:50051",
		[]grpc.DialOption{grpc.WithInsecure()},
	); err != nil {
		return err
	}

	// 创建额外的http路由
	r := mux.NewRouter()
	r.NotFoundHandler = gw

	// 注册路由配置
	registerRoutes(r)

	// 启动服务
	if err := multiservices.AddMod("grpc-server", &multiservices.GRPCMod{
		Port:   "50051",
		Server: s,
	}); err != nil {
		return err
	}

	if err := multiservices.AddMod("http-server", &multiservices.HTTPMod{
		Port: "8080",
		Server: &http.Server{
			Handler: applyMiddlewares(r),
		},
	}); err != nil {
		return err
	}

	return nil
}

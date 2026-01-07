package main

import (
	"context"
	"log"
	"net"

	"github.com/casnerano/protoc-gen-go-guard/example/internal/demo"
	desc "github.com/casnerano/protoc-gen-go-guard/example/pb/demo"
	"github.com/casnerano/protoc-gen-go-guard/pkg/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func createSubjectResolver() interceptor.SubjectResolver {
	return func(ctx context.Context) (*interceptor.Subject, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, nil
		}

		subject := interceptor.Subject{}
		if roles, exists := md["roles"]; exists {
			subject.Roles = roles
		}

		return &subject, nil
	}
}

func buildPolicies() interceptor.Policies {
	return interceptor.Policies{
		"demo-period": func(ctx context.Context, input *interceptor.Input) (bool, error) {
			// check demo period with input
			return true, nil
		},
		"premium": func(ctx context.Context, input *interceptor.Input) (bool, error) {
			// check premium with input
			return false, nil
		},
	}
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:9091")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	guardInterceptor := interceptor.New(
		createSubjectResolver(),
		interceptor.WithPolicies(buildPolicies()),
		interceptor.WithDebug(),
	)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			guardInterceptor.Unary(),
		),
	)

	desc.RegisterAuthServer(server, &demo.AuthServer{})
	desc.RegisterUserServer(server, &demo.UserServer{})

	reflection.Register(server)

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

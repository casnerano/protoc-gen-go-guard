package main

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/casnerano/protoc-gen-go-guard/example/internal/demo"
	desc "github.com/casnerano/protoc-gen-go-guard/example/pb/demo"
	"github.com/casnerano/protoc-gen-go-guard/pkg/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func createAuthResolver() interceptor.AuthContextResolver {
	return func(ctx context.Context) (*interceptor.AuthContext, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return &interceptor.AuthContext{}, nil
		}

		authContext := interceptor.AuthContext{}

		if authorization, exists := md["authorization"]; exists && len(authorization) > 0 {
			authContext.Authenticated = true
		}

		if roles, exists := md["roles"]; exists && len(roles) > 0 {
			authContext.Roles = strings.Split(roles[0], ",")
		}

		return &authContext, nil
	}
}

func buildPolicies() interceptor.Policies {
	return interceptor.Policies{
		"demo-period": func(ctx context.Context, authContext *interceptor.AuthContext, request interface{}) (bool, error) {
			// check demo period with auth context and request
			return true, nil
		},
		"premium": func(ctx context.Context, authContext *interceptor.AuthContext, request interface{}) (bool, error) {
			// check premium with auth context and request
			return false, nil
		},
	}
}

func main() {
	listener, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptor.GuardUnary(
				createAuthResolver(),
				interceptor.WithPolicies(buildPolicies()),
				interceptor.WithDebug(),
			),
		),
	)

	desc.RegisterAuthServer(server, &demo.AuthServer{})
	desc.RegisterUserServer(server, &demo.UserServer{})

	reflection.Register(server)

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

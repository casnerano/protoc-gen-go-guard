package main

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/casnerano/protoc-gen-go-guard/example/internal/demo"
	desc "github.com/casnerano/protoc-gen-go-guard/example/pb/demo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func createAuthResolver() test.AuthContextResolver {
    return func(ctx context.Context) (*test.AuthContext, error) {
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return &test.AuthContext{}, nil
        }

        authContext := test.AuthContext{}

        if authorization, exists := md["authorization"]; exists && len(authorization) > 0 {
            authContext.Authenticated = true
        }

        if roles, exists := md["roles"]; exists && len(roles) > 0 {
            authContext.Roles = strings.Split(roles[0], ",")
        }

        return &authContext, nil
    }
}

func buildPolicies() test.Policies {
    return test.Policies{
        "demo-period": func(ctx context.Context, authContext *test.AuthContext, request interface{}) (bool, error) {
            // check demo period with auth context and request
            return true, nil
        },
        "premium": func(ctx context.Context, authContext *test.AuthContext, request interface{}) (bool, error) {
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
            test.GuardUnary(
                createAuthResolver(),
                test.WithPolicies(buildPolicies()),
                test.WithDebug(),
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

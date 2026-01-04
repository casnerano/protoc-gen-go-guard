package interceptor

import (
    "context"
    "github.com/casnerano/protoc-gen-go-guard/pkg/guard"
    "google.golang.org/grpc"
)

type (
    Subject struct {
        Roles []string
        Attrs map[string]any
    }

    SubjectResolver func(ctx context.Context) (*Subject, error)
)

type Input struct {
    Request any
    Subject Subject
}

type (
    Policy   func(ctx context.Context, input *Input) (bool, error)
    Policies map[string]Policy
)

type (
    OnErrorHandler        func(ctx context.Context, input *Input, err error)
    OnAccessDeniedHandler func(ctx context.Context, input *Input)

    EventHandlers struct {
        OnError        OnErrorHandler
        OnAccessDenied OnAccessDeniedHandler
    }
)

type interceptor struct {
    debug           bool
    policies        Policies
    defaultRules    guard.Rules
    eventHandlers   EventHandlers
    subjectResolver SubjectResolver
}

func New(resolver SubjectResolver, opts ...Option) *interceptor {
    i := interceptor{
        subjectResolver: resolver,
    }

    for _, opt := range opts {
        opt(&i)
    }

    return &i
}

func (i *interceptor) Unary() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
        return handler(ctx, req)
    }
}

func (i *interceptor) Stream() grpc.StreamServerInterceptor {
    return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        return handler(srv, ss)
    }
}

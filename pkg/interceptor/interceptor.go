package interceptor

import (
	"context"
	"log"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	Subject *Subject
}

func (i *Input) Authenticated() bool {
	return i.Subject != nil
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

func (i *interceptor) authorize(ctx context.Context, server any, fullMethod string, req any) error {
	input := Input{
		Request: req,
	}

	subject, err := i.subjectResolver(ctx)
	if err != nil {
		if i.debug {
			log.Printf("Failed to resolve subject for %s: %v", fullMethod, err)
		}

		if i.eventHandlers.OnError != nil {
			i.eventHandlers.OnError(ctx, &input, err)
		}

		return status.Error(codes.Internal, "failed to resolve subject")
	}

	input.Subject = subject

	rules := i.getRules(server, fullMethod)

	allowed, err := i.evaluateRules(ctx, rules, &input)
	if err != nil {
		if i.debug {
			log.Printf("Evaluation error for %s: %v", fullMethod, err)
		}

		if i.eventHandlers.OnError != nil {
			i.eventHandlers.OnError(ctx, &input, err)
		}

		return status.Error(codes.Internal, "evaluation error")
	}

	if !allowed {
		if i.debug {
			log.Printf("Access denied for %s", fullMethod)
		}

		if i.eventHandlers.OnAccessDenied != nil {
			i.eventHandlers.OnAccessDenied(ctx, &input)
		}

		return status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	if i.debug {
		log.Printf("Access granted for %s", fullMethod)
	}

	return nil
}

func (i *interceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if err = i.authorize(ctx, info.Server, info.FullMethod, req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (i *interceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := i.authorize(ss.Context(), srv, info.FullMethod, nil); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

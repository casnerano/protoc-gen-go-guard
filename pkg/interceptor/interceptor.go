// Package interceptor provides a gRPC server interceptor that enforces
// access control rules defined in .proto files via protoc-gen-go-guard.
//
// Rules are evaluated at runtime based on the current request context,
// an injected subject resolver, and optional custom policy functions.
// The default behavior follows a zero-trust model: if no rule explicitly allows access,
// the request is denied.
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
	// Subject represents the authenticated principal making the request.
	// It carries identity attributes such as roles and arbitrary custom data.
	Subject struct {
		Roles []string
		Attrs map[string]any
	}

	// SubjectResolver is a function that extracts a Subject from the request context.
	// If the user is unauthenticated, it should return (nil, nil).
	// Any error returned will cause the interceptor to reject the request with an internal error.
	SubjectResolver func(ctx context.Context) (*Subject, error)
)

// Input encapsulates the data available during rule evaluation.
type Input struct {
	Request any      // The original gRPC request message (nil for streaming calls).
	Subject *Subject // The resolved subject (nil if unauthenticated).
}

// Authenticated returns true if the request is associated with an authenticated subject.
func (i *Input) Authenticated() bool {
	return i.Subject != nil
}

type (
	// Policy is a function that evaluates a custom authorization condition.
	// It receives the current context and input, and returns whether the policy allows access.
	// Any error returned will cause the interceptor to reject the request with an internal error.
	Policy func(ctx context.Context, input *Input) (bool, error)
	// Policies is a registry of named policy functions referenced in .proto guard rules.
	Policies map[string]Policy
)

type (
	// OnErrorHandler is called when an error occurs during subject resolution or rule evaluation.
	OnErrorHandler func(ctx context.Context, input *Input, err error)
	// OnAccessDeniedHandler is called when access is denied by guard rules.
	OnAccessDeniedHandler func(ctx context.Context, input *Input)

	EventHandlers struct {
		OnError        OnErrorHandler
		OnAccessDenied OnAccessDeniedHandler
	}
)

type Interceptor struct {
	debug           bool
	policies        Policies
	defaultRules    guard.Rules
	eventHandlers   EventHandlers
	subjectResolver SubjectResolver
}

// New creates a new guard interceptor.
// It requires a SubjectResolver and accepts optional configuration via Options.
func New(resolver SubjectResolver, opts ...Option) *Interceptor {
	i := Interceptor{
		subjectResolver: resolver,
	}

	for _, opt := range opts {
		opt(&i)
	}

	return &i
}

// authorize evaluates whether the current request is allowed based on the resolved subject
// and the applicable access rules. Returns nil on success, or a gRPC error on denial/failure.
func (i *Interceptor) authorize(ctx context.Context, server any, fullMethod string, req any) error {
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

// Unary returns a grpc.UnaryServerInterceptor that enforces guard rules
// on unary (request-response) gRPC methods.
func (i *Interceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if err = i.authorize(ctx, info.Server, info.FullMethod, req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// Stream returns a grpc.StreamServerInterceptor that enforces guard rules
// on streaming gRPC methods.
func (i *Interceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := i.authorize(ss.Context(), srv, info.FullMethod, nil); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

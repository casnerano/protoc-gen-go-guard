package interceptor

import (
	"context"
	"fmt"
	"path"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type guardServiceProvider interface {
	GetGuardService() *guard.Service
}

type evaluator interface {
	Evaluate(ctx context.Context, rules *guard.Rules, authContext *AuthContext, request any) (bool, error)
}

type AuthContext struct {
	Authenticated bool
	Roles         []string
	Metadata      map[string]any
}

type AuthContextResolver func(ctx context.Context) (*AuthContext, error)

func GuardUnary(authContextResolver AuthContextResolver, opts ...Option) grpc.UnaryServerInterceptor {
	guardOptions := &options{}

	for _, opt := range opts {
		opt(guardOptions)
	}

	evaluators := map[evaluatorType]evaluator{
		evaluatorTypeAllowPublic:  newAllowedPublicEvaluator(),
		evaluatorTypeRequireAuthn: newRequireAuthnEvaluator(),
		evaluatorTypeRoleBased:    newRoleBasedEvaluator(),
	}

	if guardOptions.policies != nil {
		evaluators[evaluatorTypePolicyBased] = newPolicyBasedEvaluator(guardOptions.policies)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if provider, ok := info.Server.(guardServiceProvider); ok {
			rules := findRulesForService(provider.GetGuardService(), info.FullMethod)

			if rules == nil {
				return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
			}

			selectedEvaluatorType := getEvaluatorType(rules)

			if selectedEvaluatorType == evaluatorTypeUnknown {
				return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
			}

			authContext, err := authContextResolver(ctx)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("failed resolve auth context: %s", err))
			}

			selectedEvaluator, exists := evaluators[selectedEvaluatorType]
			if !exists {
				return nil, status.Error(codes.Internal, "failed evaluators configuration")
			}

			allowed, err := selectedEvaluator.Evaluate(ctx, rules, authContext, req)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("failed evaluate access for method %q: %s", info.FullMethod, err))
			}

			if !allowed {
				return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
			}

			return handler(ctx, req)
		}

		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}
}

func findRulesForService(service *guard.Service, fullMethod string) *guard.Rules {
	if service == nil {
		return nil
	}

	if method, exists := service.Methods[path.Base(fullMethod)]; exists && method.Rules != nil {
		return method.Rules
	}

	return service.Rules
}

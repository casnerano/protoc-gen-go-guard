package interceptor

import (
	"context"
	"fmt"
	"path"

	"github.com/casnerano/protoc-gen-go-rbac/pkg/rbac"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rbacServiceProvider interface {
	GetRBACService() *rbac.Service
}

type evaluator interface {
	Evaluate(ctx context.Context, rules *rbac.Rules, authContext *AuthContext, request any) (bool, error)
}

type AuthContext struct {
	Authenticated bool
	Roles         []string
	Metadata      map[string]any
}

type AuthContextResolver func(ctx context.Context) (*AuthContext, error)

func RbacUnary(authContextResolver AuthContextResolver, opts ...Option) grpc.UnaryServerInterceptor {
	rbacOptions := &options{}

	for _, opt := range opts {
		opt(rbacOptions)
	}

	evaluators := map[evaluatorType]evaluator{
		evaluatorTypeAllowPublic:  newAllowedPublicEvaluator(),
		evaluatorTypeRequireAuthn: newRequireAuthnEvaluator(),
		evaluatorTypeRoleBased:    newRoleBasedEvaluator(),
	}

	if rbacOptions.policies != nil {
		evaluators[evaluatorTypePolicyBased] = newPolicyBasedEvaluator(rbacOptions.policies)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if provider, ok := info.Server.(rbacServiceProvider); ok {
			rules := findRulesForService(provider.GetRBACService(), info.FullMethod)

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

func findRulesForService(service *rbac.Service, fullMethod string) *rbac.Rules {
	if service == nil {
		return nil
	}

	if method, exists := service.Methods[path.Base(fullMethod)]; exists && method.Rules != nil {
		return method.Rules
	}

	return service.Rules
}

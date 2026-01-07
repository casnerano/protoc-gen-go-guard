package interceptor

import (
	"path"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
)

type guardServiceProvider interface {
	GuardService() *guard.Service
}

// getGuardService retrieves the guard.Service metadata associated with a gRPC server.
func (i *interceptor) getGuardService(server any) *guard.Service {
	provider, ok := server.(guardServiceProvider)
	if !ok {
		return nil
	}

	return provider.GuardService()
}

// getRules returns the effective access rules for a specific gRPC method.
// It applies the precedence order: method rules → service rules → default rules.
func (i *interceptor) getRules(server any, fullMethod string) guard.Rules {
	service := i.getGuardService(server)
	if service == nil {
		return nil
	}

	if method, exists := service.Methods[path.Base(fullMethod)]; exists && method.Rules != nil {
		return method.Rules
	}

	if service.Rules != nil {
		return service.Rules
	}

	return i.defaultRules
}

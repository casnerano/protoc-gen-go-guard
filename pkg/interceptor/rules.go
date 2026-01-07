package interceptor

import (
	"path"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
)

type guardServiceProvider interface {
	GuardService() *guard.Service
}

func (i *interceptor) getGuardService(server any) *guard.Service {
	provider, ok := server.(guardServiceProvider)
	if !ok {
		return nil
	}

	return provider.GuardService()
}

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

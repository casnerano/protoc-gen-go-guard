package interceptor

import (
    "github.com/casnerano/protoc-gen-go-guard/pkg/guard"
    "path"
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

func (i *interceptor) getServiceRules(server any) guard.Rules {
    service := i.getGuardService(server)
    if service == nil {
        return nil
    }

    return service.Rules
}

func (i *interceptor) getMethodRules(server any, fullMethod string) guard.Rules {
    service := i.getGuardService(server)
    if service == nil {
        return nil
    }

    method, exists := service.Methods[path.Base(fullMethod)]
    if !exists {
        return nil
    }

    return method.Rules
}

func (i *interceptor) getRules(server any, fullMethod string) guard.Rules {
    methodRules := i.getMethodRules(server, fullMethod)
    if methodRules != nil {
        return methodRules
    }

    serviceRules := i.getServiceRules(server)
    if serviceRules != nil {
        return serviceRules
    }

    return i.defaultRules
}

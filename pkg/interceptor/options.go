package interceptor

import "github.com/casnerano/protoc-gen-go-guard/pkg/guard"

type Option func(i *interceptor)

func WithDebug() Option {
    return func(i *interceptor) {
        i.debug = true
    }
}

func WithPolicies(policies Policies) Option {
    return func(i *interceptor) {
        i.policies = policies
    }
}

func WithDefaultRules(rules guard.Rules) Option {
    return func(i *interceptor) {
        i.defaultRules = rules
    }
}

func WithOnError(handler OnErrorHandler) Option {
    return func(i *interceptor) {
        if handler != nil {
            i.eventHandlers.OnError = handler
        }
    }
}

func WithOnAccessDenied(handler OnAccessDeniedHandler) Option {
    return func(i *interceptor) {
        if handler != nil {
            i.eventHandlers.OnAccessDenied = handler
        }
    }
}

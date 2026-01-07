package interceptor

import "github.com/casnerano/protoc-gen-go-guard/pkg/guard"

// Option configures the behavior of the interceptor.
type Option func(i *Interceptor)

// WithDebug enables debug logging (e.g., "access granted/denied" messages).
func WithDebug() Option {
	return func(i *Interceptor) {
		i.debug = true
	}
}

// WithPolicies registers named policy functions that can be referenced
// in AuthenticatedAccess.PolicyBased rules in .proto files.
func WithPolicies(policies Policies) Option {
	return func(i *Interceptor) {
		i.policies = policies
	}
}

// WithDefaultRules sets global fallback rules applied when a service
// has no service-level or method-level rules defined.
// By default, the interceptor uses a zero-trust model (deny all).
func WithDefaultRules(rules guard.Rules) Option {
	return func(i *Interceptor) {
		i.defaultRules = rules
	}
}

// WithOnError registers a handler invoked when an internal error occurs
// during subject resolution or rule evaluation.
func WithOnError(handler OnErrorHandler) Option {
	return func(i *Interceptor) {
		if handler != nil {
			i.eventHandlers.OnError = handler
		}
	}
}

// WithOnAccessDenied registers a handler invoked when a request is denied
// due to guard rules.
func WithOnAccessDenied(handler OnAccessDeniedHandler) Option {
	return func(i *Interceptor) {
		if handler != nil {
			i.eventHandlers.OnAccessDenied = handler
		}
	}
}

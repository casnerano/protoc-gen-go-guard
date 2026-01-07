package interceptor

import "github.com/casnerano/protoc-gen-go-guard/pkg/guard"

// Option configures the behavior of the interceptor.
type Option func(i *interceptor)

// WithDebug enables debug logging (e.g., "access granted/denied" messages).
func WithDebug() Option {
	return func(i *interceptor) {
		i.debug = true
	}
}

// WithPolicies registers named policy functions that can be referenced
// in AuthenticatedAccess.PolicyBased rules in .proto files.
func WithPolicies(policies Policies) Option {
	return func(i *interceptor) {
		i.policies = policies
	}
}

// WithDefaultRules sets global fallback rules applied when a service
// has no service-level or method-level rules defined.
// By default, the interceptor uses a zero-trust model (deny all).
func WithDefaultRules(rules guard.Rules) Option {
	return func(i *interceptor) {
		i.defaultRules = rules
	}
}

// WithOnError registers a handler invoked when an internal error occurs
// during subject resolution or rule evaluation.
func WithOnError(handler OnErrorHandler) Option {
	return func(i *interceptor) {
		if handler != nil {
			i.eventHandlers.OnError = handler
		}
	}
}

// WithOnAccessDenied registers a handler invoked when a request is denied
// due to guard rules.
func WithOnAccessDenied(handler OnAccessDeniedHandler) Option {
	return func(i *interceptor) {
		if handler != nil {
			i.eventHandlers.OnAccessDenied = handler
		}
	}
}

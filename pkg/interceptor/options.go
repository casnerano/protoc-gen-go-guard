package interceptor

type options struct {
	debug    bool
	policies Policies
}

type Option func(*options)

func WithDebug() Option {
	return func(o *options) {
		o.debug = true
	}
}

func WithPolicies(policies Policies) Option {
	return func(o *options) { o.policies = policies }
}

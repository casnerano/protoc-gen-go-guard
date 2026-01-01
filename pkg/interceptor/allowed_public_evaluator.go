package interceptor

import (
	"context"

	"github.com/casnerano/protoc-gen-go-rbac/pkg/rbac"
)

type allowedPublicEvaluator struct{}

func newAllowedPublicEvaluator() *allowedPublicEvaluator {
	return &allowedPublicEvaluator{}
}

func (e allowedPublicEvaluator) Evaluate(_ context.Context, rules *rbac.Rules, _ *AuthContext, _ any) (bool, error) {
	if rules == nil {
		return false, nil
	}

	return rules.AllowPublic != nil && *rules.AllowPublic, nil
}

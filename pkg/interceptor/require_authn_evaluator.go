package interceptor

import (
	"context"

	"github.com/casnerano/protoc-gen-go-rbac/pkg/rbac"
)

type requireAuthnEvaluator struct{}

func newRequireAuthnEvaluator() *requireAuthnEvaluator {
	return &requireAuthnEvaluator{}
}

func (e requireAuthnEvaluator) Evaluate(_ context.Context, _ *rbac.Rules, authContext *AuthContext, _ any) (bool, error) {
	return authContext.Authenticated, nil
}

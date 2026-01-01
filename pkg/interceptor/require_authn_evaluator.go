package interceptor

import (
	"context"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
)

type requireAuthnEvaluator struct{}

func newRequireAuthnEvaluator() *requireAuthnEvaluator {
	return &requireAuthnEvaluator{}
}

func (e requireAuthnEvaluator) Evaluate(_ context.Context, _ *guard.Rules, authContext *AuthContext, _ any) (bool, error) {
	return authContext.Authenticated, nil
}

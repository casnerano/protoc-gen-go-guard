package interceptor

import (
	"context"
	"errors"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
)

var (
	errUnauthenticated = errors.New("unauthenticated")
)

func (i *interceptor) rulesEvaluate(ctx context.Context, rules guard.Rules, input *Input) (bool, error) {
	for _, rule := range rules {
		allowed, err := i.ruleEvaluate(ctx, rule, input)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

func (i *interceptor) ruleEvaluate(ctx context.Context, rule *guard.Rule, input *Input) (bool, error) {
	return false, nil
}

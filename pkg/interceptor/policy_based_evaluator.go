package interceptor

import (
	"context"
	"fmt"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
)

type (
	Policy   func(ctx context.Context, authContext *AuthContext, request any) (bool, error)
	Policies map[string]Policy
)

type policyBasedEvaluator struct {
	policies Policies
}

func newPolicyBasedEvaluator(policies Policies) *policyBasedEvaluator {
	return &policyBasedEvaluator{
		policies: policies,
	}
}

func (e policyBasedEvaluator) Evaluate(ctx context.Context, rules *guard.Rules, authContext *AuthContext, request any) (bool, error) {
	if rules.PolicyBased == nil {
		return false, nil
	}

	for _, policy := range rules.PolicyBased.PolicyNames {
		if _, ok := e.policies[policy]; !ok {
			return false, fmt.Errorf("undefined policy: %s", policy)
		}
	}

	switch rules.PolicyBased.Requirement {
	case guard.RequirementAny:
		for _, policy := range rules.PolicyBased.PolicyNames {
			result, err := e.policies[policy](ctx, authContext, request)
			if err != nil {
				return false, err
			}

			if result {
				return true, nil
			}
		}
		return false, nil
	case guard.RequirementAll:
		for _, policy := range rules.PolicyBased.PolicyNames {
			result, err := e.policies[policy](ctx, authContext, request)
			if err != nil {
				return false, err
			}

			if !result {
				return false, nil
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("unknown requirement type: %v", rules.PolicyBased.Requirement)
	}
}

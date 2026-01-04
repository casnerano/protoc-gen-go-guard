package interceptor

import (
	"context"
	"fmt"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
)

func (i *interceptor) evaluateRules(ctx context.Context, rules guard.Rules, input *Input) (bool, error) {
	for _, rule := range rules {
		allowed, err := i.evaluateRule(ctx, rule, input)
		if err != nil {
			return false, err
		}

		if allowed {
			return true, nil
		}
	}

	return false, nil
}

func (i *interceptor) evaluateRule(ctx context.Context, rule *guard.Rule, input *Input) (bool, error) {
	if rule.AllowPublic != nil && *rule.AllowPublic {
		return true, nil
	}

	if rule.RequireAuthentication != nil && *rule.RequireAuthentication {
		if !input.Authenticated() {
			return false, nil
		}

		return true, nil
	}

	if rule.AuthenticatedAccess != nil {
		if !input.Authenticated() {
			return false, nil
		}

		if rule.AuthenticatedAccess.RoleBased != nil {
			allowed, err := i.evaluateRoleBasedAccess(ctx, rule.AuthenticatedAccess.RoleBased, input)
			if err != nil {
				return false, err
			}

			if !allowed {
				return false, nil
			}
		}

		if rule.AuthenticatedAccess.PolicyBased != nil {
			allowed, err := i.evaluatePolicyBasedAccess(ctx, rule.AuthenticatedAccess.PolicyBased, input)
			if err != nil {
				return false, err
			}

			if !allowed {
				return false, nil
			}
		}

		return true, nil
	}

	return false, nil
}

func (i *interceptor) evaluateRoleBasedAccess(_ context.Context, roleBased *guard.RoleBased, input *Input) (bool, error) {
	var matchedRoles int
	for _, requiredRole := range roleBased.Roles {
		for _, subjectRole := range input.Subject.Roles {
			if requiredRole == subjectRole {
				matchedRoles++
				break
			}
		}
	}

	switch roleBased.Match {
	case guard.MatchAll:
		return matchedRoles == len(roleBased.Roles), nil
	case guard.MatchAtLeastOne:
		return matchedRoles > 0, nil
	default:
		return false, fmt.Errorf("unknown roles match type")
	}
}

func (i *interceptor) evaluatePolicyBasedAccess(ctx context.Context, policyBased *guard.PolicyBased, input *Input) (bool, error) {
	var matchedPolicies int
	for _, policyName := range policyBased.Policies {
		policy, exists := i.policies[policyName]
		if !exists {
			continue
		}

		allowed, err := policy(ctx, input)
		if err != nil {
			return false, err
		}

		if allowed {
			matchedPolicies++
		}
	}

	switch policyBased.Match {
	case guard.MatchAll:
		return matchedPolicies == len(policyBased.Policies), nil
	case guard.MatchAtLeastOne:
		return matchedPolicies > 0, nil
	default:
		return false, fmt.Errorf("unknown policies match type")
	}
}

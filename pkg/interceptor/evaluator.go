package interceptor

import (
	"context"
	"errors"
	"fmt"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
)

var (
	ErrUndefinedPolicy = errors.New("undefined policy")
	ErrInvalidPolicy   = errors.New("invalid policy")
)

// evaluateRules checks a list of rules in order. Access is granted if any rule allows it.
// Returns true if access is granted, false if all rules deny access.
func (i *Interceptor) evaluateRules(ctx context.Context, rules guard.Rules, input *Input) (bool, error) {
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

// evaluateRule checks a single rule.
// Returns true if the rule allows access.
func (i *Interceptor) evaluateRule(ctx context.Context, rule *guard.Rule, input *Input) (bool, error) {
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

		var (
			err error

			allowedRoleBased   bool
			allowedPolicyBased bool
		)

		if rule.AuthenticatedAccess.RoleBased != nil {
			allowedRoleBased, err = i.evaluateRoleBasedAccess(ctx, rule.AuthenticatedAccess.RoleBased, input)
			if err != nil {
				return false, err
			}

			if !allowedRoleBased {
				return false, nil
			}
		}

		if rule.AuthenticatedAccess.PolicyBased != nil {
			allowedPolicyBased, err = i.evaluatePolicyBasedAccess(ctx, rule.AuthenticatedAccess.PolicyBased, input)
			if err != nil {
				return false, err
			}

			if !allowedPolicyBased {
				return false, nil
			}
		}

		return allowedRoleBased || allowedPolicyBased, nil
	}

	return false, nil
}

// evaluateRoleBasedAccess checks if the subject satisfies the role-based conditions.
// Returns true if access should be granted based on roles.
func (i *Interceptor) evaluateRoleBasedAccess(_ context.Context, roleBased *guard.RoleBased, input *Input) (bool, error) {
	if len(roleBased.Roles) == 0 {
		return false, nil
	}

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

// evaluateRule checks if a single rule allows access.
//   - AllowPublic: grants access unconditionally.
//   - RequireAuthentication: grants access if subject is authenticated.
//   - AuthenticatedAccess: requires authentication and optionally checks roles and/or policies.
//     If both roles and policies are specified, both must allow access.
func (i *Interceptor) evaluatePolicyBasedAccess(ctx context.Context, policyBased *guard.PolicyBased, input *Input) (bool, error) {
	if len(policyBased.Policies) == 0 {
		return false, nil
	}

	var matchedPolicies int
	for _, policyName := range policyBased.Policies {
		policy, exists := i.policies[policyName]
		if !exists {
			return false, fmt.Errorf("policy %q not defined: %w", policyName, ErrUndefinedPolicy)
		}

		if policy == nil {
			return false, fmt.Errorf("policy %q is nil: %w", policyName, ErrInvalidPolicy)
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

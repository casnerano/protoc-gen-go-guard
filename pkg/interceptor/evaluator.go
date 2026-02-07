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
func (i *Interceptor) evaluateRules(ctx context.Context, rules guard.Rules, input *Input) (*EvaluationResult, error) {
	var firstDeniedResult *EvaluationResult

	for _, rule := range rules {
		result, err := i.evaluateRule(ctx, rule, input)
		if err != nil {
			return nil, err
		}

		if result.Allowed {
			return result, nil
		}

		if firstDeniedResult == nil {
			firstDeniedResult = result
		}
	}

	if firstDeniedResult != nil {
		return firstDeniedResult, nil
	}

	return &EvaluationResult{Allowed: false, Rule: RuleKindPrivate}, nil
}

// evaluateRule checks a single rule.
func (i *Interceptor) evaluateRule(ctx context.Context, rule *guard.Rule, input *Input) (*EvaluationResult, error) {
	if rule.AllowPublic != nil && *rule.AllowPublic {
		return &EvaluationResult{Allowed: true, Rule: RuleKindPublic}, nil
	}

	if rule.RequireAuthentication != nil && *rule.RequireAuthentication {
		if !input.Authenticated() {
			return &EvaluationResult{Allowed: false, Rule: RuleKindAuthenticated}, nil
		}

		return &EvaluationResult{Allowed: true, Rule: RuleKindAuthenticated}, nil
	}

	if rule.AuthenticatedAccess != nil {
		if !input.Authenticated() {
			return &EvaluationResult{Allowed: false, Rule: RuleKindAuthenticated}, nil
		}

		var (
			err error

			allowedRoleBased   bool
			allowedPolicyBased bool
		)

		if rule.AuthenticatedAccess.RoleBased != nil {
			allowedRoleBased, err = i.evaluateRoleBasedAccess(ctx, rule.AuthenticatedAccess.RoleBased, input)
			if err != nil {
				return nil, err
			}

			if !allowedRoleBased {
				return &EvaluationResult{Allowed: false, Rule: RuleKindRoleBased}, nil
			}
		}

		if rule.AuthenticatedAccess.PolicyBased != nil {
			allowedPolicyBased, err = i.evaluatePolicyBasedAccess(ctx, rule.AuthenticatedAccess.PolicyBased, input)
			if err != nil {
				return nil, err
			}

			if !allowedPolicyBased {
				return &EvaluationResult{Allowed: false, Rule: RuleKindPolicyBased}, nil
			}
		}

		allowed := allowedRoleBased || allowedPolicyBased
		ruleKind := RuleKindRoleBased
		if rule.AuthenticatedAccess.PolicyBased != nil {
			ruleKind = RuleKindPolicyBased
		}
		if !allowed {
			return &EvaluationResult{Allowed: false, Rule: RuleKindPrivate}, nil
		}
		return &EvaluationResult{Allowed: true, Rule: ruleKind}, nil
	}

	return &EvaluationResult{Allowed: false, Rule: RuleKindPrivate}, nil
}

// evaluateRoleBasedAccess checks if the subject satisfies the role-based conditions.
func (i *Interceptor) evaluateRoleBasedAccess(_ context.Context, roleBased *guard.RoleBased, input *Input) (bool, error) {
	if len(roleBased.Roles) == 0 {
		return false, nil
	}

	var matchedRolesCount int
	for _, requiredRole := range roleBased.Roles {
		for _, subjectRole := range input.Subject.Roles {
			if requiredRole == subjectRole {
				matchedRolesCount++
				break
			}
		}
	}

	switch roleBased.Requirement {
	case guard.RequirementAll:
		return matchedRolesCount == len(roleBased.Roles), nil
	case guard.RequirementAtLeastOne:
		return matchedRolesCount > 0, nil
	default:
		return false, fmt.Errorf("unknown roles requirement type")
	}
}

// evaluatePolicyBasedAccess checks if policies allow access.
func (i *Interceptor) evaluatePolicyBasedAccess(ctx context.Context, policyBased *guard.PolicyBased, input *Input) (bool, error) {
	if len(policyBased.Policies) == 0 {
		return false, nil
	}

	var passedPoliciesCount int
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
			passedPoliciesCount++
		}
	}

	switch policyBased.Requirement {
	case guard.RequirementAll:
		return passedPoliciesCount == len(policyBased.Policies), nil
	case guard.RequirementAtLeastOne:
		return passedPoliciesCount > 0, nil
	default:
		return false, fmt.Errorf("unknown requredPolicies requirement type")
	}
}

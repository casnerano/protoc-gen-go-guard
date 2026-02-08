package interceptor

import (
	"context"
	"errors"
	"testing"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_interceptor_evaluateRules(t *testing.T) {
	tests := []struct {
		name     string
		input    Input
		rules    guard.Rules
		policies Policies

		want       *EvaluationResult
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name: "nil rules without subject",
			rules: nil,
			want: &EvaluationResult{Allowed: false, Rule: RuleKindPrivate},
		},
		{
			name: "empty rules without subject",
			rules: guard.Rules{},
			want: &EvaluationResult{Allowed: false, Rule: RuleKindPrivate},
		},
		{
			name: "one allow public rule without subject",
			rules: guard.Rules{
				{AllowPublic: guard.Ptr(true)},
			},
			want: &EvaluationResult{Allowed: true, Rule: RuleKindPublic},
		},
		{
			name: "one require authentication rule without subject",
			rules: guard.Rules{
				{RequireAuthentication: guard.Ptr(true)},
			},
			want: &EvaluationResult{Allowed: false, Rule: RuleKindAuthenticated},
		},
		{
			name:  "multiple rules when one allow rule with subject",
			input: Input{Subject: &Subject{}},
			rules: guard.Rules{
				{RequireAuthentication: guard.Ptr(true)},
				{AllowPublic: guard.Ptr(false)},
			},
			want: &EvaluationResult{Allowed: true, Rule: RuleKindAuthenticated},
		},
		{
			name:  "multiple rules when not allow rules with subject",
			input: Input{Subject: &Subject{Roles: []string{}}},
			rules: guard.Rules{
				{
					AuthenticatedAccess: &guard.AuthenticatedAccess{
						PolicyBased: &guard.PolicyBased{
							Policies:    []string{"negative-policy"},
							Requirement: guard.RequirementAll,
						},
					},
				},
				{
					AuthenticatedAccess: &guard.AuthenticatedAccess{
						RoleBased: &guard.RoleBased{
							Roles:       []string{"non-existing-role"},
							Requirement: guard.RequirementAll,
						},
					},
				},
			},
			policies: Policies{"negative-policy": func(ctx context.Context, input *Input) (bool, error) {
				return false, nil
			}},
			want: &EvaluationResult{Allowed: false, Rule: RuleKindPolicyBased},
		},
		{
			name:  "one policy rule when unknown requirement type with subject",
			input: Input{Subject: &Subject{}},
			rules: guard.Rules{
				{
					AuthenticatedAccess: &guard.AuthenticatedAccess{
						PolicyBased: &guard.PolicyBased{
							Policies:    []string{"positive-policy"},
							Requirement: guard.Requirement(-1),
						},
					},
				},
			},
			policies: Policies{"positive-policy": func(ctx context.Context, input *Input) (bool, error) {
				return true, nil
			}},
			errAssertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{
				policies: tt.policies,
			}

			result, err := i.evaluateRules(context.Background(), tt.rules, &tt.input)
			if tt.errAssertion != nil {
				tt.errAssertion(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, tt.want)
			assert.Equal(t, tt.want.Allowed, result.Allowed)
			assert.Equal(t, tt.want.Rule, result.Rule)
		})
	}
}

func Test_interceptor_evaluateRule(t *testing.T) {
	tests := []struct {
		name     string
		input    Input
		rule     *guard.Rule
		policies Policies

		want         *EvaluationResult
		errAssertion assert.ErrorAssertionFunc
	}{
		{
			name: "empty rule without subject",
			input: Input{},
			rule:  &guard.Rule{},
			want:  &EvaluationResult{Allowed: false, Rule: RuleKindPrivate},
		},
		{
			name: "allow public rule without subject",
			input: Input{},
			rule:  &guard.Rule{AllowPublic: guard.Ptr(true)},
			want:  &EvaluationResult{Allowed: true, Rule: RuleKindPublic},
		},
		{
			name: "require authentication without subject",
			input: Input{},
			rule:  &guard.Rule{RequireAuthentication: guard.Ptr(true)},
			want:  &EvaluationResult{Allowed: false, Rule: RuleKindAuthenticated},
		},
		{
			name: "require authentication with subject",
			input: Input{Subject: &Subject{}},
			rule:  &guard.Rule{RequireAuthentication: guard.Ptr(true)},
			want:  &EvaluationResult{Allowed: true, Rule: RuleKindAuthenticated},
		},
		{
			name:  "role based access with no requirement",
			input: Input{Subject: &Subject{Roles: []string{"user"}}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles:       []string{"admin"},
						Requirement: guard.RequirementAll,
					},
				},
			},
			want: &EvaluationResult{Allowed: false, Rule: RuleKindRoleBased},
		},
		{
			name:  "role based access with requirement",
			input: Input{Subject: &Subject{Roles: []string{"admin"}}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles:       []string{"admin"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
			want: &EvaluationResult{Allowed: true, Rule: RuleKindRoleBased},
		},
		{
			name:  "authenticated access with empty authenticated-rules and without subject",
			input: Input{},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{},
			},
			want: &EvaluationResult{Allowed: false, Rule: RuleKindAuthenticated},
		},
		{
			name:  "authenticated access with empty authenticated-rules and with subject",
			input: Input{Subject: &Subject{}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{},
			},
			want: &EvaluationResult{Allowed: false, Rule: RuleKindPrivate},
		},
		{
			name:  "policy based access with no requirement",
			input: Input{Subject: &Subject{}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"positive-policy"},
						Requirement: guard.RequirementAll,
					},
				},
			},
			policies: Policies{
				"positive-policy": func(ctx context.Context, input *Input) (bool, error) {
					return false, nil
				},
			},
			want: &EvaluationResult{Allowed: false, Rule: RuleKindPolicyBased},
		},
		{
			name:  "policy based access with requirement",
			input: Input{Subject: &Subject{}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"positive-policy"},
						Requirement: guard.RequirementAll,
					},
				},
			},
			policies: Policies{
				"positive-policy": func(ctx context.Context, input *Input) (bool, error) {
					return true, nil
				},
			},
			want: &EvaluationResult{Allowed: true, Rule: RuleKindPolicyBased},
		},
		{
			name:  "policy based access with unknown requirement type",
			input: Input{Subject: &Subject{}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"positive-policy"},
						Requirement: guard.Requirement(-1),
					},
				},
			},
			policies: Policies{
				"positive-policy": func(ctx context.Context, input *Input) (bool, error) {
					return true, nil
				},
			},
			errAssertion: assert.Error,
		},
		{
			name:  "policy based access with undefined policy",
			input: Input{Subject: &Subject{}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"undefined-policy"},
						Requirement: guard.RequirementAll,
					},
				},
			},
			policies: Policies{},
			errAssertion: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrUndefinedPolicy)
			},
		},
		{
			name:  "policy based access with nil policy",
			input: Input{Subject: &Subject{}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"nil-policy"},
						Requirement: guard.RequirementAll,
					},
				},
			},
			policies: Policies{"nil-policy": nil},
			errAssertion: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrInvalidPolicy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{
				policies: tt.policies,
			}

			result, err := i.evaluateRule(context.Background(), tt.rule, &tt.input)
			if tt.errAssertion != nil {
				tt.errAssertion(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, tt.want)
			assert.Equal(t, tt.want.Allowed, result.Allowed)
			assert.Equal(t, tt.want.Rule, result.Rule)
		})
	}
}

func Test_interceptor_evaluateRoleBasedAccess(t *testing.T) {
	tests := []struct {
		name          string
		requiredRoles []string
		subjectRoles  []string
		requirement   guard.Requirement

		allowAssertion assert.BoolAssertionFunc
		errAssertion   assert.ErrorAssertionFunc
	}{
		{
			name:           "nil required roles",
			requiredRoles:  nil,
			allowAssertion: assert.False,
		},
		{
			name:           "empty required roles",
			requiredRoles:  []string{},
			allowAssertion: assert.False,
		},
		{
			name:           "requirement all required roles",
			requiredRoles:  []string{"admin", "manager"},
			subjectRoles:   []string{"admin", "manager"},
			requirement:    guard.RequirementAll,
			allowAssertion: assert.True,
		},
		{
			name:           "no requirement all required roles",
			requiredRoles:  []string{"admin", "manager"},
			subjectRoles:   []string{"admin"},
			requirement:    guard.RequirementAll,
			allowAssertion: assert.False,
		},
		{
			name:           "requirement at least one required roles",
			requiredRoles:  []string{"admin", "manager"},
			subjectRoles:   []string{"user", "manager"},
			requirement:    guard.RequirementAtLeastOne,
			allowAssertion: assert.True,
		},
		{
			name:           "no requirement at least one required roles",
			requiredRoles:  []string{"admin", "manager"},
			subjectRoles:   []string{},
			requirement:    guard.RequirementAtLeastOne,
			allowAssertion: assert.False,
		},
		{
			name:          "unknown requirement type",
			requiredRoles: []string{"admin"},
			subjectRoles:  []string{"admin"},
			requirement:   guard.Requirement(-1),
			errAssertion:  assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{}

			roleBased := &guard.RoleBased{
				Roles:       tt.requiredRoles,
				Requirement: tt.requirement,
			}

			input := &Input{
				Subject: &Subject{
					Roles: tt.subjectRoles,
				},
			}

			allowed, err := i.evaluateRoleBasedAccess(context.Background(), roleBased, input)
			if tt.errAssertion != nil {
				tt.errAssertion(t, err)
				return
			}

			require.NoError(t, err)
			tt.allowAssertion(t, allowed)
		})
	}
}

func Test_interceptor_evaluatePolicyBasedAccess(t *testing.T) {
	tests := []struct {
		name             string
		requiredPolicies []string
		declaredPolicies Policies
		requirement      guard.Requirement

		allowAssertion assert.BoolAssertionFunc
		errAssertion   assert.ErrorAssertionFunc
	}{
		{
			name:             "no required policies",
			requiredPolicies: []string{},
			allowAssertion:   assert.False,
		},
		{
			name:             "requirement all required policies",
			requiredPolicies: []string{"positive-policy-1", "positive-policy-2"},
			declaredPolicies: Policies{
				"positive-policy-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
				"positive-policy-2": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			requirement:    guard.RequirementAll,
			allowAssertion: assert.True,
		},
		{
			name:             "no requirement all required policies",
			requiredPolicies: []string{"positive-policy-1", "negative-policy-1"},
			declaredPolicies: Policies{
				"positive-policy-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
				"negative-policy-1": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
			},
			requirement:    guard.RequirementAll,
			allowAssertion: assert.False,
		},
		{
			name:             "requirement at least one required policies",
			requiredPolicies: []string{"negative-policy-1", "positive-policy-1"},
			declaredPolicies: Policies{
				"negative-policy-1": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
				"positive-policy-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			requirement:    guard.RequirementAtLeastOne,
			allowAssertion: assert.True,
		},
		{
			name:             "no requirement at least one required policies",
			requiredPolicies: []string{"negative-policy-1", "negative-policy-2"},
			declaredPolicies: Policies{
				"negative-policy-1": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
				"negative-policy-2": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
			},
			requirement:    guard.RequirementAtLeastOne,
			allowAssertion: assert.False,
		},
		{
			name:             "unknown requirement type",
			requiredPolicies: []string{"positive-1", "positive-2"},
			declaredPolicies: Policies{
				"positive-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
				"positive-2": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			requirement:  guard.Requirement(-1),
			errAssertion: assert.Error,
		},
		{
			name:             "undefined policy",
			requiredPolicies: []string{"undefined-policy"},
			declaredPolicies: Policies{},
			requirement:      guard.Requirement(-1),
			errAssertion: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrUndefinedPolicy)
			},
		},
		{
			name:             "nil policy",
			requiredPolicies: []string{"undefined-policy"},
			declaredPolicies: Policies{},
			requirement:      guard.Requirement(-1),
			errAssertion: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrUndefinedPolicy)
			},
		},
		{
			name:             "error policy",
			requiredPolicies: []string{"positive-1", "positive-2"},
			declaredPolicies: Policies{
				"positive-1": func(ctx context.Context, input *Input) (bool, error) { return false, errors.New("error") },
				"positive-2": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			requirement:  guard.RequirementAll,
			errAssertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{
				policies: tt.declaredPolicies,
			}

			policyBased := &guard.PolicyBased{
				Policies:    tt.requiredPolicies,
				Requirement: tt.requirement,
			}

			allowed, err := i.evaluatePolicyBasedAccess(context.Background(), policyBased, &Input{})
			if tt.errAssertion != nil {
				tt.errAssertion(t, err)
				return
			}

			require.NoError(t, err)
			tt.allowAssertion(t, allowed)
		})
	}
}

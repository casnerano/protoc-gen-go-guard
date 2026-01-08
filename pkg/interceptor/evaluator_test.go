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

		allowAssertion assert.BoolAssertionFunc
		errAssertion   assert.ErrorAssertionFunc
	}{
		{
			name:           "nil rules without subject",
			rules:          nil,
			allowAssertion: assert.False,
		},
		{
			name:           "empty rules without subject",
			rules:          guard.Rules{},
			allowAssertion: assert.False,
		},
		{
			name: "one allow public rule without subject",
			rules: guard.Rules{
				{AllowPublic: guard.Ptr(true)},
			},
			allowAssertion: assert.True,
		},
		{
			name: "one require authentication rule without subject",
			rules: guard.Rules{
				{RequireAuthentication: guard.Ptr(true)},
			},
			allowAssertion: assert.False,
		},
		{
			name:  "multiple rules when one allow rule with subject",
			input: Input{Subject: &Subject{}},
			rules: guard.Rules{
				{RequireAuthentication: guard.Ptr(true)},
				{AllowPublic: guard.Ptr(false)},
			},
			allowAssertion: assert.True,
		},
		{
			name:  "multiple rules when not allow rules with subject",
			input: Input{Subject: &Subject{Roles: []string{}}},
			rules: guard.Rules{
				{
					AuthenticatedAccess: &guard.AuthenticatedAccess{
						PolicyBased: &guard.PolicyBased{
							Policies: []string{"negative-policy"},
							Match:    guard.MatchAll,
						},
					},
				},
				{
					AuthenticatedAccess: &guard.AuthenticatedAccess{
						RoleBased: &guard.RoleBased{
							Roles: []string{"non-existing-role"},
							Match: guard.MatchAll,
						},
					},
				},
			},
			policies: Policies{"negative-policy": func(ctx context.Context, input *Input) (bool, error) {
				return false, nil
			}},
			allowAssertion: assert.False,
		},
		{
			name:  "one policy rule when unknown match type with subject",
			input: Input{Subject: &Subject{}},
			rules: guard.Rules{
				{
					AuthenticatedAccess: &guard.AuthenticatedAccess{
						PolicyBased: &guard.PolicyBased{
							Policies: []string{"positive-policy"},
							Match:    guard.Match(-1),
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

			allowed, err := i.evaluateRules(context.Background(), tt.rules, &tt.input)
			if tt.errAssertion != nil {
				tt.errAssertion(t, err)
				return
			}

			require.NoError(t, err)
			tt.allowAssertion(t, allowed)
		})
	}
}

func Test_interceptor_evaluateRule(t *testing.T) {
	tests := []struct {
		name     string
		input    Input
		rule     *guard.Rule
		policies Policies

		allowAssertion assert.BoolAssertionFunc
		errAssertion   assert.ErrorAssertionFunc
	}{
		{
			name:           "empty rule without subject",
			input:          Input{},
			rule:           &guard.Rule{},
			allowAssertion: assert.False,
		},
		{
			name:           "allow public rule without subject",
			input:          Input{},
			rule:           &guard.Rule{AllowPublic: guard.Ptr(true)},
			allowAssertion: assert.True,
		},
		{
			name:           "require authentication without subject",
			input:          Input{},
			rule:           &guard.Rule{RequireAuthentication: guard.Ptr(true)},
			allowAssertion: assert.False,
		},
		{
			name:           "require authentication with subject",
			input:          Input{Subject: &Subject{}},
			rule:           &guard.Rule{RequireAuthentication: guard.Ptr(true)},
			allowAssertion: assert.True,
		},
		{
			name:  "role based access with no match",
			input: Input{Subject: &Subject{Roles: []string{"user"}}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles: []string{"admin"},
						Match: guard.MatchAll,
					},
				},
			},
			allowAssertion: assert.False,
		},
		{
			name:  "role based access with match",
			input: Input{Subject: &Subject{Roles: []string{"admin"}}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles: []string{"admin"},
						Match: guard.MatchAtLeastOne,
					},
				},
			},
			allowAssertion: assert.True,
		},
		{
			name:  "authenticated access with empty authenticated-rules and without subject",
			input: Input{},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{},
			},
			allowAssertion: assert.False,
		},
		{
			name:  "authenticated access with empty authenticated-rules and with subject",
			input: Input{Subject: &Subject{}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{},
			},
			allowAssertion: assert.False,
		},
		{
			name:  "policy based access with no match",
			input: Input{Subject: &Subject{}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies: []string{"positive-policy"},
						Match:    guard.MatchAll,
					},
				},
			},
			policies: Policies{
				"positive-policy": func(ctx context.Context, input *Input) (bool, error) {
					return false, nil
				},
			},
			allowAssertion: assert.False,
		},
		{
			name:  "policy based access with match",
			input: Input{Subject: &Subject{}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies: []string{"positive-policy"},
						Match:    guard.MatchAll,
					},
				},
			},
			policies: Policies{
				"positive-policy": func(ctx context.Context, input *Input) (bool, error) {
					return true, nil
				},
			},
			allowAssertion: assert.True,
		},
		{
			name:  "policy based access with unknown match type",
			input: Input{Subject: &Subject{}},
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies: []string{"positive-policy"},
						Match:    guard.Match(-1),
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
						Policies: []string{"undefined-policy"},
						Match:    guard.MatchAll,
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
						Policies: []string{"nil-policy"},
						Match:    guard.MatchAll,
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

			allowed, err := i.evaluateRule(context.Background(), tt.rule, &tt.input)
			if tt.errAssertion != nil {
				tt.errAssertion(t, err)
				return
			}

			require.NoError(t, err)
			tt.allowAssertion(t, allowed)
		})
	}
}

func Test_interceptor_evaluateRoleBasedAccess(t *testing.T) {
	tests := []struct {
		name          string
		requiredRoles []string
		subjectRoles  []string
		match         guard.Match

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
			name:           "match all required roles",
			requiredRoles:  []string{"admin", "manager"},
			subjectRoles:   []string{"admin", "manager"},
			match:          guard.MatchAll,
			allowAssertion: assert.True,
		},
		{
			name:           "no match all required roles",
			requiredRoles:  []string{"admin", "manager"},
			subjectRoles:   []string{"admin"},
			match:          guard.MatchAll,
			allowAssertion: assert.False,
		},
		{
			name:           "match at least one required roles",
			requiredRoles:  []string{"admin", "manager"},
			subjectRoles:   []string{"user", "manager"},
			match:          guard.MatchAtLeastOne,
			allowAssertion: assert.True,
		},
		{
			name:           "no match at least one required roles",
			requiredRoles:  []string{"admin", "manager"},
			subjectRoles:   []string{},
			match:          guard.MatchAtLeastOne,
			allowAssertion: assert.False,
		},
		{
			name:          "unknown match type",
			requiredRoles: []string{"admin"},
			subjectRoles:  []string{"admin"},
			match:         guard.Match(-1),
			errAssertion:  assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{}

			roleBased := &guard.RoleBased{
				Roles: tt.requiredRoles,
				Match: tt.match,
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
		match            guard.Match

		allowAssertion assert.BoolAssertionFunc
		errAssertion   assert.ErrorAssertionFunc
	}{
		{
			name:             "no required policies",
			requiredPolicies: []string{},
			allowAssertion:   assert.False,
		},
		{
			name:             "match all required policies",
			requiredPolicies: []string{"positive-policy-1", "positive-policy-2"},
			declaredPolicies: Policies{
				"positive-policy-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
				"positive-policy-2": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			match:          guard.MatchAll,
			allowAssertion: assert.True,
		},
		{
			name:             "no match all required policies",
			requiredPolicies: []string{"positive-policy-1", "negative-policy-1"},
			declaredPolicies: Policies{
				"positive-policy-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
				"negative-policy-1": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
			},
			match:          guard.MatchAll,
			allowAssertion: assert.False,
		},
		{
			name:             "match at least one required policies",
			requiredPolicies: []string{"negative-policy-1", "positive-policy-1"},
			declaredPolicies: Policies{
				"negative-policy-1": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
				"positive-policy-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			match:          guard.MatchAtLeastOne,
			allowAssertion: assert.True,
		},
		{
			name:             "no match at least one required policies",
			requiredPolicies: []string{"negative-policy-1", "negative-policy-2"},
			declaredPolicies: Policies{
				"negative-policy-1": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
				"negative-policy-2": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
			},
			match:          guard.MatchAtLeastOne,
			allowAssertion: assert.False,
		},
		{
			name:             "unknown match type",
			requiredPolicies: []string{"positive-1", "positive-2"},
			declaredPolicies: Policies{
				"positive-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
				"positive-2": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			match:        guard.Match(-1),
			errAssertion: assert.Error,
		},
		{
			name:             "undefined policy",
			requiredPolicies: []string{"undefined-policy"},
			declaredPolicies: Policies{},
			match:            guard.Match(-1),
			errAssertion: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrUndefinedPolicy)
			},
		},
		{
			name:             "nil policy",
			requiredPolicies: []string{"undefined-policy"},
			declaredPolicies: Policies{},
			match:            guard.Match(-1),
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
			match:        guard.MatchAll,
			errAssertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{
				policies: tt.declaredPolicies,
			}

			policyBased := &guard.PolicyBased{
				Policies: tt.requiredPolicies,
				Match:    tt.match,
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

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
		name         string
		roles        []string
		subjectRoles []string
		match        guard.Match
		wantAllowed  bool
		errAssertion bool
	}{
		{
			name:        "no roles",
			roles:       []string{},
			wantAllowed: false,
		},
		{
			name:         "match all: success",
			roles:        []string{"admin", "manager"},
			subjectRoles: []string{"admin", "manager"},
			match:        guard.MatchAll,
			wantAllowed:  true,
		},
		{
			name:         "match all: fail",
			roles:        []string{"admin", "manager"},
			subjectRoles: []string{"admin"},
			match:        guard.MatchAll,
			wantAllowed:  false,
		},
		{
			name:         "match at least one: success",
			roles:        []string{"admin", "manager"},
			subjectRoles: []string{"user", "manager"},
			match:        guard.MatchAtLeastOne,
			wantAllowed:  true,
		},
		{
			name:         "unknown match type",
			roles:        []string{"admin"},
			subjectRoles: []string{"admin"},
			match:        guard.Match(-1),
			wantAllowed:  false,
			errAssertion: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{}

			roleBased := &guard.RoleBased{
				Roles: tt.roles,
				Match: tt.match,
			}

			input := &Input{
				Subject: &Subject{
					Roles: tt.subjectRoles,
				},
			}

			allowed, err := i.evaluateRoleBasedAccess(context.Background(), roleBased, input)
			if tt.errAssertion {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantAllowed, allowed)
		})
	}
}

func Test_interceptor_evaluatePolicyBasedAccess(t *testing.T) {
	tests := []struct {
		name         string
		policies     []string
		policyMap    Policies
		match        guard.Match
		wantAllowed  bool
		errAssertion bool
	}{
		{
			name:        "no policies",
			policies:    []string{},
			wantAllowed: false,
		},
		{
			name:     "match all: success",
			policies: []string{"positive-1", "positive-2"},
			policyMap: Policies{
				"positive-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
				"positive-2": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			match:       guard.MatchAll,
			wantAllowed: true,
		},
		{
			name:     "match all: fail",
			policies: []string{"positive-1", "negative-1"},
			policyMap: Policies{
				"positive-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
				"negative-1": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
			},
			match:       guard.MatchAll,
			wantAllowed: false,
		},
		{
			name:     "match at least one: success",
			policies: []string{"negative-1", "positive-1"},
			policyMap: Policies{
				"negative-1": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
				"positive-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			match:       guard.MatchAtLeastOne,
			wantAllowed: true,
		},
		{
			name:     "match at least one: fail",
			policies: []string{"negative-1", "negative-2"},
			policyMap: Policies{
				"negative-1": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
				"negative-2": func(ctx context.Context, input *Input) (bool, error) { return false, nil },
			},
			match:       guard.MatchAtLeastOne,
			wantAllowed: false,
		},
		{
			name:     "unknown match type",
			policies: []string{"positive-1", "positive-2"},
			policyMap: Policies{
				"positive-1": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
				"positive-2": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			match:        guard.Match(-1),
			wantAllowed:  false,
			errAssertion: true,
		},
		{
			name:         "undefined policy",
			policies:     []string{"undefined-policy"},
			policyMap:    Policies{},
			match:        guard.Match(-1),
			wantAllowed:  false,
			errAssertion: true,
		},
		{
			name:     "policy error",
			policies: []string{"positive-1", "positive-2"},
			policyMap: Policies{
				"positive-1": func(ctx context.Context, input *Input) (bool, error) { return false, errors.New("error") },
				"positive-2": func(ctx context.Context, input *Input) (bool, error) { return true, nil },
			},
			match:        guard.MatchAll,
			wantAllowed:  false,
			errAssertion: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{
				policies: tt.policyMap,
			}

			policyBased := &guard.PolicyBased{
				Policies: tt.policies,
				Match:    tt.match,
			}

			allowed, err := i.evaluatePolicyBasedAccess(context.Background(), policyBased, &Input{})
			if tt.errAssertion {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantAllowed, allowed)
		})
	}
}

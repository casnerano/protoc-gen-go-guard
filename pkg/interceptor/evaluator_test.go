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
		name        string
		rule        *guard.Rule
		policies    Policies
		input       *Input
		wantAllowed bool
	}{
		{
			name:        "allow public",
			rule:        &guard.Rule{AllowPublic: guard.Ptr(true)},
			input:       &Input{},
			wantAllowed: true,
		},
		{
			name:        "require auth without subject",
			rule:        &guard.Rule{RequireAuthentication: guard.Ptr(true)},
			input:       &Input{},
			wantAllowed: false,
		},
		{
			name:        "require auth with subject",
			rule:        &guard.Rule{RequireAuthentication: guard.Ptr(true)},
			input:       &Input{Subject: &Subject{}},
			wantAllowed: true,
		},
		{
			name: "role based access: no match",
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles: []string{"admin"},
						Match: guard.MatchAll,
					},
				},
			},
			input:       &Input{Subject: &Subject{Roles: []string{"user"}}},
			wantAllowed: false,
		},
		{
			name: "role based access: match",
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles: []string{"admin"},
						Match: guard.MatchAtLeastOne,
					},
				},
			},
			input:       &Input{Subject: &Subject{Roles: []string{"admin"}}},
			wantAllowed: true,
		},
		{
			name: "policy based access: no match",
			rule: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies: []string{"test"},
						Match:    guard.MatchAll,
					},
				},
			},
			policies: Policies{
				"test": func(ctx context.Context, input *Input) (bool, error) {
					return false, nil
				},
			},
			input:       &Input{Subject: &Subject{}},
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{
				policies: tt.policies,
			}

			allowed, err := i.evaluateRule(context.Background(), tt.rule, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantAllowed, allowed)
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
			policies:     []string{"non-exists-1"},
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

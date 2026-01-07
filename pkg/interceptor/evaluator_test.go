package interceptor

import (
	"context"
	"testing"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
	"github.com/stretchr/testify/assert"
)

func Test_interceptor_evaluateRules(t *testing.T) {
	tests := []struct {
		name        string
		rules       guard.Rules
		input       *Input
		policies    Policies
		wantAllowed bool
		wantErr     bool
	}{
		{
			name:        "no rules returns false",
			rules:       nil,
			input:       &Input{},
			wantAllowed: false,
		},
		{
			name: "allow public rule grants access without authenticated",
			rules: guard.Rules{
				{AllowPublic: guard.Ptr(true)},
			},
			input:       &Input{Subject: nil},
			wantAllowed: true,
		},
		{
			name: "require auth rule denies access without auth",
			rules: guard.Rules{
				{RequireAuthentication: guard.Ptr(true)},
			},
			input:       &Input{Subject: nil},
			wantAllowed: false,
		},
		{
			name: "require auth rule grants access with auth",
			rules: guard.Rules{
				{RequireAuthentication: guard.Ptr(true)},
			},
			input:       &Input{Subject: &Subject{}},
			wantAllowed: true,
		},
		{
			name: "role based rule denies access without required roles",
			rules: guard.Rules{
				{
					AuthenticatedAccess: &guard.AuthenticatedAccess{
						RoleBased: &guard.RoleBased{
							Roles: []string{"admin"},
							Match: guard.MatchAll,
						},
					},
				},
			},
			input:       &Input{Subject: &Subject{Roles: []string{"user"}}},
			wantAllowed: false,
		},
		{
			name: "role based rule grants access with required roles",
			rules: guard.Rules{
				{
					AuthenticatedAccess: &guard.AuthenticatedAccess{
						RoleBased: &guard.RoleBased{
							Roles: []string{"admin"},
							Match: guard.MatchAll,
						},
					},
				},
			},
			input:       &Input{Subject: &Subject{Roles: []string{"admin"}}},
			wantAllowed: true,
		},
		{
			name: "policy based rule grants access when policy passes",
			rules: guard.Rules{
				{
					AuthenticatedAccess: &guard.AuthenticatedAccess{
						PolicyBased: &guard.PolicyBased{
							Policies: []string{"test-policy"},
							Match:    guard.MatchAtLeastOne,
						},
					},
				},
			},
			input:       &Input{Subject: &Subject{}},
			policies:    Policies{"test-policy": func(ctx context.Context, input *Input) (bool, error) { return true, nil }},
			wantAllowed: true,
		},
		{
			name: "multiple rules grants access if any rule passes",
			rules: guard.Rules{
				{RequireAuthentication: guard.Ptr(true)},
				{AllowPublic: guard.Ptr(true)},
			},
			input:       &Input{Subject: nil},
			wantAllowed: true,
		},
		{
			name: "empty authenticated access rule requires authenticated",
			rules: guard.Rules{
				{AuthenticatedAccess: &guard.AuthenticatedAccess{}},
			},
			input:       &Input{Subject: nil},
			wantAllowed: false,
		},
		{
			name: "invalid match type returns error",
			rules: guard.Rules{
				{
					AuthenticatedAccess: &guard.AuthenticatedAccess{
						RoleBased: &guard.RoleBased{
							Roles: []string{"admin"},
							Match: guard.Match(-1),
						},
					},
				},
			},
			input:       &Input{Subject: &Subject{Roles: []string{"admin"}}},
			wantAllowed: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{
				policies: tt.policies,
			}

			allowed, err := i.evaluateRules(context.Background(), tt.rules, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantAllowed, allowed)
		})
	}
}

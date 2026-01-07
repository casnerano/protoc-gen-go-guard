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
		policies    Policies
		input       *Input
		wantAllowed bool
		wantErr     bool
	}{
		{
			name:        "nil rules",
			rules:       nil,
			wantAllowed: false,
		},
		{
			name:        "empty rules",
			rules:       guard.Rules{},
			wantAllowed: false,
		},
		{
			name: "one allow rule",
			rules: guard.Rules{
				{AllowPublic: guard.Ptr(true)},
			},
			wantAllowed: true,
		},
		{
			name: "one deny rule",
			rules: guard.Rules{
				{RequireAuthentication: guard.Ptr(true)},
			},
			input:       &Input{Subject: nil},
			wantAllowed: false,
		},
		{
			name: "multiple rules with one allow",
			rules: guard.Rules{
				{RequireAuthentication: guard.Ptr(true)},
				{AllowPublic: guard.Ptr(false)},
			},
			input:       &Input{Subject: &Subject{}},
			wantAllowed: true,
		},
		{
			name: "multiple rules without allowed",
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
			policies:    Policies{"negative-policy": nil},
			input:       &Input{Subject: &Subject{Roles: []string{"user"}}},
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Interceptor{
				policies: tt.policies,
			}

			input := tt.input
			if input == nil {
				input = &Input{}
			}

			allowed, err := i.evaluateRules(context.Background(), tt.rules, input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantAllowed, allowed)
		})
	}
}

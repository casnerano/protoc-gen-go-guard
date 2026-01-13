package plugin

import (
	"testing"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
	desc "github.com/casnerano/protoc-gen-go-guard/proto"
	"github.com/stretchr/testify/assert"
)

func Test_extractRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		pbRule *desc.Rule
		want   *guard.Rule
	}{
		{
			name:   "nil rule",
			pbRule: nil,
			want:   &guard.Rule{},
		},
		{
			name: "allow public rule",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AllowPublic{AllowPublic: true},
			},
			want: &guard.Rule{
				AllowPublic: guard.Ptr(true),
			},
		},
		{
			name: "require authentication rule",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_RequireAuthentication{RequireAuthentication: true},
			},
			want: &guard.Rule{
				RequireAuthentication: guard.Ptr(true),
			},
		},
		{
			name: "authenticated access with role based",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						RoleBased: &desc.RoleBased{
							Roles: []string{"role1", "role2"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_ALL
								return &r
							}(),
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles:       []string{"role1", "role2"},
						Requirement: guard.RequirementAll,
					},
				},
			},
		},
		{
			name: "authenticated access with role based without requirement",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						RoleBased: &desc.RoleBased{
							Roles: []string{"role1", "role2"},
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles:       []string{"role1", "role2"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
		},
		{
			name: "authenticated access with role based at least one",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						RoleBased: &desc.RoleBased{
							Roles: []string{"role1", "role2"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_AT_LEAST_ONE
								return &r
							}(),
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles:       []string{"role1", "role2"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
		},
		{
			name: "authenticated access with policy based",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						PolicyBased: &desc.PolicyBased{
							Policies: []string{"policy1", "policy2"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_ALL
								return &r
							}(),
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"policy1", "policy2"},
						Requirement: guard.RequirementAll,
					},
				},
			},
		},
		{
			name: "authenticated access with policy based without requirement",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						PolicyBased: &desc.PolicyBased{
							Policies: []string{"policy1"},
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"policy1"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
		},
		{
			name: "authenticated access with policy based at least one",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						PolicyBased: &desc.PolicyBased{
							Policies: []string{"policy1", "policy2"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_AT_LEAST_ONE
								return &r
							}(),
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"policy1", "policy2"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
		},
		{
			name: "authenticated access with both role based and policy based",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						RoleBased: &desc.RoleBased{
							Roles: []string{"role1"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_ALL
								return &r
							}(),
						},
						PolicyBased: &desc.PolicyBased{
							Policies: []string{"policy1"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_AT_LEAST_ONE
								return &r
							}(),
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles:       []string{"role1"},
						Requirement: guard.RequirementAll,
					},
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"policy1"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
		},
		{
			name: "authenticated access with nil authenticated access",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: nil,
				},
			},
			want: &guard.Rule{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractRule(tt.pbRule)
			assert.Equal(t, tt.want, got)
		})
	}
}

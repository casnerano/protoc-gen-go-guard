package interceptor

import (
	"context"
	"fmt"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
)

type roleBasedEvaluator struct{}

func newRoleBasedEvaluator() *roleBasedEvaluator {
	return &roleBasedEvaluator{}
}

func (e roleBasedEvaluator) Evaluate(_ context.Context, rules *guard.Rules, authContext *AuthContext, _ any) (bool, error) {
	if !authContext.Authenticated {
		return false, nil
	}

	if rules.RoleBased == nil {
		return false, nil
	}

	switch rules.RoleBased.Requirement {
	case guard.RequirementAny:
		for _, allowedRole := range rules.RoleBased.AllowedRoles {
			for _, authRole := range authContext.Roles {
				if authRole == allowedRole {
					return true, nil
				}
			}
		}
		return false, nil
	case guard.RequirementAll:
		authContextRolesSet := make(map[string]struct{})
		for _, authRole := range authContext.Roles {
			authContextRolesSet[authRole] = struct{}{}
		}

		for _, requiredRole := range rules.RoleBased.AllowedRoles {
			if _, exists := authContextRolesSet[requiredRole]; !exists {
				return false, nil
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("unknown requirement type: %v", rules.RoleBased.Requirement)
	}
}

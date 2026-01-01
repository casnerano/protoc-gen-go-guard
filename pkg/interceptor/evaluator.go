package interceptor

import (
	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
)

type evaluatorType int

const (
	evaluatorTypeUnknown evaluatorType = iota
	evaluatorTypeAllowPublic
	evaluatorTypeRequireAuthn
	evaluatorTypeRoleBased
	evaluatorTypePolicyBased
)

func getEvaluatorType(rules *guard.Rules) evaluatorType {
	switch {
	case rules.AllowPublic != nil:
		return evaluatorTypeAllowPublic
	case rules.RequireAuthn != nil:
		return evaluatorTypeRequireAuthn
	case rules.RoleBased != nil:
		return evaluatorTypeRoleBased
	case rules.PolicyBased != nil:
		return evaluatorTypePolicyBased
	default:
		return evaluatorTypeUnknown
	}
}

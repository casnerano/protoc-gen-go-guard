package guard

type Requirement int

const (
	RequirementAny Requirement = iota
	RequirementAll
)

type Rules struct {
	AllowPublic  *bool
	RequireAuthn *bool
	RoleBased    *RoleBased
	PolicyBased  *PolicyBased
}

type RoleBased struct {
	AllowedRoles []string
	Requirement  Requirement
}

type PolicyBased struct {
	PolicyNames []string
	Requirement Requirement
}

type Service struct {
	Name    string
	Rules   *Rules
	Methods map[string]*Method
}

type Method struct {
	Rules *Rules
}

func Ptr[T any](v T) *T {
	return &v
}

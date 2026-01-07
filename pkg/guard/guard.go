package guard

type Match int

const (
	MatchAtLeastOne Match = iota
	MatchAll
)

type Rule struct {
	AllowPublic           *bool
	RequireAuthentication *bool
	AuthenticatedAccess   *AuthenticatedAccess
}

type Rules []*Rule

type AuthenticatedAccess struct {
	RoleBased   *RoleBased
	PolicyBased *PolicyBased
}

type RoleBased struct {
	Roles []string
	Match Match
}

type PolicyBased struct {
	Policies []string
	Match    Match
}

type Service struct {
	Name    string
	Rules   Rules
	Methods map[string]*Method
}

type Method struct {
	Rules Rules
}

func Ptr[T any](v T) *T {
	return &v
}

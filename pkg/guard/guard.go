// Package guard defines the data structures
// that represent access control rules for gRPC services and methods.
package guard

type Match int

const (
	MatchAtLeastOne Match = iota
	MatchAll
)

// Rule represents a single access control condition.
// Exactly one of
//   - AllowPublic — allows unauthenticated access;
//   - RequireAuthentication — requires authentication but no further checks;
//   - AuthenticatedAccess — fine-grained role- or policy-based access control.
type Rule struct {
	AllowPublic           *bool
	RequireAuthentication *bool
	AuthenticatedAccess   *AuthenticatedAccess
}

type Rules []*Rule

// AuthenticatedAccess defines access conditions for authenticated users,
// supporting role-based and/or policy-based checks.
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

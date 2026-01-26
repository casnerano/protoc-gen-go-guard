# protoc-gen-go-guard

Guard gRPC driven by Protobuf contract rules.

[![GitHub Release](https://img.shields.io/github/v/release/casnerano/protoc-gen-go-guard?color=green)](https://github.com/casnerano/protoc-gen-go-guard/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/casnerano/protoc-gen-go-guard)](https://goreportcard.com/report/github.com/casnerano/protoc-gen-go-guard)
[![Coverage](https://coveralls.io/repos/casnerano/protoc-gen-go-guard/badge.svg)](https://coveralls.io/r/casnerano/protoc-gen-go-guard)
[![GoDoc](https://pkg.go.dev/badge/github.com/casnerano/protoc-gen-go-guard)](https://godoc.org/github.com/casnerano/protoc-gen-go-guard)

## Features

- Define access rules directly in `.proto` files.
- Support for public, authenticated, role-based, and policy-based access
- Service-level and method-level rule inheritance.
- Zero-trust by default (deny all unless explicitly allowed).
- Simple interceptor: easy to plug into any gRPC server.
- No runtime reflection: fast and efficient.

## Installation

1. Install the plugin:
```sh
go install github.com/casnerano/protoc-gen-go-guard/cmd/protoc-gen-go-guard@latest
```

2. Add to your build system: for example using `protoc` directly:
```bash
protoc \
  --proto_path=. \
  \
  --plugin=protoc-gen-go-guard=bin/protoc-gen-go-guard \
  --go-guard_out=. \
  --go-guard_opt=paths=source_relative \
  \
  your_service.proto
```

## Usage

### 1. Annotate your proto
```protobuf
import "github.com/casnerano/protoc-gen-go-guard/proto/guard.proto";

service Auth {
  // Service is public by default.
  option (guard.service_rules) = { allow_public: true };

  // Inherits service rules (public access).
  rpc Register(google.protobuf.Empty) returns (google.protobuf.Empty);

  // Overrides service rules: requires authentication.
  rpc Logout(google.protobuf.Empty) returns (google.protobuf.Empty) {
    option (guard.method_rules) = { require_authentication: true };
  }

  // Overrides service rules: requirement passed at least one policies.
  rpc UpdateProfileStatus(google.protobuf.Empty) returns (google.protobuf.Empty) {
    option (guard.method_rules) = {
      authenticated_access: {
        policy_based: {
          policies: ["demo-period", "premium"],
          requirement: AT_LEAST_ONE
        }
      }
    };
  }

  // Overrides service rules: requirement to have all roles.
  rpc DeleteProfile(google.protobuf.Empty) returns (google.protobuf.Empty) {
    option (guard.method_rules) = {
      authenticated_access: {
        role_based: {
          roles: ["user", "verified"],
          requirement: ALL
        }
      }
    };
  }
}
```

### 2. Configure interceptor
```go
func main() {
	guard := interceptor.New(
		createSubjectResolver(),
		interceptor.WithPolicies(buildPolicies()),
		interceptor.WithDebug(),
	)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(guard.Unary()),
		grpc.StreamInterceptor(guard.Stream()),
	)
	
	// register services ...
}

// SubjectResolver is a function that extracts a Subject from the request context.
// If the user is unauthenticated, it should return (nil, nil).
//
// For example, extract roles from context metadata.
func createSubjectResolver() interceptor.SubjectResolver {
	return func(ctx context.Context) (*interceptor.Subject, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, nil
		}

		subject := interceptor.Subject{}
		if roles, exists := md["roles"]; exists {
			subject.Roles = roles
		}

		return &subject, nil
	}
}

func buildPolicies() interceptor.Policies {
	return interceptor.Policies{
		"active-subscription": func(ctx context.Context, input *interceptor.Input) (bool, error) {
			// Check if active subscription for current subject.
			return true, nil
		},
	}
}
```

### Rule inheritance hierarchy
Rules are evaluated in this order of precedence:
- **Method rules** — override service rules for specific methods.
- **Service rules** — apply to all methods in the service unless overridden.
- **Default rules** — apply when no other rules exist (configurable via interceptor, zero trust by default).

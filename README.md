# protoc-gen-go-guard

Guard gRPC driven by Protobuf contract rules.


[![GitHub Release](https://img.shields.io/github/v/release/casnerano/protoc-gen-go-guard?color=green)](https://github.com/casnerano/protoc-gen-go-guard/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/casnerano/protoc-gen-go-guard)](https://goreportcard.com/report/github.com/casnerano/protoc-gen-go-guard)
[![GoDoc](https://pkg.go.dev/badge/github.com/casnerano/protoc-gen-go-guard)](https://godoc.org/github.com/casnerano/protoc-gen-go-guard)
[![Coverage](https://coveralls.io/repos/casnerano/protoc-gen-go-guard/badge.svg)](https://coveralls.io/r/casnerano/protoc-gen-go-guard)

## Features

- Define Guard rules directly in `.proto` files.
- Support for public, authenticated, role-based, and policy-based access
- Service-level and method-level rule inheritance.
- Zero-trust by default (deny all unless explicitly allowed).
- Simple interceptor: Easy to plug into any gRPC server.
- No runtime reflection: Fast and efficient.

## Installation

1. Install the plugin:
```sh
go install github.com/casnerano/protoc-gen-go-guard/cmd/protoc-gen-go-guard@latest
```

2. Add to your build system (example for buf):
```yaml
# buf.yaml
plugins:
  - name: guard
    out: .
    path: protoc-gen-go-guard
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

  // Overrides service rules — requires passed at least one policy.
  rpc UpdateProfileStatus(google.protobuf.Empty) returns (google.protobuf.Empty) {
    option (guard.method_rules) = {
      authenticated_access: {
        policy_based: {
          policies: ["demo-period", "premium"],
          match: AT_LEAST_ONE
        }
      }
    };
  }

  // Overrides service rules — requires all specified roles.
  rpc DeleteProfile(google.protobuf.Empty) returns (google.protobuf.Empty) {
    option (guard.method_rules) = {
      authenticated_access: {
        role_based: {
          roles: ["user", "verified"],
          match: ALL
        }
      }
    };
  }
}
```

### 2. Configure interceptor
```go
guardInterceptor := interceptor.New(
    createSubjectResolver(),
    interceptor.WithPolicies(buildPolicies()),
    interceptor.WithDebug(),
)

server := grpc.NewServer(
    grpc.UnaryInterceptor(
        guardInterceptor.Unary(),
    ),
)

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
		"demo-period": func(ctx context.Context, input *interceptor.Input) (bool, error) {
			// check demo period with input
			return true, nil
		},
		"premium": func(ctx context.Context, input *interceptor.Input) (bool, error) {
			// check premium with input
			return false, nil
		},
	}
}
```
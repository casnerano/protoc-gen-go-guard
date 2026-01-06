# protoc-gen-go-guard

Guard gRPC driven by Protobuf contract rules.

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
server := grpc.NewServer(
  grpc.UnaryInterceptor(guard.Interceptor(
    func(ctx context.Context) (*guard.AuthContext, error) {
      // Extract user claims
      return &guard.AuthContext{Roles: []string{"admin"}}, nil
    },
    guard.WithPolicies(map[string]guard.Policy{
      "owner": func(ctx context.Context, auth *guard.AuthContext, req any) (bool, error) {
        // Custom policy logic
        return true, nil
      }
    }),
  )),
)
```
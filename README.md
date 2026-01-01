# protoc-gen-go-guard

A protoc plugin that generates gRPC Guard middleware from proto annotations.

## Features

- Define Guard rules directly in `.proto` files
- Support for public, authenticated, role-based, and policy-based access
- Zero-trust default (deny all unless explicitly allowed)
- Service-level and method-level rule inheritance

## Installation

1. Install the plugin:
```sh
go install github.com/casnerano/protoc-gen-go-guard/cmd/protoc-gen-go-guard@latest
```

2. Add to your build:
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
service UserService {
  option (guard.service_rules).allow_authenticated = true;
  
  rpc GetUser(GetRequest) returns (User) {
    option (guard.method_rules).role_based = {
      allowed_roles: ["admin", "support"],
      requirement: ANY
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
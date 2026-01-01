# protoc-gen-go-rbac

A protoc plugin that generates gRPC RBAC middleware from proto annotations.

## Features

- Define RBAC rules directly in `.proto` files
- Support for public, authenticated, role-based, and policy-based access
- Zero-trust default (deny all unless explicitly allowed)
- Service-level and method-level rule inheritance

## Installation

1. Install the plugin:
```sh
go install github.com/casnerano/protoc-gen-go-rbac/cmd/protoc-gen-go-rbac@latest
```

2. Add to your build:
```yaml
# buf.yaml
plugins:
  - name: rbac
    out: .
    path: protoc-gen-go-rbac
```

## Usage

### 1. Annotate your proto
```protobuf
service UserService {
  option (rbac.service_rules).allow_authenticated = true;
  
  rpc GetUser(GetRequest) returns (User) {
    option (rbac.method_rules).role_based = {
      allowed_roles: ["admin", "support"],
      requirement: ANY
    };
  }
}
```

### 2. Configure interceptor
```go
server := grpc.NewServer(
  grpc.UnaryInterceptor(rbac.Interceptor(
    func(ctx context.Context) (*rbac.AuthContext, error) {
      // Extract user claims
      return &rbac.AuthContext{Roles: []string{"admin"}}, nil
    },
    rbac.WithPolicies(map[string]rbac.Policy{
      "owner": func(ctx context.Context, auth *rbac.AuthContext, req any) (bool, error) {
        // Custom policy logic
        return true, nil
      }
    }),
  )),
)
```
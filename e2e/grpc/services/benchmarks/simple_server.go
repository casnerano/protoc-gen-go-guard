package benchmarks

import (
	"context"

	desc "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/benchmarks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	roleBasedAccessRequiredRoles = []string{"first", "second"}
)

type subject struct {
	roles []string
}

type SimpleServer struct {
	desc.UnimplementedBenchmarksServer
}

func (s *SimpleServer) PublicAccessMethod(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *SimpleServer) AuthenticatedAccessMethod(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	curSubject := subjectFromContext(ctx)
	if curSubject == nil {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return &emptypb.Empty{}, nil
}

func (s *SimpleServer) RoleBasedAccessMethod(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	curSubject := subjectFromContext(ctx)
	if curSubject == nil || !matchAllRequiredRoles(curSubject.roles) {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return &emptypb.Empty{}, nil
}

func NewSimpleServer() *grpc.Server {
	server := grpc.NewServer()

	desc.RegisterBenchmarksServer(server, &SimpleServer{})
	return server
}

func subjectFromContext(ctx context.Context) *subject {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	if _, authenticated := md["authenticated"]; !authenticated {
		return nil
	}

	curSubject := subject{}
	if roles, exists := md["roles"]; exists {
		curSubject.roles = roles
	}

	return nil
}

func matchAllRequiredRoles(subjectRoles []string) bool {
	if len(subjectRoles) == 0 {
		return false
	}

	var found bool
	for _, requiredRole := range roleBasedAccessRequiredRoles {
		found = false
		for _, subjectRole := range subjectRoles {
			if requiredRole == subjectRole {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

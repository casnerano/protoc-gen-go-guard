package benchmarks

import (
	"context"

	desc "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/benchmarks"
	"github.com/casnerano/protoc-gen-go-guard/pkg/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GuardServer struct {
	desc.UnimplementedBenchmarksServer
}

func (s *GuardServer) PublicAccessMethod(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *GuardServer) AuthenticatedAccessMethod(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *GuardServer) RoleBasedAccessMethod(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func NewGuardServer() *grpc.Server {
	guardInterceptor := interceptor.New(
		func(ctx context.Context) (*interceptor.Subject, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, nil
			}

			if _, authenticated := md["authenticated"]; !authenticated {
				return nil, nil
			}

			subject := interceptor.Subject{}
			if roles, exists := md["roles"]; exists {
				subject.Roles = roles
			}

			return &subject, nil
		},
	)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			guardInterceptor.Unary(),
		),
	)

	desc.RegisterBenchmarksServer(server, &GuardServer{})
	return server
}

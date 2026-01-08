package corner_cases

import (
	"context"

	desc "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/corner_cases"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MixedTypesAccessServer struct {
	desc.UnimplementedMixedTypesAccessServer
}

func (m *MixedTypesAccessServer) OverrideWithPolicies(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *MixedTypesAccessServer) OverrideWithRoles(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (m *MixedTypesAccessServer) OverrideWithAuthentication(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

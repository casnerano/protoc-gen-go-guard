package corner_cases

import (
	"context"

	"github.com/casnerano/protoc-gen-go-guard/example/pb/corner_cases"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type MixedTypesAccessServer struct {
	corner_cases.UnimplementedMixedTypesAccessServer
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

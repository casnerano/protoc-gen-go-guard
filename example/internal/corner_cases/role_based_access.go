package corner_cases

import (
	"context"

	"github.com/casnerano/protoc-gen-go-guard/example/pb/corner_cases"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RoleBasedAccessServer struct {
	corner_cases.UnimplementedRoleBasedAccessServer
}

func (a *RoleBasedAccessServer) EmptyRolesWithAnyRequirement(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (a *RoleBasedAccessServer) EmptyRolesWithAllRequirement(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (a *RoleBasedAccessServer) MultipleRolesWithAnyRequirement(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (a *RoleBasedAccessServer) MultipleRolesWithAllRequirement(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

package corner_cases

import (
	"context"

	desc "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/corner_cases"
	"google.golang.org/protobuf/types/known/emptypb"
)

type InheritAndOverrideOneServer struct {
	desc.UnimplementedInheritAndOverrideOneServer
}

func (i InheritAndOverrideOneServer) InheritedMethod(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (i InheritAndOverrideOneServer) OverriddenMethod(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type InheritAndOverrideTwoServer struct {
	desc.UnimplementedInheritAndOverrideTwoServer
}

func (i InheritAndOverrideTwoServer) InheritedMethod(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (i InheritAndOverrideTwoServer) OverriddenMethod(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

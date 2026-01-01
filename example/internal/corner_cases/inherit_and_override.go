package corner_cases

import (
	"context"

	"github.com/casnerano/protoc-gen-go-guard/example/pb/corner_cases"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type InheritAndOverrideOneServer struct {
	corner_cases.UnimplementedInheritAndOverrideOneServer
}

func (i InheritAndOverrideOneServer) InheritedMethod(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (i InheritAndOverrideOneServer) OverriddenMethod(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type InheritAndOverrideTwoServer struct {
	corner_cases.UnimplementedInheritAndOverrideTwoServer
}

func (i InheritAndOverrideTwoServer) InheritedMethod(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (i InheritAndOverrideTwoServer) OverriddenMethod(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

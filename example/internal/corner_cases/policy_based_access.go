package corner_cases

import (
	"context"

	"github.com/casnerano/protoc-gen-go-rbac/example/pb/corner_cases"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type PolicyBasedAccessServer struct {
	corner_cases.UnimplementedPolicyBasedAccessServer
}

func (p *PolicyBasedAccessServer) EmptyPoliciesWithAnyRequirement(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (p *PolicyBasedAccessServer) EmptyPoliciesWithAllRequirement(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (p *PolicyBasedAccessServer) MultiplePoliciesWithAnyRequirement(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (p *PolicyBasedAccessServer) MultiplePoliciesWithAllRequirement(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

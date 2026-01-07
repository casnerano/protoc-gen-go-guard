//go:build integration

package corner_cases

import (
	"context"
	"testing"

	desc "github.com/casnerano/protoc-gen-go-guard/example/pb/corner_cases"
	"github.com/casnerano/protoc-gen-go-guard/pkg/interceptor"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PolicyBasedAccessTestsSuite struct {
	CornerCasesServerTestSuite

	client desc.PolicyBasedAccessClient
}

func (s *PolicyBasedAccessTestsSuite) SetupSuite() {
	s.CornerCasesServerTestSuite.SetupSuite()
	desc.RegisterPolicyBasedAccessServer(s.server, &PolicyBasedAccessServer{})
	s.StartServer()

	client, err := s.GetClientConn()
	s.Require().NoError(err, "Failed to dial test server")

	s.client = desc.NewPolicyBasedAccessClient(client)
}

func (s *PolicyBasedAccessTestsSuite) TestEmptyPoliciesWithAnyRequirement() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied for authenticated and empty policies with at least one matching",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.EmptyPoliciesWithAnyRequirement(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

func (s *PolicyBasedAccessTestsSuite) TestMultiplePoliciesWithAllRequirement() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied for authenticated and empty policies with all matching",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.MultiplePoliciesWithAllRequirement(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

func TestPolicyBasedAccessTests(t *testing.T) {
	suite.Run(t, new(PolicyBasedAccessTestsSuite))
}

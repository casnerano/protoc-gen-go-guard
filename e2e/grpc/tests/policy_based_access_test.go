//go:build e2e

package tests

import (
	"context"
	"testing"

	desc "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/corner_cases"
	"github.com/casnerano/protoc-gen-go-guard/e2e/grpc/services"
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
	desc.RegisterPolicyBasedAccessServer(s.server, &services.PolicyBasedAccessServer{})
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
			name:      "access denied for authenticated and not all policies passed",
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

func (s *PolicyBasedAccessTestsSuite) TestMultiplePoliciesWithAnyRequirement() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied for authenticated and not all policies passed",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.MultiplePoliciesWithAnyRequirement(tt.context, &emptypb.Empty{})
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

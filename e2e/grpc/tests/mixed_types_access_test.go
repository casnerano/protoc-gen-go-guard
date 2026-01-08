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

type MixedTypesAccessServerTestSuite struct {
	CornerCasesServerTestSuite

	client desc.MixedTypesAccessClient
}

func (s *MixedTypesAccessServerTestSuite) SetupSuite() {
	s.CornerCasesServerTestSuite.SetupSuite()
	desc.RegisterMixedTypesAccessServer(s.server, &services.MixedTypesAccessServer{})
	s.StartServer()

	client, err := s.GetClientConn()
	s.Require().NoError(err, "Failed to dial test server")

	s.client = desc.NewMixedTypesAccessClient(client)
}

func (s *MixedTypesAccessServerTestSuite) TestOverrideWithAuthentication() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied for unauthenticated",
			context:   context.Background(),
			canAccess: false,
		},
		{
			name:      "access allowed for authenticated",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: true,
		},
	}
	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.OverrideWithAuthentication(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

func (s *MixedTypesAccessServerTestSuite) TestOverrideWithRoles() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied for unauthenticated",
			context:   context.Background(),
			canAccess: false,
		},
		{
			name:      "access allowed for authenticated and with all necessary roles",
			context:   testContextWithSubject(interceptor.Subject{Roles: []string{"admin"}}),
			canAccess: true,
		},
		{
			name:      "access allowed for authenticated and without necessary roles",
			context:   testContextWithSubject(interceptor.Subject{Roles: []string{"user"}}),
			canAccess: false,
		},
	}
	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.OverrideWithRoles(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

func (s *MixedTypesAccessServerTestSuite) TestOverrideWithPolicies() {
	testCases := []struct {
		name      string
		context   context.Context
		canAccess bool
	}{
		{
			name:      "access denied for unauthenticated",
			context:   context.Background(),
			canAccess: false,
		},
		{
			name:      "access allowed for authenticated and passed all necessary policies",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: true,
		},
	}
	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.OverrideWithPolicies(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

func TestMixedTypesAccessServer(t *testing.T) {
	suite.Run(t, new(MixedTypesAccessServerTestSuite))
}

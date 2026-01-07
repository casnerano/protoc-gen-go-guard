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

type RoleBasedAccessServerTestSuite struct {
	CornerCasesServerTestSuite

	client desc.RoleBasedAccessClient
}

func (s *RoleBasedAccessServerTestSuite) SetupSuite() {
	s.CornerCasesServerTestSuite.SetupSuite()
	desc.RegisterRoleBasedAccessServer(s.server, &RoleBasedAccessServer{})
	s.StartServer()

	client, err := s.GetClientConn()
	s.Require().NoError(err, "Failed to dial test server")

	s.client = desc.NewRoleBasedAccessClient(client)
}

func (s *RoleBasedAccessServerTestSuite) TestEmptyRolesWithAnyRequirement() {
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
			name:      "access denied for authenticated and without roles",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.EmptyRolesWithAnyRequirement(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

func (s *RoleBasedAccessServerTestSuite) TestEmptyRolesWithAllRequirement() {
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
			name:      "access denied for authenticated and without roles",
			context:   testContextWithSubject(interceptor.Subject{}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.EmptyRolesWithAllRequirement(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

func (s *RoleBasedAccessServerTestSuite) TestMultipleRolesWithAnyRequirement() {
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
			name: "access allowed with token and with one required role",
			context: testContextWithSubject(interceptor.Subject{
				Roles: []string{"admin"},
			}),
			canAccess: true,
		},
		{
			name: "access allowed with token and without required roles",
			context: testContextWithSubject(interceptor.Subject{
				Roles: []string{"non-exists-role"},
			}),
			canAccess: false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.MultipleRolesWithAnyRequirement(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

func (s *RoleBasedAccessServerTestSuite) TestMultipleRolesWithAllRequirement() {
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
			name: "access denied with token and with one required role",
			context: testContextWithSubject(interceptor.Subject{
				Roles: []string{"admin"},
			}),
			canAccess: false,
		},
		{
			name: "access allowed with token and with all required roles",
			context: testContextWithSubject(interceptor.Subject{
				Roles: []string{"admin", "manager"},
			}),
			canAccess: true,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.MultipleRolesWithAllRequirement(tt.context, &emptypb.Empty{})
			if tt.canAccess {
				s.NoError(err)
			} else {
				s.Equal(codes.PermissionDenied.String(), status.Code(err).String())
			}
		})
	}
}

func TestRoleBasedAccessServer(t *testing.T) {
	suite.Run(t, new(RoleBasedAccessServerTestSuite))
}

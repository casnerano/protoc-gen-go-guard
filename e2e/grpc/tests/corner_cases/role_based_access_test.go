//go:build e2e

package corner_cases

import (
	"context"
	"testing"

	desc "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/pb/corner_cases"
	services "github.com/casnerano/protoc-gen-go-guard/e2e/grpc/services/corner_cases"
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
	desc.RegisterRoleBasedAccessServer(s.server, &services.RoleBasedAccessServer{})
	s.StartServer()

	client, err := s.GetClientConn()
	s.Require().NoError(err, "Failed to dial test server")

	s.client = desc.NewRoleBasedAccessClient(client)
}

func (s *RoleBasedAccessServerTestSuite) TestEmptyRolesWithAnyRequirement() {
	testCases := []struct {
		name         string
		context      context.Context
		expectedCode codes.Code
	}{
		{
			name:         "access denied for unauthenticated",
			context:      context.Background(),
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "access denied for authenticated and without roles",
			context:      testContextWithSubject(interceptor.Subject{}),
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.EmptyRolesWithAnyRequirement(tt.context, &emptypb.Empty{})
			if tt.expectedCode == codes.OK {
				s.NoError(err)
			} else {
				s.Equal(tt.expectedCode, status.Code(err))
			}
		})
	}
}

func (s *RoleBasedAccessServerTestSuite) TestEmptyRolesWithAllRequirement() {
	testCases := []struct {
		name         string
		context      context.Context
		expectedCode codes.Code
	}{
		{
			name:         "access denied for unauthenticated",
			context:      context.Background(),
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "access denied for authenticated and without roles",
			context:      testContextWithSubject(interceptor.Subject{}),
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.EmptyRolesWithAllRequirement(tt.context, &emptypb.Empty{})
			if tt.expectedCode == codes.OK {
				s.NoError(err)
			} else {
				s.Equal(tt.expectedCode, status.Code(err))
			}
		})
	}
}

func (s *RoleBasedAccessServerTestSuite) TestMultipleRolesWithAnyRequirement() {
	testCases := []struct {
		name         string
		context      context.Context
		expectedCode codes.Code
	}{
		{
			name:         "access denied for unauthenticated",
			context:      context.Background(),
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "access allowed with token and with one required role",
			context: testContextWithSubject(interceptor.Subject{
				Roles: []string{"admin"},
			}),
			expectedCode: codes.OK,
		},
		{
			name: "access allowed with token and without required roles",
			context: testContextWithSubject(interceptor.Subject{
				Roles: []string{"non-exists-role"},
			}),
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.MultipleRolesWithAnyRequirement(tt.context, &emptypb.Empty{})
			if tt.expectedCode == codes.OK {
				s.NoError(err)
			} else {
				s.Equal(tt.expectedCode, status.Code(err))
			}
		})
	}
}

func (s *RoleBasedAccessServerTestSuite) TestMultipleRolesWithAllRequirement() {
	testCases := []struct {
		name         string
		context      context.Context
		expectedCode codes.Code
	}{
		{
			name:         "access denied for unauthenticated",
			context:      context.Background(),
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "access denied with token and with one required role",
			context: testContextWithSubject(interceptor.Subject{
				Roles: []string{"admin"},
			}),
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "access allowed with token and with all required roles",
			context: testContextWithSubject(interceptor.Subject{
				Roles: []string{"admin", "manager"},
			}),
			expectedCode: codes.OK,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			_, err := s.client.MultipleRolesWithAllRequirement(tt.context, &emptypb.Empty{})
			if tt.expectedCode == codes.OK {
				s.NoError(err)
			} else {
				s.Equal(tt.expectedCode, status.Code(err))
			}
		})
	}
}

func TestRoleBasedAccessServer(t *testing.T) {
	suite.Run(t, new(RoleBasedAccessServerTestSuite))
}

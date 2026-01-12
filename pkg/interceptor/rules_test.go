package interceptor

import (
	"testing"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
	"github.com/stretchr/testify/assert"
)

func Test_interceptor_getGuardService(t *testing.T) {
	data := struct {
		service *guard.Service
	}{
		service: &guard.Service{
			Name: "service",
		},
	}

	tests := []struct {
		name   string
		server any
		want   *guard.Service
	}{
		{
			name:   "server implements guardServiceProvider",
			server: &mockGuardServiceProvider{service: data.service},
			want:   data.service,
		},
		{
			name:   "server not implement guardServiceProvider",
			server: &struct{}{},
			want:   nil,
		},
		{
			name:   "server is nil",
			server: nil,
			want:   nil,
		},
	}

	i := &Interceptor{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := i.getGuardService(tt.server)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_interceptor_getRules(t *testing.T) {
	data := struct {
		allowPublicRule         *guard.Rule
		requireAuthRule         *guard.Rule
		authenticatedAccessRule *guard.Rule
	}{
		allowPublicRule: &guard.Rule{AllowPublic: guard.Ptr(true)},
		requireAuthRule: &guard.Rule{RequireAuthentication: guard.Ptr(true)},
		authenticatedAccessRule: &guard.Rule{
			AuthenticatedAccess: &guard.AuthenticatedAccess{
				RoleBased: &guard.RoleBased{
					Roles:       []string{"admin"},
					Requirement: guard.RequirementAtLeastOne,
				},
			},
		},
	}

	tests := []struct {
		Name         string
		Service      *guard.Service
		defaultRules guard.Rules
		fullMethod   string
		want         guard.Rules
	}{
		{
			Name:         "nil service with nil default rules returns nil",
			Service:      nil,
			defaultRules: nil,
			fullMethod:   "/pkg.Service/Method",
			want:         nil,
		},
		{
			Name: "returns default rules when no service or method rules exist",
			Service: &guard.Service{
				Name:    "Service",
				Methods: map[string]*guard.Method{},
			},
			fullMethod:   "/pkg.Service/Method",
			defaultRules: guard.Rules{data.allowPublicRule},
			want:         guard.Rules{data.allowPublicRule},
		},
		{
			Name: "method rules take precedence over service rules",
			Service: &guard.Service{
				Name:  "Service",
				Rules: guard.Rules{data.requireAuthRule},
				Methods: map[string]*guard.Method{
					"Method": {Rules: guard.Rules{data.allowPublicRule}},
				},
			},
			fullMethod:   "/pkg.Service/Method",
			defaultRules: nil,
			want:         guard.Rules{data.allowPublicRule},
		},
		{
			Name: "service rules used when no method rules exist",
			Service: &guard.Service{
				Name:    "Service",
				Rules:   guard.Rules{data.requireAuthRule},
				Methods: map[string]*guard.Method{},
			},
			fullMethod:   "/pkg.Service/Method",
			defaultRules: nil,
			want:         guard.Rules{data.requireAuthRule},
		},
		{
			Name: "empty method rules override service rules",
			Service: &guard.Service{
				Name:  "Service",
				Rules: guard.Rules{data.requireAuthRule},
				Methods: map[string]*guard.Method{
					"Method": {Rules: guard.Rules{{}}},
				},
			},
			fullMethod:   "/pkg.Service/Method",
			defaultRules: guard.Rules{data.allowPublicRule},
			want:         guard.Rules{{}},
		},
		{
			Name:         "nil service returns nil even with default rules",
			Service:      nil,
			fullMethod:   "/pkg.Unknown/Method",
			defaultRules: guard.Rules{data.allowPublicRule},
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var server any
			if tt.Service != nil {
				server = mockGuardServiceProvider{service: tt.Service}
			} else {
				server = &struct{}{}
			}

			i := &Interceptor{defaultRules: tt.defaultRules}

			got := i.getRules(server, tt.fullMethod)
			assert.Equal(t, tt.want, got)
		})
	}
}

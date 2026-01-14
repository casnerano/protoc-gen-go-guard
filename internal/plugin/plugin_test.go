package plugin

import (
	"testing"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
	desc "github.com/casnerano/protoc-gen-go-guard/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func Test_extractRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		pbRule *desc.Rule
		want   *guard.Rule
	}{
		{
			name:   "nil rule",
			pbRule: nil,
			want:   nil,
		},
		{
			name: "allow public rule",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AllowPublic{AllowPublic: true},
			},
			want: &guard.Rule{
				AllowPublic: guard.Ptr(true),
			},
		},
		{
			name: "require authentication rule",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_RequireAuthentication{RequireAuthentication: true},
			},
			want: &guard.Rule{
				RequireAuthentication: guard.Ptr(true),
			},
		},
		{
			name: "authenticated access with role based",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						RoleBased: &desc.RoleBased{
							Roles: []string{"role1", "role2"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_ALL
								return &r
							}(),
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles:       []string{"role1", "role2"},
						Requirement: guard.RequirementAll,
					},
				},
			},
		},
		{
			name: "authenticated access with role based without requirement",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						RoleBased: &desc.RoleBased{
							Roles: []string{"role1", "role2"},
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles:       []string{"role1", "role2"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
		},
		{
			name: "authenticated access with role based at least one",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						RoleBased: &desc.RoleBased{
							Roles: []string{"role1", "role2"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_AT_LEAST_ONE
								return &r
							}(),
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles:       []string{"role1", "role2"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
		},
		{
			name: "authenticated access with policy based",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						PolicyBased: &desc.PolicyBased{
							Policies: []string{"policy1", "policy2"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_ALL
								return &r
							}(),
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"policy1", "policy2"},
						Requirement: guard.RequirementAll,
					},
				},
			},
		},
		{
			name: "authenticated access with policy based without requirement",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						PolicyBased: &desc.PolicyBased{
							Policies: []string{"policy1"},
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"policy1"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
		},
		{
			name: "authenticated access with policy based at least one",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						PolicyBased: &desc.PolicyBased{
							Policies: []string{"policy1", "policy2"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_AT_LEAST_ONE
								return &r
							}(),
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"policy1", "policy2"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
		},
		{
			name: "authenticated access with both role based and policy based",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: &desc.AuthenticatedAccess{
						RoleBased: &desc.RoleBased{
							Roles: []string{"role1"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_ALL
								return &r
							}(),
						},
						PolicyBased: &desc.PolicyBased{
							Policies: []string{"policy1"},
							Requirement: func() *desc.Requirement {
								r := desc.Requirement_AT_LEAST_ONE
								return &r
							}(),
						},
					},
				},
			},
			want: &guard.Rule{
				AuthenticatedAccess: &guard.AuthenticatedAccess{
					RoleBased: &guard.RoleBased{
						Roles:       []string{"role1"},
						Requirement: guard.RequirementAll,
					},
					PolicyBased: &guard.PolicyBased{
						Policies:    []string{"policy1"},
						Requirement: guard.RequirementAtLeastOne,
					},
				},
			},
		},
		{
			name: "authenticated access with nil authenticated access",
			pbRule: &desc.Rule{
				Mode: &desc.Rule_AuthenticatedAccess{
					AuthenticatedAccess: nil,
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractRule(tt.pbRule)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func testConvertGuardRuleToProtoRule(rule *guard.Rule) *desc.Rule {
	if rule == nil {
		return nil
	}

	if rule.AllowPublic != nil {
		return &desc.Rule{Mode: &desc.Rule_AllowPublic{AllowPublic: *rule.AllowPublic}}
	}

	if rule.RequireAuthentication != nil {
		return &desc.Rule{Mode: &desc.Rule_RequireAuthentication{RequireAuthentication: *rule.RequireAuthentication}}
	}

	if rule.AuthenticatedAccess != nil {
		authAccess := &desc.AuthenticatedAccess{}

		if rule.AuthenticatedAccess.RoleBased != nil {
			req := desc.Requirement_AT_LEAST_ONE
			if rule.AuthenticatedAccess.RoleBased.Requirement == guard.RequirementAll {
				req = desc.Requirement_ALL
			}

			authAccess.RoleBased = &desc.RoleBased{
				Roles:       rule.AuthenticatedAccess.RoleBased.Roles,
				Requirement: &req,
			}
		}

		if rule.AuthenticatedAccess.PolicyBased != nil {
			req := desc.Requirement_AT_LEAST_ONE
			if rule.AuthenticatedAccess.PolicyBased.Requirement == guard.RequirementAll {
				req = desc.Requirement_ALL
			}

			authAccess.PolicyBased = &desc.PolicyBased{
				Policies:    rule.AuthenticatedAccess.PolicyBased.Policies,
				Requirement: &req,
			}
		}

		return &desc.Rule{Mode: &desc.Rule_AuthenticatedAccess{AuthenticatedAccess: authAccess}}
	}

	return nil
}

func testCreateServiceDescriptors(services []*guard.Service) protoreflect.ServiceDescriptors {
	messageDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("Empty"),
	}

	serviceProtos := make([]*descriptorpb.ServiceDescriptorProto, 0, len(services))
	for _, service := range services {
		serviceMethods := make([]*descriptorpb.MethodDescriptorProto, 0, len(service.Methods))

		for methodName, method := range service.Methods {
			methodProto := &descriptorpb.MethodDescriptorProto{
				Name:       proto.String(methodName),
				InputType:  proto.String("test.Empty"),
				OutputType: proto.String("test.Empty"),
			}

			if len(method.Rules) > 0 {
				pbRules := make([]*desc.Rule, 0, len(method.Rules))
				for _, rule := range method.Rules {
					if pbRule := testConvertGuardRuleToProtoRule(rule); pbRule != nil {
						pbRules = append(pbRules, pbRule)
					}
				}

				if len(pbRules) > 0 {
					opts := &descriptorpb.MethodOptions{}
					proto.SetExtension(opts, desc.E_MethodRules, pbRules)
					methodProto.Options = opts
				}
			}

			serviceMethods = append(serviceMethods, methodProto)
		}

		serviceOptions := &descriptorpb.ServiceOptions{}
		if len(service.Rules) > 0 {
			pbRules := make([]*desc.Rule, 0, len(service.Rules))
			for _, rule := range service.Rules {
				if pbRule := testConvertGuardRuleToProtoRule(rule); pbRule != nil {
					pbRules = append(pbRules, pbRule)
				}
			}

			if len(pbRules) > 0 {
				proto.SetExtension(serviceOptions, desc.E_ServiceRules, pbRules)
			}
		}

		serviceProtos = append(serviceProtos, &descriptorpb.ServiceDescriptorProto{
			Name:    proto.String(service.Name),
			Method:  serviceMethods,
			Options: serviceOptions,
		})
	}

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:        proto.String("test.proto"),
		Package:     proto.String("test"),
		Syntax:      proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{messageDesc},
		Service:     serviceProtos,
	}

	fd, err := protodesc.NewFile(fileDesc, nil)
	if err != nil {
		panic(err)
	}

	return fd.Services()
}

func Test_collectMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service *guard.Service
		want    map[string]*guard.Method
	}{
		{
			name: "empty methods",
			service: &guard.Service{
				Name:    "Service1",
				Methods: map[string]*guard.Method{},
			},
			want: map[string]*guard.Method{},
		},
		{
			name: "methods without rules",
			service: &guard.Service{
				Name: "Service1",
				Methods: map[string]*guard.Method{
					"Method1": {Rules: nil},
					"Method2": {Rules: guard.Rules{}},
				},
			},
			want: map[string]*guard.Method{},
		},
		{
			name: "method with allow public rule",
			service: &guard.Service{
				Name: "Service1",
				Methods: map[string]*guard.Method{
					"Method1": {
						Rules: guard.Rules{
							{AllowPublic: guard.Ptr(true)},
						},
					},
				},
			},
			want: map[string]*guard.Method{
				"Method1": {
					Rules: guard.Rules{
						{AllowPublic: guard.Ptr(true)},
					},
				},
			},
		},
		{
			name: "method with require authentication rule",
			service: &guard.Service{
				Name: "Service1",
				Methods: map[string]*guard.Method{
					"Method1": {
						Rules: guard.Rules{
							{RequireAuthentication: guard.Ptr(true)},
						},
					},
				},
			},
			want: map[string]*guard.Method{
				"Method1": {
					Rules: guard.Rules{
						{RequireAuthentication: guard.Ptr(true)},
					},
				},
			},
		},
		{
			name: "multiple methods with different rules",
			service: &guard.Service{
				Name: "Service1",
				Methods: map[string]*guard.Method{
					"Method1": {
						Rules: guard.Rules{
							{AllowPublic: guard.Ptr(true)},
						},
					},
					"Method2": {
						Rules: guard.Rules{
							{RequireAuthentication: guard.Ptr(true)},
						},
					},
					"Method3": {
						Rules: nil,
					},
				},
			},
			want: map[string]*guard.Method{
				"Method1": {
					Rules: guard.Rules{
						{AllowPublic: guard.Ptr(true)},
					},
				},
				"Method2": {
					Rules: guard.Rules{
						{RequireAuthentication: guard.Ptr(true)},
					},
				},
			},
		},
		{
			name: "method with empty rules",
			service: &guard.Service{
				Name: "Service1",
				Methods: map[string]*guard.Method{
					"Method1": {
						Rules: guard.Rules{},
					},
				},
			},
			want: map[string]*guard.Method{},
		},
		{
			name: "method with authenticated access rule",
			service: &guard.Service{
				Name: "Service1",
				Methods: map[string]*guard.Method{
					"Method1": {
						Rules: guard.Rules{
							{
								AuthenticatedAccess: &guard.AuthenticatedAccess{
									RoleBased: &guard.RoleBased{
										Roles:       []string{"role1", "role2"},
										Requirement: guard.RequirementAll,
									},
								},
							},
						},
					},
				},
			},
			want: map[string]*guard.Method{
				"Method1": {
					Rules: guard.Rules{
						{
							AuthenticatedAccess: &guard.AuthenticatedAccess{
								RoleBased: &guard.RoleBased{
									Roles:       []string{"role1", "role2"},
									Requirement: guard.RequirementAll,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			serviceDescs := testCreateServiceDescriptors([]*guard.Service{tt.service})
			got := collectMethods(serviceDescs.Get(0).Methods())
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func Test_collectServices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		services []*guard.Service
		want     []*guard.Service
	}{
		{
			name:     "empty services",
			services: []*guard.Service{},
			want:     nil,
		},
		{
			name: "services without rules",
			services: []*guard.Service{
				{
					Name:    "Service1",
					Rules:   nil,
					Methods: map[string]*guard.Method{},
				},
				{
					Name:    "Service2",
					Rules:   guard.Rules{},
					Methods: map[string]*guard.Method{},
				},
			},
			want: nil,
		},
		{
			name: "service with service-level rules only",
			services: []*guard.Service{
				{
					Name: "Service1",
					Rules: guard.Rules{
						{AllowPublic: guard.Ptr(true)},
					},
					Methods: map[string]*guard.Method{},
				},
			},
			want: []*guard.Service{
				{
					Name: "Service1",
					Rules: guard.Rules{
						{AllowPublic: guard.Ptr(true)},
					},
					Methods: nil,
				},
			},
		},
		{
			name: "service with method-level rules only",
			services: []*guard.Service{
				{
					Name:  "Service1",
					Rules: nil,
					Methods: map[string]*guard.Method{
						"Method1": {
							Rules: guard.Rules{
								{RequireAuthentication: guard.Ptr(true)},
							},
						},
					},
				},
			},
			want: []*guard.Service{
				{
					Name:  "Service1",
					Rules: nil,
					Methods: map[string]*guard.Method{
						"Method1": {
							Rules: guard.Rules{
								{RequireAuthentication: guard.Ptr(true)},
							},
						},
					},
				},
			},
		},
		{
			name: "service with both service-level and method-level rules",
			services: []*guard.Service{
				{
					Name: "Service1",
					Rules: guard.Rules{
						{AllowPublic: guard.Ptr(true)},
					},
					Methods: map[string]*guard.Method{
						"Method1": {
							Rules: guard.Rules{
								{RequireAuthentication: guard.Ptr(true)},
							},
						},
					},
				},
			},
			want: []*guard.Service{
				{
					Name: "Service1",
					Rules: guard.Rules{
						{AllowPublic: guard.Ptr(true)},
					},
					Methods: map[string]*guard.Method{
						"Method1": {
							Rules: guard.Rules{
								{RequireAuthentication: guard.Ptr(true)},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple services with different rules",
			services: []*guard.Service{
				{
					Name: "Service1",
					Rules: guard.Rules{
						{AllowPublic: guard.Ptr(true)},
					},
					Methods: map[string]*guard.Method{},
				},
				{
					Name:  "Service2",
					Rules: nil,
					Methods: map[string]*guard.Method{
						"Method1": {
							Rules: guard.Rules{
								{RequireAuthentication: guard.Ptr(true)},
							},
						},
					},
				},
				{
					Name: "Service3",
					Rules: guard.Rules{
						{RequireAuthentication: guard.Ptr(true)},
					},
					Methods: map[string]*guard.Method{
						"Method1": {
							Rules: guard.Rules{
								{AllowPublic: guard.Ptr(true)},
							},
						},
					},
				},
				{
					Name:    "Service4",
					Rules:   nil,
					Methods: map[string]*guard.Method{},
				},
			},
			want: []*guard.Service{
				{
					Name: "Service1",
					Rules: guard.Rules{
						{AllowPublic: guard.Ptr(true)},
					},
					Methods: nil,
				},
				{
					Name:  "Service2",
					Rules: nil,
					Methods: map[string]*guard.Method{
						"Method1": {
							Rules: guard.Rules{
								{RequireAuthentication: guard.Ptr(true)},
							},
						},
					},
				},
				{
					Name: "Service3",
					Rules: guard.Rules{
						{RequireAuthentication: guard.Ptr(true)},
					},
					Methods: map[string]*guard.Method{
						"Method1": {
							Rules: guard.Rules{
								{AllowPublic: guard.Ptr(true)},
							},
						},
					},
				},
			},
		},
		{
			name: "service with authenticated access rules",
			services: []*guard.Service{
				{
					Name: "Service1",
					Rules: guard.Rules{
						{
							AuthenticatedAccess: &guard.AuthenticatedAccess{
								RoleBased: &guard.RoleBased{
									Roles:       []string{"role1", "role2"},
									Requirement: guard.RequirementAll,
								},
							},
						},
					},
					Methods: map[string]*guard.Method{},
				},
			},
			want: []*guard.Service{
				{
					Name: "Service1",
					Rules: guard.Rules{
						{
							AuthenticatedAccess: &guard.AuthenticatedAccess{
								RoleBased: &guard.RoleBased{
									Roles:       []string{"role1", "role2"},
									Requirement: guard.RequirementAll,
								},
							},
						},
					},
					Methods: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			serviceDescs := testCreateServiceDescriptors(tt.services)
			got := collectServices(serviceDescs)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

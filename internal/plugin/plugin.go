package plugin

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/casnerano/protoc-gen-go-rbac/pkg/rbac"
	desc "github.com/casnerano/protoc-gen-go-rbac/proto"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	outputFileSuffix = ".rbac.go"
)

//go:embed plugin.go.tmpl
var templateFS embed.FS

type TemplateData struct {
	Meta     Meta
	File     File
	Services []*rbac.Service

	Test string
}

type Meta struct {
	ModuleVersion string
	ProtocVersion string
}

type File struct {
	Name    string
	Package string
	Source  string
}

func Execute(plugin *protogen.Plugin) error {
	tmpl, err := parseTemplate()
	if err != nil {
		return err
	}

	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		services := collectServices(file.Services)
		if len(services) == 0 {
			continue
		}

		templateData := TemplateData{
			Meta: Meta{
				ProtocVersion: func() string {
					if ver := plugin.Request.CompilerVersion; ver != nil {
						return fmt.Sprintf("v%d.%d.%d", ver.Major, ver.Minor, ver.Patch)
					}
					return "(unknown)"
				}(),
				ModuleVersion: "(unknown)",
			},
			File: File{
				Name:    filepath.Base(file.GeneratedFilenamePrefix),
				Package: string(file.GoPackageName),
				Source:  file.Desc.Path(),
			},
			Services: services,
			Test:     fmt.Sprintf("%#v", services),
		}

		filename := file.GeneratedFilenamePrefix + outputFileSuffix
		if err = tmpl.Execute(plugin.NewGeneratedFile(filename, file.GoImportPath), templateData); err != nil {
			return fmt.Errorf("failed execute template: %w", err)
		}
	}

	return nil
}

func collectServices(protoServices []*protogen.Service) []*rbac.Service {
	var services []*rbac.Service
	for _, protoService := range protoServices {
		service := rbac.Service{
			Name: string(protoService.Desc.Name()),
		}

		if options := protoService.Desc.Options().(*descriptorpb.ServiceOptions); options != nil {
			if protoServiceRules := proto.GetExtension(options, desc.E_ServiceRules).(*desc.Rules); protoServiceRules != nil {
				service.Rules = extractSelectedRules(protoServiceRules)
			}
		}

		if methods := collectMethods(protoService.Methods); len(methods) > 0 {
			service.Methods = methods
		}

		if service.Rules == nil && service.Methods == nil {
			continue
		}

		services = append(services, &service)
	}

	return services
}

func collectMethods(protoMethods []*protogen.Method) map[string]*rbac.Method {
	methods := make(map[string]*rbac.Method)

	for _, protoMethod := range protoMethods {
		if options := protoMethod.Desc.Options().(*descriptorpb.MethodOptions); options != nil {
			if protoMethodRules := proto.GetExtension(options, desc.E_MethodRules).(*desc.Rules); protoMethodRules != nil {
				methods[string(protoMethod.Desc.Name())] = &rbac.Method{
					Rules: extractSelectedRules(protoMethodRules),
				}
			}
		}
	}

	return methods
}

func extractSelectedRules(descRules *desc.Rules) *rbac.Rules {
	rbacRules := rbac.Rules{}

	switch selectedRules := descRules.OneOf.(type) {
	case *desc.Rules_AllowPublic:
		rbacRules.AllowPublic = &selectedRules.AllowPublic
	case *desc.Rules_RequireAuthentication:
		rbacRules.RequireAuthn = &selectedRules.RequireAuthentication
	case *desc.Rules_RoleBased:
		if selectedRules.RoleBased != nil {
			rbacRules.RoleBased = &rbac.RoleBased{
				AllowedRoles: selectedRules.RoleBased.AllowedRoles,
				Requirement:  rbac.RequirementAny,
			}

			if requirement := selectedRules.RoleBased.Requirement; requirement != nil {
				rbacRules.RoleBased.Requirement = rbac.Requirement(*requirement)
			}
		}
	case *desc.Rules_PolicyBased:
		if selectedRules.PolicyBased != nil {
			rbacRules.PolicyBased = &rbac.PolicyBased{
				PolicyNames: selectedRules.PolicyBased.PolicyNames,
				Requirement: rbac.RequirementAll,
			}

			if requirement := selectedRules.PolicyBased.Requirement; requirement != nil {
				rbacRules.PolicyBased.Requirement = rbac.Requirement(*requirement)
			}
		}
	}

	return &rbacRules
}

func parseTemplate() (*template.Template, error) {
	templateContent, err := templateFS.ReadFile("plugin.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl := template.New("plugin.rbac").
		Funcs(template.FuncMap{
			"toLower": strings.ToLower,
		})

	return tmpl.Parse(string(templateContent))
}

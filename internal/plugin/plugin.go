package plugin

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/casnerano/protoc-gen-go-guard/pkg/guard"
	desc "github.com/casnerano/protoc-gen-go-guard/proto"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	outputFileSuffix = ".guard.go"
)

//go:embed plugin.go.tmpl
var templateFS embed.FS

type TemplateData struct {
	Meta     Meta
	File     File
	Services []*guard.Service

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

func collectServices(protoServices []*protogen.Service) []*guard.Service {
	var services []*guard.Service
	for _, protoService := range protoServices {
		service := guard.Service{
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

func collectMethods(protoMethods []*protogen.Method) map[string]*guard.Method {
	methods := make(map[string]*guard.Method)

	for _, protoMethod := range protoMethods {
		if options := protoMethod.Desc.Options().(*descriptorpb.MethodOptions); options != nil {
			if protoMethodRules := proto.GetExtension(options, desc.E_MethodRules).(*desc.Rules); protoMethodRules != nil {
				methods[string(protoMethod.Desc.Name())] = &guard.Method{
					Rules: extractSelectedRules(protoMethodRules),
				}
			}
		}
	}

	return methods
}

func extractSelectedRules(descRules *desc.Rules) *guard.Rules {
	guardRules := guard.Rules{}

	switch selectedRules := descRules.OneOf.(type) {
	case *desc.Rules_AllowPublic:
		guardRules.AllowPublic = &selectedRules.AllowPublic
	case *desc.Rules_RequireAuthentication:
		guardRules.RequireAuthn = &selectedRules.RequireAuthentication
	case *desc.Rules_RoleBased:
		if selectedRules.RoleBased != nil {
			guardRules.RoleBased = &guard.RoleBased{
				AllowedRoles: selectedRules.RoleBased.AllowedRoles,
				Requirement:  guard.RequirementAny,
			}

			if requirement := selectedRules.RoleBased.Requirement; requirement != nil {
				guardRules.RoleBased.Requirement = guard.Requirement(*requirement)
			}
		}
	case *desc.Rules_PolicyBased:
		if selectedRules.PolicyBased != nil {
			guardRules.PolicyBased = &guard.PolicyBased{
				PolicyNames: selectedRules.PolicyBased.PolicyNames,
				Requirement: guard.RequirementAll,
			}

			if requirement := selectedRules.PolicyBased.Requirement; requirement != nil {
				guardRules.PolicyBased.Requirement = guard.Requirement(*requirement)
			}
		}
	}

	return &guardRules
}

func parseTemplate() (*template.Template, error) {
	templateContent, err := templateFS.ReadFile("plugin.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl := template.New("plugin.guard").
		Funcs(template.FuncMap{
			"toLower": strings.ToLower,
		})

	return tmpl.Parse(string(templateContent))
}

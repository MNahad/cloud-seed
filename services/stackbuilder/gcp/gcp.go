package gcp

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
	"github.com/mnahad/cloud-seed/services/config/module"
)

type GcpStackConfig struct {
	Environment                 *string
	Project                     *string
	Region                      *string
	Dir                         *string
	OutDir                      *string
	RuntimeEnvironmentVariables *map[string]string
	SecretVariableNames         *[]string
	Manifests                   *[]module.Manifest
}

func NewGcpStack(scope *cdktf.App, id string, config GcpStackConfig) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(*scope, &id)
	google.NewGoogleProvider(stack, jsii.String("Google"), &google.GoogleProviderConfig{
		Zone:    config.Region,
		Project: config.Project,
	})
	var functions []*module.Module
	predicates := []func(*module.Module) bool{func(m *module.Module) bool {
		return m.Service.Function.Gcp != (module.Service{}.Function.Gcp)
	}}
	for i := range *config.Manifests {
		functionModules := (*config.Manifests)[i].FilterModules(predicates)[0]
		functions = append(functions, functionModules...)
	}
	for f := range functions {
		NewFunction(&stack, functions[f])
	}
	return stack
}

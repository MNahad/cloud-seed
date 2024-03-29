package stackbuilder

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-null-go/null/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp"
)

type StackConfig struct {
	Environment *string
	Config      *project.Config
	Manifests   []module.Manifest
}

func NewStack(scope *cdktf.App, id string, config *StackConfig) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(*scope, &id)
	null.NewNullProvider(stack, jsii.String("Null"), &null.NullProviderConfig{})
	gcp.NewGcpStack(&stack, id, &gcp.GcpStackConfig{
		Environment: config.Environment,
		Options:     config.Config,
		Manifests:   config.Manifests,
	})
	if config.Config.TfConfig.Backend.Gcs != (cdktf.GcsBackendProps{}) {
		cdktf.NewGcsBackend(stack, &config.Config.TfConfig.Backend.Gcs)
	} else if config.Config.TfConfig.Backend.S3 != (cdktf.S3BackendProps{}) {
		cdktf.NewS3Backend(stack, &config.Config.TfConfig.Backend.S3)
	} else {
		cdktf.NewLocalBackend(stack, &config.Config.TfConfig.Backend.Local)
	}
	return stack
}

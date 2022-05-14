package stackbuilder

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp"
)

func NewStack(
	scope *cdktf.App,
	id string,
	environment *string,
	manifests *[]module.Manifest,
	config *project.Config,
) cdktf.TerraformStack {
	stack := gcp.NewGcpStack(scope, id, gcp.GcpStackConfig{
		Environment: environment,
		Options:     config,
		Manifests:   manifests,
	})
	if config.TfConfig.Backend.Gcs != (cdktf.GcsBackendProps{}) {
		cdktf.NewGcsBackend(stack, &config.TfConfig.Backend.Gcs)
	} else if config.TfConfig.Backend.S3 != (cdktf.S3BackendProps{}) {
		cdktf.NewS3Backend(stack, &config.TfConfig.Backend.S3)
	} else {
		cdktf.NewLocalBackend(stack, &config.TfConfig.Backend.Local)
	}
	return stack
}

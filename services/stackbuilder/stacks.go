package stackbuilder

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp"
)

func NewStack(scope *cdktf.App, id string, environment *string, config *project.Config) cdktf.TerraformStack {
	stack := gcp.NewGcpStack(scope, id, gcp.GcpStackConfig{
		Project:                     &config.Cloud.Gcp.Project,
		Region:                      &config.Cloud.Gcp.Region,
		Dir:                         &config.BuildConfig.Dir,
		OutDir:                      &config.BuildConfig.OutDir,
		Environment:                 environment,
		RuntimeEnvironmentVariables: &config.RuntimeEnvironmentVariables,
		SecretVariableNames:         &config.SecretVariableNames,
	})
	switch config.TfConfig.Backend.Type {
	case "gcs":
		cdktf.NewGcsBackend(stack, &config.TfConfig.Backend.Gcs)
	case "s3":
		cdktf.NewS3Backend(stack, &config.TfConfig.Backend.S3)
	case "local":
	default:
		cdktf.NewLocalBackend(stack, &config.TfConfig.Backend.Local)
	}
	return stack
}

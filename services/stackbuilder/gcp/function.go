package gcp

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
	"github.com/mnahad/cloud-seed/services/config/module"
)

func NewFunction(scope *cdktf.TerraformStack, config *module.Module) {
	google.NewCloudfunctionsFunction(*scope, &config.Name, &config.Service.Function.Gcp.Config)
}

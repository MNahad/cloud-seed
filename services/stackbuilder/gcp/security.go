package gcp

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
	"github.com/mnahad/cloud-seed/generated/google_beta"
	"github.com/mnahad/cloud-seed/services/config/module"
)

func newAllUsersCloudFunctionInvoker(
	scope *cdktf.TerraformStack,
	function *google_beta.GoogleCloudfunctions2Function,
	module *module.Module,
) *google.CloudRunServiceIamMember {
	iamMember := google.NewCloudRunServiceIamMember(
		*scope,
		jsii.String(module.Name+"AllUsersInvoker"),
		&google.CloudRunServiceIamMemberConfig{
			Service: (*function).Name(),
			Member:  jsii.String("allUsers"),
			Role:    jsii.String("roles/run.invoker"),
		})
	return &iamMember
}

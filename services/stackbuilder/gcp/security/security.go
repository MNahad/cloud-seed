package security

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
)

func NewAllUsersCloudFunctionInvoker(
	scope *cdktf.TerraformStack,
	function *google.Cloudfunctions2Function,
	module *module.Module,
) *google.CloudRunServiceIamMember {
	iamMember := google.NewCloudRunServiceIamMember(
		*scope,
		jsii.String(module.Name+"AllUsersInvoker"),
		&google.CloudRunServiceIamMemberConfig{
			Service: (*function).ServiceConfig().Service(),
			Member:  jsii.String("allUsers"),
			Role:    jsii.String("roles/run.invoker"),
		})
	return &iamMember
}

func NewServiceAccountCloudFunctionInvoker(
	scope *cdktf.TerraformStack,
	function *google.Cloudfunctions2Function,
	serviceAccountName *string,
	serviceAccountEmail *string,
	module *module.Module,
) *google.CloudRunServiceIamMember {
	iamMember := google.NewCloudRunServiceIamMember(
		*scope,
		jsii.String(module.Name+*serviceAccountName+"Invoker"),
		&google.CloudRunServiceIamMemberConfig{
			Service: (*function).ServiceConfig().Service(),
			Member:  jsii.String("serviceAccount:" + *serviceAccountEmail),
			Role:    jsii.String("roles/run.invoker"),
		})
	return &iamMember
}

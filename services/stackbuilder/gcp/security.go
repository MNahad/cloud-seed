package gcp

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
)

func newAllUsersCloudFunctionInvoker(
	scope *cdktf.TerraformStack,
	functionName string,
) *google.CloudfunctionsFunctionIamMember {
	iamMember := google.NewCloudfunctionsFunctionIamMember(
		*scope,
		jsii.String(functionName+"AllUsersInvoker"),
		&google.CloudfunctionsFunctionIamMemberConfig{
			CloudFunction: &functionName,
			Member:        jsii.String("allUsers"),
			Role:          jsii.String("roles/cloudfunctions.invoker"),
		})
	return &iamMember
}

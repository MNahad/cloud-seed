package stacks

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
)

type GcpStackOptions struct {
	Project string
	Region  string
}

func NewGcpStack(scope constructs.Construct, id string, options GcpStackOptions) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, &id)
	google.NewGoogleProvider(stack, jsii.String("Google"), &google.GoogleProviderConfig{
		Zone:    jsii.String(options.Region),
		Project: jsii.String(options.Project),
	})

	return stack
}

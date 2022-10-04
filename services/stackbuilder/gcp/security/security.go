package security

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

var runtimeServiceAccount *google.ServiceAccount
var computeDefaultServiceAccount *google.DataGoogleComputeDefaultServiceAccount
var secrets = make(map[string]*google.SecretManagerSecret)

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

func GenerateRuntimeServiceAccount(scope *cdktf.TerraformStack, options *project.Config) *string {
	if options.Cloud.Gcp.Security.RuntimeServiceAccount != (google.ServiceAccountConfig{}) {
		serviceAccount := google.NewServiceAccount(
			*scope,
			options.Cloud.Gcp.Security.RuntimeServiceAccount.AccountId,
			&options.Cloud.Gcp.Security.RuntimeServiceAccount,
		)
		runtimeServiceAccount = &serviceAccount
		return serviceAccount.Email()
	} else {
		serviceAccount := google.NewDataGoogleComputeDefaultServiceAccount(
			*scope,
			jsii.String("ComputeDefaultServiceAccount"),
			&google.DataGoogleComputeDefaultServiceAccountConfig{},
		)
		computeDefaultServiceAccount = &serviceAccount
		return serviceAccount.Email()
	}
}

func NewSecretManagerSecret(
	scope *cdktf.TerraformStack,
	name *string,
	options *project.Config,
) *google.SecretManagerSecret {
	if secrets[*name] == nil {
		secretConfig := new(google.SecretManagerSecretConfig)
		(*secretConfig) = options.Cloud.Gcp.Security.SecretManagerSecret
		if secretConfig.SecretId == nil {
			secretConfig.SecretId = name
		}
		if secretConfig.Replication == nil {
			secretConfig.Replication = &google.SecretManagerSecretReplication{
				UserManaged: &google.SecretManagerSecretReplicationUserManaged{
					Replicas: []google.SecretManagerSecretReplicationUserManagedReplicas{
						{Location: options.Cloud.Gcp.Provider.Region},
					},
				},
			}
		}
		secret := google.NewSecretManagerSecret(*scope, secretConfig.SecretId, secretConfig)
		secrets[*name] = &secret
	}
	return secrets[*name]
}

func NewServiceAccountSecretManagerSecretAccessor(
	scope *cdktf.TerraformStack,
	secretId *string,
	secret *google.SecretManagerSecret,
	serviceAccountEmail *string,
) *google.SecretManagerSecretIamMember {
	secretIamMember := google.NewSecretManagerSecretIamMember(
		*scope,
		jsii.String(*secretId+"ServiceAccountSecretAccessor"),
		&google.SecretManagerSecretIamMemberConfig{
			SecretId: (*secret).SecretId(),
			Member:   jsii.String("serviceAccount:" + *serviceAccountEmail),
			Role:     jsii.String("roles/secretmanager.secretAccessor"),
		},
	)
	return &secretIamMember
}
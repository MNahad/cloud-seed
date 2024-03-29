package security

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

type security struct {
	secrets map[string]*google.SecretManagerSecret
}

func NewSecurity() *security {
	return &security{
		secrets: make(map[string]*google.SecretManagerSecret),
	}
}

func (s *security) NewAllUsersCloudFunctionInvoker(
	scope *cdktf.TerraformStack,
	function *google.Cloudfunctions2Function,
	module *module.Module,
	options *project.Config,
) *google.CloudRunServiceIamMember {
	return s.NewAllUsersCloudRunInvoker(scope, (*function).ServiceConfig().Service(), module, options)
}

func (s *security) NewServiceAccountCloudFunctionInvoker(
	scope *cdktf.TerraformStack,
	function *google.Cloudfunctions2Function,
	serviceAccountName *string,
	serviceAccountEmail *string,
	module *module.Module,
	options *project.Config,
) *google.CloudRunServiceIamMember {
	return s.NewServiceAccountCloudRunInvoker(
		scope,
		(*function).ServiceConfig().Service(),
		serviceAccountName,
		serviceAccountEmail,
		module,
		options,
	)
}

func (s *security) NewServiceAccountCloudRunInvoker(
	scope *cdktf.TerraformStack,
	service *string,
	serviceAccountName *string,
	serviceAccountEmail *string,
	module *module.Module,
	options *project.Config,
) *google.CloudRunServiceIamMember {
	iamMember := google.NewCloudRunServiceIamMember(
		*scope,
		jsii.String(module.Name+*serviceAccountName+"Invoker"),
		&google.CloudRunServiceIamMemberConfig{
			Service:  service,
			Member:   jsii.String("serviceAccount:" + *serviceAccountEmail),
			Role:     jsii.String("roles/run.invoker"),
			Location: options.Cloud.Gcp.Provider.Region,
		})
	return &iamMember
}

func (s *security) NewAllUsersCloudRunInvoker(
	scope *cdktf.TerraformStack,
	service *string,
	module *module.Module,
	options *project.Config,
) *google.CloudRunServiceIamMember {
	iamMember := google.NewCloudRunServiceIamMember(
		*scope,
		jsii.String(module.Name+"AllUsersInvoker"),
		&google.CloudRunServiceIamMemberConfig{
			Service:  service,
			Member:   jsii.String("allUsers"),
			Role:     jsii.String("roles/run.invoker"),
			Location: options.Cloud.Gcp.Provider.Region,
		})
	return &iamMember
}

func (s *security) NewRuntimeServiceAccount(scope *cdktf.TerraformStack, options *project.Config) *google.ServiceAccount {
	serviceAccountConfig := new(google.ServiceAccountConfig)
	(*serviceAccountConfig) = options.Cloud.Gcp.Security.RuntimeServiceAccount
	if serviceAccountConfig.AccountId == nil {
		serviceAccountConfig.AccountId = jsii.String("runtime")
	}
	serviceAccount := google.NewServiceAccount(*scope, serviceAccountConfig.AccountId, serviceAccountConfig)
	return &serviceAccount
}

func (s *security) NewSecretManagerSecret(
	scope *cdktf.TerraformStack,
	name *string,
	options *project.Config,
) *google.SecretManagerSecret {
	if s.secrets[*name] == nil {
		secretConfig := new(google.SecretManagerSecretConfig)
		(*secretConfig) = options.Cloud.Gcp.Security.SecretManagerSecret
		if secretConfig.SecretId == nil {
			secretConfig.SecretId = name
		}
		if secretConfig.Replication == nil {
			secretConfig.Replication = &google.SecretManagerSecretReplication{
				UserManaged: &google.SecretManagerSecretReplicationUserManaged{
					Replicas: &[]google.SecretManagerSecretReplicationUserManagedReplicas{
						{Location: options.Cloud.Gcp.Provider.Region},
					},
				},
			}
		}
		secret := google.NewSecretManagerSecret(*scope, secretConfig.SecretId, secretConfig)
		s.secrets[*name] = &secret
	}
	return s.secrets[*name]
}

func (s *security) NewServiceAccountSecretManagerSecretAccessor(
	scope *cdktf.TerraformStack,
	secretId *string,
	secret *google.SecretManagerSecret,
	serviceAccountName *string,
	serviceAccountEmail *string,
) *google.SecretManagerSecretIamMember {
	secretIamMember := google.NewSecretManagerSecretIamMember(
		*scope,
		jsii.String(*secretId+*serviceAccountName+"SecretAccessor"),
		&google.SecretManagerSecretIamMemberConfig{
			SecretId: (*secret).SecretId(),
			Member:   jsii.String("serviceAccount:" + *serviceAccountEmail),
			Role:     jsii.String("roles/secretmanager.secretAccessor"),
		},
	)
	return &secretIamMember
}

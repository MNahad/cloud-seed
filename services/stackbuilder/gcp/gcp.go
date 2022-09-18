package gcp

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/eventsource"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/function"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/networking"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/orchestration"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/security"
)

type modulesCollection struct {
	function  []*module.Module
	container []*module.Module
	workflow  []*module.Module
}

type GcpStackConfig struct {
	Environment *string
	Options     *project.Config
	Manifests   []module.Manifest
}

func NewGcpStack(scope *cdktf.App, id string, config *GcpStackConfig) cdktf.TerraformStack {
	support := supportInfrastructure{}
	stack := cdktf.NewTerraformStack(*scope, &id)
	google.NewGoogleProvider(stack, jsii.String("Google"), &config.Options.Cloud.Gcp.Provider)
	if len(config.Manifests) > 0 {
		support.generateInfrastructure(&stack, kindCommon, config.Options)
	}
	var modules modulesCollection
	predicates := []func(*module.Module) bool{
		func(m *module.Module) bool {
			return m.Service.Function.Gcp != (module.Service{}.Function.Gcp)
		},
		func(m *module.Module) bool {
			return orchestration.IsWorkflow(&m.Orchestration)
		},
	}
	for i := range config.Manifests {
		manifest := &(config.Manifests)[i]
		results := manifest.FilterModules(predicates)
		functionModules := results[0]
		workflowModules := results[1]
		modules.function = append(modules.function, functionModules...)
		modules.workflow = append(modules.workflow, workflowModules...)
	}
	functionModules := modules.function
	if support.function == (supportInfrastructure{}.function) && len(functionModules) > 0 {
		support.generateInfrastructure(&stack, kindFunction, config.Options)
	}
	serviceEndpoints := make(orchestration.ServiceEndpoints, len(functionModules))
	for i := range functionModules {
		functionModule := functionModules[i]
		functionConstruct := *function.NewFunction(
			&stack,
			functionModule,
			support.function.archiveBucket,
			support.getRuntimeServiceAccountEmail(),
			config.Options,
		)
		security.NewServiceAccountCloudFunctionInvoker(
			&stack,
			&functionConstruct,
			jsii.String("RuntimeServiceAccount"),
			support.getRuntimeServiceAccountEmail(),
			functionModule,
		)
		if functionModule.EventSource.EventSpec != (module.EventSource{}.EventSpec) {
			eventTopic := *eventsource.NewTopicEventSource(&stack, &functionModule.EventSource, config.Options)
			if functionConstruct.EventTriggerInput() == nil {
				functionConstruct.PutEventTrigger(&google.Cloudfunctions2FunctionEventTrigger{})
			}
			eventTrigger := functionConstruct.EventTriggerInput()
			if eventTrigger.PubsubTopic == nil {
				eventTrigger.PubsubTopic = eventTopic.Id()
			}
			if eventTrigger.EventType == nil {
				eventTrigger.EventType = jsii.String("google.cloud.pubsub.topic.v1.messagePublished")
			}
			if eventTrigger.RetryPolicy == nil {
				eventTrigger.RetryPolicy = jsii.String("RETRY_POLICY_DO_NOT_RETRY")
			}
			if eventTrigger.TriggerRegion == nil {
				eventTrigger.TriggerRegion = config.Options.Cloud.Gcp.Provider.Region
			}
			if eventTrigger.ServiceAccountEmail == nil {
				eventTrigger.ServiceAccountEmail = support.getRuntimeServiceAccountEmail()
			}
			functionConstruct.PutEventTrigger(eventTrigger)
		} else if functionModule.EventSource.QueueSpec != (module.EventSource{}.QueueSpec) {
			eventsource.NewQueueEventSource(&stack, &functionModule.EventSource, config.Options)
		} else if functionModule.EventSource.ScheduleSpec != (module.EventSource{}.ScheduleSpec) {
			schedule := *eventsource.NewScheduleEventSource(&stack, &functionModule.EventSource, config.Options)
			if schedule.HttpTargetInput() == nil {
				schedule.PutHttpTarget(
					&google.CloudSchedulerJobHttpTarget{Uri: jsii.String(eventsource.PlaceholderHttpTargetUri)},
				)
			}
			if httpTarget := schedule.HttpTargetInput(); httpTarget != nil {
				if httpTarget.Uri == nil || *httpTarget.Uri == eventsource.PlaceholderHttpTargetUri {
					httpTarget.Uri = functionConstruct.ServiceConfig().Uri()
				}
				if httpTarget.HttpMethod == nil {
					httpTarget.HttpMethod = jsii.String("POST")
				}
				if httpTarget.OidcToken == nil {
					httpTarget.OidcToken = &google.CloudSchedulerJobHttpTargetOidcToken{
						ServiceAccountEmail: support.getRuntimeServiceAccountEmail(),
						Audience:            functionConstruct.ServiceConfig().Uri(),
					}
				}
				schedule.PutHttpTarget(httpTarget)
			}
		}
		if functionModule.Security.NoAuthentication {
			security.NewAllUsersCloudFunctionInvoker(&stack, &functionConstruct, functionModule)
		}
		if functionModule.Networking.Internal {
			serviceConfig := functionConstruct.ServiceConfigInput()
			if serviceConfig == nil {
				serviceConfig = &google.Cloudfunctions2FunctionServiceConfig{}
			}
			if serviceConfig.IngressSettings == nil {
				serviceConfig.IngressSettings = jsii.String("ALLOW_INTERNAL_ONLY")
				functionConstruct.PutServiceConfig(serviceConfig)
			}
		}
		if functionModule.Networking.Egress.StaticIp {
			connector := networking.GenerateVpcAccessConnector(&stack, config.Options)
			serviceConfig := functionConstruct.ServiceConfigInput()
			if serviceConfig == nil {
				serviceConfig = &google.Cloudfunctions2FunctionServiceConfig{}
			}
			if serviceConfig.VpcConnector == nil {
				serviceConfig.VpcConnector = (*connector).Id()
			}
			if serviceConfig.VpcConnectorEgressSettings == nil {
				serviceConfig.VpcConnectorEgressSettings = jsii.String("ALL_TRAFFIC")
			}
			functionConstruct.PutServiceConfig(serviceConfig)
		}
		serviceEndpoints[functionModule.Name] = orchestration.Endpoint{Uri: *functionConstruct.ServiceConfig().Uri()}
	}
	if len(config.Manifests) > 0 && len(config.Options.OrchestrationConfig.Gcp.FilePath) > 0 {
		orchestration.NewWorkflow(&stack, nil, nil, support.getRuntimeServiceAccountEmail(), config.Options)
	} else if workflowModules := modules.workflow; len(workflowModules) > 0 {
		orchestration.NewWorkflow(
			&stack,
			workflowModules,
			&serviceEndpoints,
			support.getRuntimeServiceAccountEmail(),
			config.Options,
		)
	}
	return stack
}

type resourceKind uint

const (
	kindCommon resourceKind = iota
	kindFunction
	kindContainer
)

type supportInfrastructure struct {
	common struct {
		runtimeServiceAccount        *google.ServiceAccount
		computeDefaultServiceAccount *google.DataGoogleComputeDefaultServiceAccount
		secrets                      []*google.SecretManagerSecret
		secretsIamMembers            []*google.SecretManagerSecretIamMember
	}
	function struct {
		archiveBucket *google.StorageBucket
	}
	container any
}

func (s *supportInfrastructure) generateInfrastructure(
	scope *cdktf.TerraformStack,
	kind resourceKind,
	options *project.Config,
) {
	switch kind {
	case kindCommon:
		{
			if options.Cloud.Gcp.Security.RuntimeServiceAccount != (google.ServiceAccountConfig{}) {
				serviceAccount := google.NewServiceAccount(
					*scope,
					options.Cloud.Gcp.Security.RuntimeServiceAccount.AccountId,
					&options.Cloud.Gcp.Security.RuntimeServiceAccount,
				)
				s.common.runtimeServiceAccount = &serviceAccount
			} else {
				computeDefaultServiceAccount := google.NewDataGoogleComputeDefaultServiceAccount(
					*scope,
					jsii.String("ComputeDefaultServiceAccount"),
					&google.DataGoogleComputeDefaultServiceAccountConfig{},
				)
				s.common.computeDefaultServiceAccount = &computeDefaultServiceAccount
			}
			s.common.secrets = make(
				[]*google.SecretManagerSecret,
				len(options.EnvironmentConfig.SecretVariableNames),
			)
			s.common.secretsIamMembers = make(
				[]*google.SecretManagerSecretIamMember,
				len(options.EnvironmentConfig.SecretVariableNames),
			)
			for i := range options.EnvironmentConfig.SecretVariableNames {
				secretConfig := new(google.SecretManagerSecretConfig)
				(*secretConfig) = options.Cloud.Gcp.SecretsManagement.Secrets
				if secretConfig.SecretId == nil {
					secretConfig.SecretId = jsii.String(options.EnvironmentConfig.SecretVariableNames[i])
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
				s.common.secrets = append(s.common.secrets, &secret)
				secretIamMember := google.NewSecretManagerSecretIamMember(
					*scope,
					jsii.String(*secretConfig.SecretId+"-iam-member"),
					&google.SecretManagerSecretIamMemberConfig{
						SecretId: secret.SecretId(),
						Member:   jsii.String("serviceAccount:" + *s.getRuntimeServiceAccountEmail()),
						Role:     jsii.String("roles/secretmanager.secretAccessor"),
					},
				)
				s.common.secretsIamMembers = append(s.common.secretsIamMembers, &secretIamMember)
			}
		}
	case kindFunction:
		{
			archiveBucketConfig := new(google.StorageBucketConfig)
			(*archiveBucketConfig) = options.Cloud.Gcp.SourceCodeStorage.Bucket
			if archiveBucketConfig.Name == nil {
				archiveBucketConfig.Name = jsii.String(*options.Cloud.Gcp.Provider.Project + "-functions-src")
			}
			if archiveBucketConfig.Location == nil {
				archiveBucketConfig.Location = options.Cloud.Gcp.Provider.Region
			}
			if archiveBucketConfig.UniformBucketLevelAccess == nil {
				archiveBucketConfig.UniformBucketLevelAccess = jsii.Bool(true)
			}
			archiveBucket := google.NewStorageBucket(*scope, jsii.String("ArchiveBucket"), archiveBucketConfig)
			s.function.archiveBucket = &archiveBucket
		}
	case kindContainer:
	}
}

func (s *supportInfrastructure) getRuntimeServiceAccountEmail() *string {
	if s.common.runtimeServiceAccount != nil {
		return (*s.common.runtimeServiceAccount).Email()
	} else if s.common.computeDefaultServiceAccount != nil {
		return (*s.common.computeDefaultServiceAccount).Email()
	} else {
		return nil
	}
}

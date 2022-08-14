package gcp

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
	"github.com/mnahad/cloud-seed/generated/google_beta"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

type GcpStackConfig struct {
	Environment *string
	Options     *project.Config
	Manifests   *[]module.Manifest
}

func NewGcpStack(scope *cdktf.App, id string, config GcpStackConfig) cdktf.TerraformStack {
	support := supportInfrastructure{}
	stack := cdktf.NewTerraformStack(*scope, &id)
	google.NewGoogleProvider(stack, jsii.String("Google"), &google.GoogleProviderConfig{
		Project: &config.Options.Cloud.Gcp.Project,
		Zone:    &config.Options.Cloud.Gcp.Region,
	})
	betaProvider := google_beta.NewGoogleBetaProvider(stack, jsii.String("GoogleBeta"), &google_beta.GoogleBetaProviderConfig{
		Project: &config.Options.Cloud.Gcp.Project,
		Zone:    &config.Options.Cloud.Gcp.Region,
	})
	if len(*config.Manifests) > 0 {
		support.generateInfrastructure(&stack, kindCommon, config.Options)
	}
	predicates := []func(*module.Module) bool{
		func(m *module.Module) bool {
			return m.Service.Function.Gcp != (module.Service{}.Function.Gcp)
		},
	}
	for i := range *config.Manifests {
		manifest := &(*config.Manifests)[i]
		functionModules := manifest.FilterModules(predicates)[0]
		if support.function == (supportInfrastructure{}.function) && len(functionModules) > 0 {
			support.generateInfrastructure(&stack, kindFunction, config.Options)
		}
		for j := range functionModules {
			functionModule := functionModules[j]
			function := *newFunction(&stack, functionModule, &support, manifest, config.Options, &betaProvider)
			if functionModule.EventSource.EventSpec != (module.EventSource{}.EventSpec) {
				eventTopic := *newTopicEventSource(&stack, &functionModule.EventSource)
				if function.EventTriggerInput() == nil {
					function.PutEventTrigger(&google_beta.GoogleCloudfunctions2FunctionEventTrigger{})
				}
				eventTrigger := function.EventTriggerInput()
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
					eventTrigger.TriggerRegion = &config.Options.Cloud.Gcp.Region
				}
				function.PutEventTrigger(eventTrigger)
			} else if functionModule.EventSource.QueueSpec != (module.EventSource{}.QueueSpec) {
				newQueueEventSource(&stack, &functionModule.EventSource, config.Options)
			} else if functionModule.EventSource.ScheduleSpec != (module.EventSource{}.ScheduleSpec) {
				schedule := *newScheduleEventSource(&stack, &functionModule.EventSource, config.Options)
				if schedule.HttpTargetInput() == nil {
					schedule.PutHttpTarget(
						&google.CloudSchedulerJobHttpTarget{Uri: jsii.String(cloudSchedulerPlaceholderHttpTargetUri)},
					)
				}
				if httpTarget := schedule.HttpTargetInput(); httpTarget != nil {
					if httpTarget.Uri == nil || *httpTarget.Uri == cloudSchedulerPlaceholderHttpTargetUri {
						httpTarget.Uri = function.ServiceConfig().Uri()
					}
					if httpTarget.HttpMethod == nil {
						httpTarget.HttpMethod = jsii.String("POST")
					}
					if httpTarget.OidcToken == nil {
						computeDefaultServiceAccount := google.NewDataGoogleComputeDefaultServiceAccount(
							stack,
							jsii.String("ComputeDefaultServiceAccount"),
							&google.DataGoogleComputeDefaultServiceAccountConfig{},
						)
						httpTarget.OidcToken = &google.CloudSchedulerJobHttpTargetOidcToken{
							ServiceAccountEmail: computeDefaultServiceAccount.Email(),
							Audience:            function.ServiceConfig().Uri(),
						}
					}
					schedule.PutHttpTarget(httpTarget)
				}
			}
			if functionModule.Security.NoAuthentication {
				newAllUsersCloudFunctionInvoker(&stack, &function, functionModule)
			}
			if functionModule.Networking.Internal {
				serviceConfig := function.ServiceConfigInput()
				if serviceConfig == nil {
					serviceConfig = &google_beta.GoogleCloudfunctions2FunctionServiceConfig{}
				}
				if serviceConfig.IngressSettings == nil {
					serviceConfig.IngressSettings = jsii.String("ALLOW_INTERNAL_ONLY")
					function.PutServiceConfig(serviceConfig)
				}
			}
		}
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
		secrets []*google.SecretManagerSecret
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
			secrets := make([]*google.SecretManagerSecret, len(options.EnvironmentConfig.SecretVariableNames))
			for i := range options.EnvironmentConfig.SecretVariableNames {
				name := &options.EnvironmentConfig.SecretVariableNames[i]
				secret := google.NewSecretManagerSecret(*scope, name, &google.SecretManagerSecretConfig{
					SecretId: name,
					Replication: &google.SecretManagerSecretReplication{
						Automatic: true,
					},
				})
				secrets[i] = &secret
			}
			s.common.secrets = append(s.common.secrets, secrets...)
		}
	case kindFunction:
		{
			archiveBucket := google.NewStorageBucket(*scope, jsii.String("ArchiveBucket"), &google.StorageBucketConfig{
				Name:     jsii.String(options.Cloud.Gcp.Project + "-functions"),
				Location: &options.Cloud.Gcp.Region,
			})
			s.function.archiveBucket = &archiveBucket
		}
	case kindContainer:
	}
}

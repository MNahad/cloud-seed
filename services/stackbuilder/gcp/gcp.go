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
	google_beta.NewGoogleBetaProvider(stack, jsii.String("GoogleBeta"), &google_beta.GoogleBetaProviderConfig{
		Project: &config.Options.Cloud.Gcp.Project,
		Zone:    &config.Options.Cloud.Gcp.Region,
	})
	if len(*config.Manifests) > 0 {
		support.generateInfrastructure(&stack, kindCommon, config.Options)
	}
	predicates := []func(*module.Module) bool{func(m *module.Module) bool {
		return m.Service.Function.Gcp != (module.Service{}.Function.Gcp)
	}}
	for i := range *config.Manifests {
		manifest := &(*config.Manifests)[i]
		functionModules := manifest.FilterModules(predicates)[0]
		if support.function == (supportInfrastructure{}.function) && len(functionModules) > 0 {
			support.generateInfrastructure(&stack, kindFunction, config.Options)
		}
		for j := range functionModules {
			functionModule := functionModules[j]
			function := *newFunction(&stack, functionModule, &support, manifest, config.Options)
			if functionModule.EventSource.EventSpec != (module.EventSource{}.EventSpec) {
				eventTopic := *newTopicEventSource(&stack, &functionModule.EventSource)
				if function.EventTriggerInput() == nil {
					function.PutEventTrigger(&google_beta.GoogleCloudfunctions2FunctionEventTrigger{})
				}
				if function.EventTriggerInput().PubsubTopic == nil {
					eventTrigger := function.EventTriggerInput()
					eventTrigger.PubsubTopic = eventTopic.Id()
					function.PutEventTrigger(eventTrigger)
				}
				if function.EventTriggerInput().EventType == nil {
					eventTrigger := function.EventTriggerInput()
					eventTrigger.EventType = jsii.String("google.cloud.pubsub.topic.v1.messagePublished")
					function.PutEventTrigger(eventTrigger)
				}
			} else if functionModule.EventSource.QueueSpec != (module.EventSource{}.QueueSpec) {
				newQueueEventSource(&stack, &functionModule.EventSource, config.Options)
			} else if functionModule.EventSource.ScheduleSpec != (module.EventSource{}.ScheduleSpec) {
				schedule := *newScheduleEventSource(&stack, &functionModule.EventSource)
				if schedule.HttpTargetInput() == nil {
					schedule.PutHttpTarget(&google.CloudSchedulerJobHttpTarget{Uri: function.ServiceConfig().GcfUri()})
				}
				if schedule.HttpTargetInput().Uri == nil {
					httpTarget := schedule.HttpTargetInput()
					httpTarget.Uri = function.ServiceConfig().GcfUri()
					schedule.PutHttpTarget(httpTarget)
				}
			}
			if !functionModule.Security.Authentication {
				newAllUsersCloudFunctionInvoker(&stack, functionModule.Name)
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
	case kindFunction:
		archiveBucket := google.NewStorageBucket(*scope, jsii.String("ArchiveBucket"), &google.StorageBucketConfig{
			Name:     jsii.String(options.Cloud.Gcp.Project + "-functions"),
			Location: &options.Cloud.Gcp.Region,
		})
		s.function.archiveBucket = &archiveBucket
	case kindContainer:
	}
}

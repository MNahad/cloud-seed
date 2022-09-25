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
	stack := cdktf.NewTerraformStack(*scope, &id)
	google.NewGoogleProvider(stack, jsii.String("Google"), &config.Options.Cloud.Gcp.Provider)
	var runtimeServiceAccountEmail *string
	if len(config.Manifests) > 0 {
		runtimeServiceAccountEmail = security.GenerateRuntimeServiceAccount(&stack, config.Options)
		for i := range config.Options.EnvironmentConfig.SecretVariableNames {
			secretId := config.Options.EnvironmentConfig.SecretVariableNames[i]
			secret := security.NewSecretManagerSecret(&stack, &secretId, config.Options)
			security.NewServiceAccountSecretManagerSecretAccessor(&stack, &secretId, secret, runtimeServiceAccountEmail)
		}
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
	serviceEndpoints := make(orchestration.ServiceEndpoints, len(functionModules))
	for i := range functionModules {
		functionModule := functionModules[i]
		functionConstruct := *function.NewFunction(
			&stack,
			functionModule,
			runtimeServiceAccountEmail,
			config.Options,
		)
		security.NewServiceAccountCloudFunctionInvoker(
			&stack,
			&functionConstruct,
			jsii.String("RuntimeServiceAccount"),
			runtimeServiceAccountEmail,
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
				eventTrigger.ServiceAccountEmail = runtimeServiceAccountEmail
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
						ServiceAccountEmail: runtimeServiceAccountEmail,
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
			connector := networking.NewVpcAccessConnector(&stack, config.Options)
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
		orchestration.NewWorkflow(&stack, nil, nil, runtimeServiceAccountEmail, config.Options)
	} else if workflowModules := modules.workflow; len(workflowModules) > 0 {
		orchestration.NewWorkflow(
			&stack,
			workflowModules,
			&serviceEndpoints,
			runtimeServiceAccountEmail,
			config.Options,
		)
	}
	return stack
}

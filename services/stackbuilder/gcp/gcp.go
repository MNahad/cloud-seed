package gcp

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/cdktf-provider-googlebeta-go/googlebeta/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/eventsource"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/networking"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/orchestration"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/security"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/service"
	serviceendpoint "github.com/mnahad/cloud-seed/services/stackbuilder/gcp/service/endpoint"
)

type modulesCollection struct {
	function  []*module.Module
	container []*module.Module
	gateway   []*module.Module
	workflow  []*module.Module
}

type GcpStackConfig struct {
	Environment *string
	Options     *project.Config
	Manifests   []module.Manifest
}

func NewGcpStack(stack *cdktf.TerraformStack, id string, config *GcpStackConfig) cdktf.TerraformStack {
	eventsource := eventsource.NewEventSource()
	networking := networking.NewNetworking()
	security := security.NewSecurity()
	service := service.NewService()
	google.NewGoogleProvider(*stack, jsii.String("Google"), &config.Options.Cloud.Gcp.Provider)
	betaProvider := googlebeta.NewGoogleBetaProvider(
		*stack,
		jsii.String("GoogleBeta"),
		&config.Options.Cloud.Gcp.BetaProvider,
	)
	var runtimeServiceAccount *google.ServiceAccount
	if len(config.Manifests) > 0 {
		runtimeServiceAccount = security.NewRuntimeServiceAccount(stack, config.Options)
		for i := range config.Options.EnvironmentConfig.SecretVariableNames {
			secretId := config.Options.EnvironmentConfig.SecretVariableNames[i]
			secret := security.NewSecretManagerSecret(stack, &secretId, config.Options)
			security.NewServiceAccountSecretManagerSecretAccessor(
				stack,
				&secretId,
				secret,
				jsii.String("RuntimeServiceAccount"),
				(*runtimeServiceAccount).Email(),
			)
		}
	}
	var modules modulesCollection
	predicates := []func(*module.Module) bool{
		func(m *module.Module) bool {
			return m.Service.Function.Gcp != (module.Service{}.Function.Gcp)
		},
		func(m *module.Module) bool {
			return m.Service.Container.Gcp != (module.Service{}.Container.Gcp)
		},
		func(m *module.Module) bool {
			return networking.IsGateway(&m.Networking)
		},
		func(m *module.Module) bool {
			return orchestration.IsWorkflow(&m.Orchestration)
		},
	}
	for i := range config.Manifests {
		manifest := &(config.Manifests)[i]
		results := manifest.FilterModules(predicates)
		modules.function = append(modules.function, results[0]...)
		modules.container = append(modules.container, results[1]...)
		modules.gateway = append(modules.gateway, results[2]...)
		modules.workflow = append(modules.workflow, results[3]...)
	}
	functionModules := modules.function
	containerModules := modules.container
	serviceEndpoints := make(serviceendpoint.Endpoints, len(functionModules)+len(containerModules))
	for i := range functionModules {
		functionModule := functionModules[i]
		function := *service.NewFunction(
			stack,
			functionModule,
			config.Options,
		)
		if function.ServiceConfig().ServiceAccountEmail() == nil {
			serviceConfig := function.ServiceConfigInput()
			serviceConfig.ServiceAccountEmail = (*runtimeServiceAccount).Email()
			function.PutServiceConfig(serviceConfig)
		}
		security.NewServiceAccountCloudFunctionInvoker(
			stack,
			&function,
			jsii.String("RuntimeServiceAccount"),
			(*runtimeServiceAccount).Email(),
			functionModule,
			config.Options,
		)
		if functionModule.EventSource.EventSpec != (module.EventSource{}.EventSpec) {
			event := *eventsource.NewEventarcTrigger(stack, &functionModule.EventSource, config.Options)
			destination := event.DestinationInput()
			if destination.CloudFunction == nil {
				destination.CloudFunction = function.Id()
			}
			event.PutDestination(destination)
			serviceAccount := event.ServiceAccountInput()
			if serviceAccount == nil {
				serviceAccount = (*runtimeServiceAccount).Email()
			}
			event.SetServiceAccount(serviceAccount)
		} else if functionModule.EventSource.TopicSpec != (module.EventSource{}.TopicSpec) {
			topic := *eventsource.NewTopicEventSource(stack, &functionModule.EventSource, config.Options)
			eventTrigger := function.EventTriggerInput()
			if eventTrigger == nil {
				eventTrigger = &google.Cloudfunctions2FunctionEventTrigger{}
			}
			if eventTrigger.PubsubTopic == nil {
				eventTrigger.PubsubTopic = topic.Id()
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
				eventTrigger.ServiceAccountEmail = (*runtimeServiceAccount).Email()
			}
			function.PutEventTrigger(eventTrigger)
		} else if functionModule.EventSource.QueueSpec != (module.EventSource{}.QueueSpec) {
			eventsource.NewQueueEventSource(stack, &functionModule.EventSource, config.Options)
		} else if functionModule.EventSource.ScheduleSpec != (module.EventSource{}.ScheduleSpec) {
			schedule := *eventsource.NewScheduleEventSource(stack, &functionModule.EventSource, config.Options)
			httpTarget := schedule.HttpTargetInput()
			if httpTarget == nil {
				httpTarget = &google.CloudSchedulerJobHttpTarget{}
			}
			if httpTarget.Uri == nil {
				httpTarget.Uri = function.ServiceConfig().Uri()
			}
			if httpTarget.HttpMethod == nil {
				httpTarget.HttpMethod = jsii.String("POST")
			}
			if httpTarget.OidcToken == nil {
				httpTarget.OidcToken = &google.CloudSchedulerJobHttpTargetOidcToken{
					ServiceAccountEmail: (*runtimeServiceAccount).Email(),
					Audience:            function.ServiceConfig().Uri(),
				}
			}
			schedule.PutHttpTarget(httpTarget)
		}
		if functionModule.Security.NoAuthentication {
			security.NewAllUsersCloudFunctionInvoker(stack, &function, functionModule, config.Options)
		}
		if functionModule.Networking.Internal {
			serviceConfig := function.ServiceConfigInput()
			if serviceConfig == nil {
				serviceConfig = &google.Cloudfunctions2FunctionServiceConfig{}
			}
			if serviceConfig.IngressSettings == nil {
				serviceConfig.IngressSettings = jsii.String("ALLOW_INTERNAL_ONLY")
			}
			function.PutServiceConfig(serviceConfig)
		}
		if functionModule.Networking.Egress.StaticIp {
			connector := *networking.NewVpcAccessConnector(stack, config.Options)
			serviceConfig := function.ServiceConfigInput()
			if serviceConfig == nil {
				serviceConfig = &google.Cloudfunctions2FunctionServiceConfig{}
			}
			if serviceConfig.VpcConnector == nil {
				serviceConfig.VpcConnector = connector.Id()
			}
			if serviceConfig.VpcConnectorEgressSettings == nil {
				serviceConfig.VpcConnectorEgressSettings = jsii.String("ALL_TRAFFIC")
			}
			function.PutServiceConfig(serviceConfig)
		}
		serviceEndpoints[functionModule.Name] = serviceendpoint.Endpoint{Uri: *function.ServiceConfig().Uri()}
	}
	for i := range containerModules {
		containerModule := containerModules[i]
		runService := *service.NewRunService(stack, containerModule, config.Options)
		if runService.Template().Spec().ServiceAccountName() == nil {
			template := runService.TemplateInput()
			template.Spec.ServiceAccountName = (*runtimeServiceAccount).Email()
			runService.PutTemplate(template)
		}
		security.NewServiceAccountCloudRunInvoker(
			stack,
			runService.Name(),
			jsii.String("RuntimeServiceAccount"),
			(*runtimeServiceAccount).Email(),
			containerModule,
			config.Options,
		)
		if containerModule.EventSource.EventSpec != (module.EventSource{}.EventSpec) {
			trigger := *eventsource.NewEventarcTrigger(stack, &containerModule.EventSource, config.Options)
			destination := trigger.DestinationInput()
			if destination.CloudRunService == nil {
				destination.CloudRunService = &google.EventarcTriggerDestinationCloudRunService{}
			}
			if destination.CloudRunService.Service == nil {
				destination.CloudRunService.Service = runService.Name()
			}
			if destination.CloudRunService.Region == nil {
				destination.CloudRunService.Region = config.Options.Cloud.Gcp.Provider.Region
			}
			trigger.PutDestination(destination)
			serviceAccount := trigger.ServiceAccountInput()
			if serviceAccount == nil {
				serviceAccount = (*runtimeServiceAccount).Email()
			}
			trigger.SetServiceAccount(serviceAccount)
		} else if containerModule.EventSource.TopicSpec != (module.EventSource{}.TopicSpec) {
			topic := *eventsource.NewTopicEventSource(stack, &containerModule.EventSource, config.Options)
			eventarcTriggerConfig := module.EventSource{}
			eventarcTriggerConfig.EventSpec.Gcp.Name = containerModule.EventSource.TopicSpec.Gcp.Name
			eventarcTriggerConfig.EventSpec.Gcp.MatchingCriteria = &[]*google.EventarcTriggerMatchingCriteria{
				{
					Attribute: jsii.String("type"),
					Value:     jsii.String("google.cloud.pubsub.topic.v1.messagePublished"),
				},
			}
			eventarcTriggerConfig.EventSpec.Gcp.Transport = &google.EventarcTriggerTransport{
				Pubsub: &google.EventarcTriggerTransportPubsub{Topic: topic.Id()},
			}
			trigger := *eventsource.NewEventarcTrigger(stack, &eventarcTriggerConfig, config.Options)
			destination := trigger.DestinationInput()
			if destination.CloudRunService == nil {
				destination.CloudRunService = &google.EventarcTriggerDestinationCloudRunService{}
			}
			destination.CloudRunService.Service = runService.Name()
			destination.CloudRunService.Region = config.Options.Cloud.Gcp.Provider.Region
			trigger.PutDestination(destination)
			serviceAccount := trigger.ServiceAccountInput()
			serviceAccount = (*runtimeServiceAccount).Email()
			trigger.SetServiceAccount(serviceAccount)
		} else if containerModule.EventSource.QueueSpec != (module.EventSource{}.QueueSpec) {
			eventsource.NewQueueEventSource(stack, &containerModule.EventSource, config.Options)
		} else if containerModule.EventSource.ScheduleSpec != (module.EventSource{}.ScheduleSpec) {
			schedule := *eventsource.NewScheduleEventSource(stack, &containerModule.EventSource, config.Options)
			httpTarget := schedule.HttpTargetInput()
			if httpTarget == nil {
				httpTarget = &google.CloudSchedulerJobHttpTarget{}
			}
			if httpTarget.Uri == nil {
				httpTarget.Uri = runService.Status().Get(jsii.Number(0)).Url()
			}
			if httpTarget.HttpMethod == nil {
				httpTarget.HttpMethod = jsii.String("POST")
			}
			if httpTarget.OidcToken == nil {
				httpTarget.OidcToken = &google.CloudSchedulerJobHttpTargetOidcToken{
					ServiceAccountEmail: (*runtimeServiceAccount).Email(),
					Audience:            runService.Status().Get(jsii.Number(0)).Url(),
				}
			}
			schedule.PutHttpTarget(httpTarget)
		}
		if containerModule.Security.NoAuthentication {
			security.NewAllUsersCloudRunInvoker(stack, runService.Name(), containerModule, config.Options)
		}
		runMetadata := runService.MetadataInput()
		if runMetadata == nil {
			runMetadata = &google.CloudRunServiceMetadata{}
		}
		if runMetadata.Annotations == nil {
			annotations := make(map[string]*string, 1)
			runMetadata.Annotations = &annotations
		}
		if _, ok := (*runMetadata.Annotations)["run.googleapis.com/ingress"]; !ok {
			var ingress string
			if containerModule.Networking.Internal {
				ingress = "internal"
			} else {
				ingress = "all"
			}
			(*runMetadata.Annotations)["run.googleapis.com/ingress"] = &ingress
		}
		runService.PutMetadata(runMetadata)
		runTemplate := runService.TemplateInput()
		if runTemplate.Metadata.Annotations == nil {
			annotations := make(map[string]*string, 2)
			runTemplate.Metadata.Annotations = &annotations
		}
		if containerModule.Networking.Egress.StaticIp {
			if _, ok := (*runTemplate.Metadata.Annotations)["run.googleapis.com/vpc-access-connector"]; !ok {
				connector := *networking.NewVpcAccessConnector(stack, config.Options)
				(*runTemplate.Metadata.Annotations)["run.googleapis.com/vpc-access-connector"] = connector.Id()
			}
			if _, ok := (*runTemplate.Metadata.Annotations)["run.googleapis.com/vpc-access-egress"]; !ok {
				(*runTemplate.Metadata.Annotations)["run.googleapis.com/vpc-access-egress"] = jsii.String("all-traffic")
			}
		}
		runService.PutTemplate(runTemplate)
		serviceEndpoints[containerModule.Name] =
			serviceendpoint.Endpoint{Uri: *runService.Status().Get(jsii.Number(0)).Url()}
	}
	var apiConfig *googlebeta.GoogleApiGatewayApiConfigA
	if len(config.Manifests) > 0 && len(config.Options.BuildConfig.Files.Networking.Gcp.GatewayPath) > 0 {
		_, apiConfig = networking.NewGateway(stack, nil, nil, config.Options, &betaProvider)
	} else if gatewayModules := modules.gateway; len(gatewayModules) > 0 {
		_, apiConfig = networking.NewGateway(stack, gatewayModules, &serviceEndpoints, config.Options, &betaProvider)
	}
	if apiConfig != nil {
		gatewayConfig := (*apiConfig).GatewayConfigInput()
		if gatewayConfig == nil {
			gatewayConfig = &googlebeta.GoogleApiGatewayApiConfigGatewayConfig{}
		}
		if gatewayConfig.BackendConfig == nil {
			gatewayConfig.BackendConfig = &googlebeta.GoogleApiGatewayApiConfigGatewayConfigBackendConfig{}
		}
		if gatewayConfig.BackendConfig.GoogleServiceAccount == nil {
			gatewayConfig.BackendConfig.GoogleServiceAccount = (*runtimeServiceAccount).Email()
		}
		(*apiConfig).PutGatewayConfig(gatewayConfig)
	}
	var workflow *google.WorkflowsWorkflow
	if len(config.Manifests) > 0 && len(config.Options.BuildConfig.Files.Orchestration.Gcp.WorkflowPath) > 0 {
		workflow = orchestration.NewWorkflow(stack, nil, nil, config.Options)
	} else if workflowModules := modules.workflow; len(workflowModules) > 0 {
		workflow = orchestration.NewWorkflow(stack, workflowModules, &serviceEndpoints, config.Options)
	}
	if workflow != nil {
		if (*workflow).ServiceAccountInput() == nil {
			(*workflow).SetServiceAccount((*runtimeServiceAccount).Email())
		}
	}
	return *stack
}

package networking

import (
	"os"

	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-googlebeta-go/googlebeta/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/service"
)

func NewGateway(
	scope *cdktf.TerraformStack,
	modules []*module.Module,
	endpoints *service.Endpoints,
	options *project.Config,
	betaProvider *googlebeta.GoogleBetaProvider,
) (*googlebeta.GoogleApiGatewayGateway, *googlebeta.GoogleApiGatewayApiConfigA) {
	var openapiDocumentContents *string
	if len(options.BuildConfig.Files.Networking.Gcp.GatewayPath) > 0 {
		contents, err := os.ReadFile(options.BuildConfig.Files.Networking.Gcp.GatewayPath)
		if err != nil {
			panic(err)
		}
		contentsString := string(contents)
		openapiDocumentContents = &contentsString
	} else {
		openapiDocumentContents = generateGatewayOpenapiDocumentContents(modules, endpoints, options)
	}
	apiConf := new(googlebeta.GoogleApiGatewayApiConfig)
	(*apiConf) = options.Cloud.Gcp.Networking.ApiGateway.Api
	apiConf.Provider = *betaProvider
	if apiConf.ApiId == nil {
		apiConf.ApiId = jsii.String("api")
	}
	api := googlebeta.NewGoogleApiGatewayApi(*scope, apiConf.ApiId, apiConf)
	apiConfigConf := new(googlebeta.GoogleApiGatewayApiConfigAConfig)
	(*apiConfigConf) = options.Cloud.Gcp.Networking.ApiGateway.ApiConfig
	apiConfigConf.Provider = *betaProvider
	if apiConfigConf.Api == nil {
		apiConfigConf.Api = api.ApiId()
	}
	if apiConfigConf.OpenapiDocuments == nil {
		apiConfigConf.OpenapiDocuments = &[]*googlebeta.GoogleApiGatewayApiConfigOpenapiDocuments{{
			Document: &googlebeta.GoogleApiGatewayApiConfigOpenapiDocumentsDocument{
				Path:     jsii.String("openapi.json"),
				Contents: openapiDocumentContents,
			},
		}}
	}
	if apiConfigConf.Lifecycle == nil {
		apiConfigConf.Lifecycle = &cdktf.TerraformResourceLifecycle{}
	}
	if apiConfigConf.Lifecycle.CreateBeforeDestroy == nil {
		apiConfigConf.Lifecycle.CreateBeforeDestroy = jsii.Bool(true)
	}
	apiConfig := googlebeta.NewGoogleApiGatewayApiConfigA(*scope, jsii.String(*apiConf.ApiId+"-config"), apiConfigConf)
	gatewayConf := new(googlebeta.GoogleApiGatewayGatewayConfig)
	(*gatewayConf) = options.Cloud.Gcp.Networking.ApiGateway.Gateway
	gatewayConf.Provider = *betaProvider
	if gatewayConf.GatewayId == nil {
		gatewayConf.GatewayId = jsii.String("gateway")
	}
	if gatewayConf.ApiConfig == nil {
		gatewayConf.ApiConfig = apiConfig.Id()
	}
	if gatewayConf.Region == nil {
		gatewayConf.Region = options.Cloud.Gcp.BetaProvider.Region
	}
	gateway := googlebeta.NewGoogleApiGatewayGateway(*scope, gatewayConf.GatewayId, gatewayConf)
	return &gateway, &apiConfig
}

func IsGateway(n *module.Networking) bool {
	return len(n.Ingress.Gateway.Paths) > 0 || len(n.Ingress.Gateway.Components.SecuritySchemes) > 0
}

func generateGatewayOpenapiDocumentContents(
	modules []*module.Module,
	endpoints *service.Endpoints,
	options *project.Config,
) *string {
	var gateway gateway
	gateway.Swagger = "2.0"
	gateway.Info.Title = *options.Cloud.Gcp.BetaProvider.Project + "-openapi"
	gateway.Info.Version = *options.Cloud.Gcp.BetaProvider.Project + "-openapi"
	var numPaths int
	var numSecuritySchemes int
	for i := range modules {
		numPaths += len(modules[i].Networking.Ingress.Gateway.Paths)
		numSecuritySchemes += len(modules[i].Networking.Ingress.Gateway.Components.SecuritySchemes)
	}
	gateway.Paths = make(map[string]map[string]gatewayOperation, numPaths)
	if numSecuritySchemes > 0 {
		gateway.SecurityDefinitions = make(map[string]gatewaySecurityScheme, numSecuritySchemes)
	}
	for i := range modules {
		pathsConfig := &modules[i].Networking.Ingress.Gateway.Paths
		for path := range *pathsConfig {
			gateway.Paths["\""+path+"\""] = make(map[string]gatewayOperation, len((*pathsConfig)[path]))
			for operation := range (*pathsConfig)[path] {
				operationConfig := (*pathsConfig)[path][operation]
				gatewayOperation := gatewayOperation{}
				gatewayOperation.OperationId = path + operation
				gatewayOperation.XGoogleBackend.Address = (*endpoints)[modules[i].Name].Uri
				if parametersLen :=
					len(operationConfig.Parameters) + len(operationConfig.RequestBody.Content); parametersLen > 0 {
					gatewayOperation.Parameters = make([]gatewayParameter, 0, parametersLen)
				}
				for j := range operationConfig.Parameters {
					parameter := operationConfig.Parameters[j]
					gatewayParameter := gatewayParameter{}
					gatewayParameter.Name = parameter.Name
					gatewayParameter.In = parameter.In
					gatewayParameter.Required = parameter.Required
					for contentType := range parameter.Content {
						if canInsert(gatewayOperation.Consumes, contentType) {
							gatewayOperation.Consumes = append(gatewayOperation.Consumes, contentType)
						}
						parameterSchema := parameter.Content[contentType].Schema
						if parameterType, ok := parameterSchema["type"].(string); ok {
							gatewayParameter.Type = &parameterType
						}
						if gatewayParameter.Type != nil && *gatewayParameter.Type == "array" {
							if parameterItems, ok := parameterSchema["items"].(map[string]any); ok {
								if itemsType, ok := parameterItems["type"].(string); ok {
									gatewayParameter.Items = &gatewayItems{Type: itemsType}
								}
							}
						}
					}
					gatewayOperation.Parameters = append(gatewayOperation.Parameters, gatewayParameter)
				}
				if len(operationConfig.RequestBody.Content) > 0 {
					gatewayParameter := gatewayParameter{}
					gatewayParameter.In = "body"
					for contentType := range operationConfig.RequestBody.Content {
						gatewayParameter.Name = contentType
						if canInsert(gatewayOperation.Consumes, contentType) {
							gatewayOperation.Consumes = append(gatewayOperation.Consumes, contentType)
						}
						gatewayParameter.Schema = new(gatewaySchema)
						*gatewayParameter.Schema = operationConfig.RequestBody.Content[contentType].Schema
					}
					gatewayOperation.Parameters = append(gatewayOperation.Parameters, gatewayParameter)
				}
				gatewayOperation.Responses = make(map[string]gatewayResponse, len(operationConfig.Responses))
				for code := range operationConfig.Responses {
					responseConfig := operationConfig.Responses[code]
					gatewayResponse := gatewayResponse{}
					gatewayResponse.Description = responseConfig.Description
					if headersLen := len(responseConfig.Headers); headersLen > 0 {
						gatewayResponse.Headers = make(map[string]gatewayHeader, headersLen)
					}
					for header := range responseConfig.Headers {
						for contentType := range responseConfig.Headers[header].Content {
							if canInsert(gatewayOperation.Produces, contentType) {
								gatewayOperation.Produces = append(gatewayOperation.Produces, contentType)
							}
							gatewayHeader := gatewayHeader{}
							headerSchema := responseConfig.Headers[header].Content[contentType].Schema
							if headerType, ok := headerSchema["type"].(string); ok {
								gatewayHeader.Type = headerType
							}
							if gatewayHeader.Type == "array" {
								if headerItems, ok := headerSchema["items"].(map[string]any); ok {
									if itemsType, ok := headerItems["type"].(string); ok {
										gatewayHeader.Items = &gatewayItems{Type: itemsType}
									}
								}
							}
							gatewayResponse.Headers[header] = gatewayHeader
						}
					}
					for contentType := range responseConfig.Content {
						if canInsert(gatewayOperation.Produces, contentType) {
							gatewayOperation.Produces = append(gatewayOperation.Produces, contentType)
						}
						gatewayResponse.Schema = responseConfig.Content[contentType].Schema
					}
					gatewayOperation.Responses[code] = gatewayResponse
				}
				gatewayOperation.Security = operationConfig.Security
				gateway.Paths["\""+path+"\""][operation] = gatewayOperation
			}
		}
		securityConfig := &modules[i].Networking.Ingress.Gateway.Components.SecuritySchemes
		for securityScheme := range *securityConfig {
			if _, ok := gateway.SecurityDefinitions[securityScheme]; !ok {
				gateway.SecurityDefinitions[securityScheme] = gatewaySecurityScheme((*securityConfig)[securityScheme])
			}
		}
	}
	return cdktf.Fn_Base64encode(cdktf.Fn_Jsonencode(gateway))
}

func canInsert(slice []string, element string) bool {
	canInsert := true
	for i := range slice {
		if element == slice[i] {
			canInsert = false
			break
		}
	}
	return canInsert
}

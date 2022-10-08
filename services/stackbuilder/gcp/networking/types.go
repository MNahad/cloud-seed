package networking

import (
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
)

type networking struct {
	vpcAccessConnector *google.VpcAccessConnector
}

func NewNetworking() *networking {
	return &networking{}
}

type gateway struct {
	Swagger string `json:"swagger"`
	Info    struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	Paths               map[string]map[string]gatewayOperation `json:"paths"`
	SecurityDefinitions map[string]gatewaySecurityScheme       `json:"securityDefinitions"`
}

type gatewayOperation struct {
	OperationId    string `json:"operationId"`
	XGoogleBackend struct {
		Address string `json:"address"`
	} `json:"x-google-backend"`
	Consumes   []string                   `json:"consumes"`
	Produces   []string                   `json:"produces"`
	Parameters []gatewayParameter         `json:"parameters"`
	Responses  map[string]gatewayResponse `json:"responses"`
	Security   []map[string][]string      `json:"security"`
}

type gatewayParameter struct {
	Name     string         `json:"name"`
	In       string         `json:"in"`
	Required bool           `json:"required"`
	Type     *string        `json:"type"`
	Items    *gatewayItems  `json:"items"`
	Schema   *gatewaySchema `json:"schema"`
}

type gatewayResponse struct {
	Description string                   `json:"description"`
	Headers     map[string]gatewayHeader `json:"headers"`
	Schema      gatewaySchema            `json:"schema"`
}

type gatewayHeader struct {
	Type  string        `json:"type"`
	Items *gatewayItems `json:"items"`
}

type gatewayItems struct {
	Type string `json:"type"`
}

type gatewaySchema map[string]any

type gatewaySecurityScheme struct {
	Type string `json:"type"`
	Name string `json:"name"`
	In   string `json:"in"`
}

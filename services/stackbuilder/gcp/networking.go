package gcp

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
	"github.com/mnahad/cloud-seed/services/config/project"
)

var networkingVpcAccessConnector *google.VpcAccessConnector

func generateVpcAccessConnector(scope *cdktf.TerraformStack, options *project.Config) *google.VpcAccessConnector {
	if networkingVpcAccessConnector == nil {
		networkingVpcAccessConnector = newStaticIpNetwork(scope, options)
	}
	return networkingVpcAccessConnector
}

func newStaticIpNetwork(scope *cdktf.TerraformStack, options *project.Config) *google.VpcAccessConnector {
	networkConfig := new(google.ComputeNetworkConfig)
	(*networkConfig) = options.Cloud.Gcp.StaticIpNetwork.Network
	if networkConfig.Name == nil {
		networkConfig.Name = jsii.String("static-ip-net")
	}
	if networkConfig.AutoCreateSubnetworks == nil {
		networkConfig.AutoCreateSubnetworks = false
	}
	network := google.NewComputeNetwork(*scope, networkConfig.Name, networkConfig)
	staticIpName := *networkConfig.Name + "-ip"
	staticIp := google.NewComputeAddress(*scope, &staticIpName, &google.ComputeAddressConfig{
		Name:        &staticIpName,
		Region:      &options.Cloud.Gcp.Region,
		AddressType: jsii.String("EXTERNAL"),
	})
	routerName := *networkConfig.Name + "-router"
	router := google.NewComputeRouter(*scope, &routerName, &google.ComputeRouterConfig{
		Name:    &routerName,
		Region:  &options.Cloud.Gcp.Region,
		Network: network.Id(),
	})
	natName := *networkConfig.Name + "-nat"
	google.NewComputeRouterNat(*scope, &natName, &google.ComputeRouterNatConfig{
		Name:                          &natName,
		Region:                        &options.Cloud.Gcp.Region,
		Router:                        router.Name(),
		NatIpAllocateOption:           jsii.String("MANUAL_ONLY"),
		NatIps:                        &[]*string{staticIp.SelfLink()},
		SourceSubnetworkIpRangesToNat: jsii.String("ALL_SUBNETWORKS_ALL_IP_RANGES"),
	})
	connectorName := *networkConfig.Name + "-connector"
	connector := google.NewVpcAccessConnector(*scope, &connectorName, &google.VpcAccessConnectorConfig{
		Name:        &connectorName,
		Region:      &options.Cloud.Gcp.Region,
		Network:     network.Name(),
		IpCidrRange: jsii.String("10.0.0.0/28"),
	})
	return &connector
}

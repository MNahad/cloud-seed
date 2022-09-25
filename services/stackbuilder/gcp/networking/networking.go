package networking

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/project"
)

var vpcAccessConnector *google.VpcAccessConnector

func NewVpcAccessConnector(scope *cdktf.TerraformStack, options *project.Config) *google.VpcAccessConnector {
	if vpcAccessConnector == nil {
		vpcAccessConnector = newStaticIpNetwork(scope, options)
	}
	return vpcAccessConnector
}

func newStaticIpNetwork(scope *cdktf.TerraformStack, options *project.Config) *google.VpcAccessConnector {
	networkConfig := new(google.ComputeNetworkConfig)
	(*networkConfig) = options.Cloud.Gcp.Networking.StaticIpNetwork.Network
	if networkConfig.Name == nil {
		networkConfig.Name = jsii.String("static-ip-net")
	}
	if networkConfig.AutoCreateSubnetworks == nil {
		networkConfig.AutoCreateSubnetworks = false
	}
	network := google.NewComputeNetwork(*scope, networkConfig.Name, networkConfig)
	staticIpConfig := new(google.ComputeAddressConfig)
	(*staticIpConfig) = options.Cloud.Gcp.Networking.StaticIpNetwork.Address
	if staticIpConfig.Name == nil {
		staticIpConfig.Name = jsii.String(*networkConfig.Name + "-ip")
	}
	if staticIpConfig.AddressType == nil {
		staticIpConfig.AddressType = jsii.String("EXTERNAL")
	}
	if staticIpConfig.Region == nil {
		staticIpConfig.Region = options.Cloud.Gcp.Provider.Region
	}
	staticIp := google.NewComputeAddress(*scope, staticIpConfig.Name, staticIpConfig)
	routerConfig := new(google.ComputeRouterConfig)
	(*routerConfig) = options.Cloud.Gcp.Networking.StaticIpNetwork.Router
	if routerConfig.Name == nil {
		routerConfig.Name = jsii.String(*networkConfig.Name + "-router")
	}
	if routerConfig.Network == nil {
		routerConfig.Network = network.Id()
	}
	if routerConfig.Region == nil {
		routerConfig.Region = options.Cloud.Gcp.Provider.Region
	}
	router := google.NewComputeRouter(*scope, routerConfig.Name, routerConfig)
	natConfig := new(google.ComputeRouterNatConfig)
	(*natConfig) = options.Cloud.Gcp.Networking.StaticIpNetwork.RouterNat
	if natConfig.Name == nil {
		natConfig.Name = jsii.String(*networkConfig.Name + "-nat")
	}
	if natConfig.Router == nil {
		natConfig.Router = router.Name()
	}
	if natConfig.NatIpAllocateOption == nil {
		natConfig.NatIpAllocateOption = jsii.String("MANUAL_ONLY")
	}
	if natConfig.NatIps == nil {
		natConfig.NatIps = &[]*string{staticIp.SelfLink()}
	}
	if natConfig.SourceSubnetworkIpRangesToNat == nil {
		natConfig.SourceSubnetworkIpRangesToNat = jsii.String("ALL_SUBNETWORKS_ALL_IP_RANGES")
	}
	if natConfig.Region == nil {
		natConfig.Region = options.Cloud.Gcp.Provider.Region
	}
	google.NewComputeRouterNat(*scope, natConfig.Name, natConfig)
	connectorConfig := new(google.VpcAccessConnectorConfig)
	(*connectorConfig) = options.Cloud.Gcp.Networking.StaticIpNetwork.AccessConnector
	if connectorConfig.Name == nil {
		connectorConfig.Name = jsii.String(*networkConfig.Name + "-connector")
	}
	if connectorConfig.Network == nil {
		connectorConfig.Network = network.Name()
	}
	if connectorConfig.IpCidrRange == nil {
		connectorConfig.IpCidrRange = jsii.String("10.0.0.0/28")
	}
	if connectorConfig.Region == nil {
		connectorConfig.Region = options.Cloud.Gcp.Provider.Region
	}
	connector := google.NewVpcAccessConnector(*scope, connectorConfig.Name, connectorConfig)
	return &connector
}

# Ingress over API Gateway and Egress over Static IP NAT

Cloud Seed provides an easy way to build ingress API Gateways from multiple compute services. It also provides an easy way to route egress traffic via a static IP using VPC-native NAT.

## Manifest

1. A GCP HTTP Python Cloud Function which demonstrates how an API Gateway spec can be attached to it.
1. A GCP HTTP Python Cloud Function which is attached to the API Gateway, but with additional config that sets up Security Requirements. This binds the function to Security Schemes that are both defined on this function and on the previous function, demonstrating the distributed nature of the config. This function also has egress traffic routed through a static IP address.
1. A GCP HTTP Python Cloud Function that is locked-down to internal traffic only.

## Details

Services configured with Cloud Seed's generated IaC can optionally have their ingress traffic routed through an API Gateway. This can be achieved by adding a `networking.ingress.gateway` object to the relevant service's config.

The Gateway CDKTF Resource is built dynamically using all services which have this config defined.

The spec of this config is based on OpenAPI Specification (OAS) 3.1.0. Please refer to the official specification for more information.

For more complex API Gateway architectures, a cloud-platform specific file can be provided by pointing its filepath in the `buildConfig.files.networking` property of the project's `cloudseed.json` config. This will override any API Gateways defined by other means.

Service egress traffic can be routed optionally via a static IP, by setting the `networking.egress.staticIp` flag in the config. This will generate CDKTF VPC Resources that create and attach the service to a VPC-native static IP NAT.

The network of a service can be made internal-only by setting the flag `networking.internal` to `true`. Refer to the example config.

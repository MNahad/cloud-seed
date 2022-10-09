# Function-as-a-Service

This example demonstrates how to configure IaC targeting FaaS.

## Manifest

1. A GCP HTTP Node.js Cloud Function
1. A subdirectory with:
    1. A GCP Pubsub Python function
    1. A GCP Queue Python function
    1. A GCP Schedule Python function

## Details

Both event sources and function services can be defined in the same `module.cloudseed.json` file, and the CDKTF output will link the resources together.

The function and event source CDKTF resource attributes can be modified, by passing in custom arguments in the respective resource attribute objects in the config.

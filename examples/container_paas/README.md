# Container Platform-as-a-Service

This example demonstrates how to configure IaC targeting Container PaaS.

## Manifest

1. A GCP Node.js Cloud Run application triggered by:
    1. An HTTP trigger
    1. An Eventarc trigger on a Cloud Storage bucket

## Details

This example prepares two Cloud Run services, each with its own respective event source.

The container and event source CDKTF resource attributes can be modified, by passing in custom arguments in the respective resource attribute objects in the config.

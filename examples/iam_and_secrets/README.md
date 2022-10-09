# IAM and Secrets

This example shows the various IAM and Secrets management options that are available.

## Manifest

1. A GCP Node.js HTTP Cloud Function which has `security.noAuthentication` set to `true`
1. Secrets added to GCP Secret Manager

## Details

This example shows how allowed incoming requests can be unauthenticated by setting the flag to `true`.

All Cloud Seed IaC deployments have a custom runtime service account created by default, which is then bound to the identity of compute services. This is also the account that incoming requests are authenticated against when Event Sources or Workflows need to invoke the service.

This example shows how an override can be added that modifies the CDKTF `accountId` attribute of that created service account

This example also shows how to add secret names in GCP Secret Manager, by adding their names to the `cloudseed.json` project config.

Additionally, when selecting the example `production` cloud seed environment override (as defined in `cloudseed.json`), the CDKTF GCP Secret Manager Secret resources will be set with the `replication` attribute set to automatic.

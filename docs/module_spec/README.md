# `module.cloudseed.json` File Specification

For each entrypoint in the source code (function or containerised app), one file that ends in `.cloudseed.json` is added in the respective directory. This file contains a Manifest of all Modules in that directory.

A Module is defined as a grouping the CDKTF Constructs that enable an independent compute service (e.g. FaaS / container PaaS) to be defined as IaC. This is typically one event source resource with one service resource (e.g. a schedule-based event source and a FaaS function).

A Manifest is defined as the listing of all Modules in a given directory.

Services can share CDKTF resources that are defined by the Cloud Seed IaC (e.g. a shared Workflow / a shared ingress API Gateway / a shared VPC attachment for egress traffic).

The file's specification is defined in [module.cloudseed.jsonc](/docs/module_spec/module.cloudseed.jsonc).

# Project setup

Assume that the following example project structure is in place:

```text
src/
|____ {nested_dir}/
|     |____ {entrypoint}
|     |____ {entrypoint}.cloudseed.json
|     |____ ...
|____ {entrypoint}
|____ {entrypoint}.cloudseed.json
|____ ...
generated/
|____ ...
cloudseed.json
```

Source code is assumed to be in the `src/` directory and the desired IaC will be assumed to be in the `generated/` directory. (These directory names can be changed in the `cloudseed.json` project config file.)

A `cloudseed.json` file is required at the top-level, which will contain project-level config that will be ingested by the CLI.

When Cloud Seed is executed in the directory, it will walk the directory and detect "marker" files that match the pattern `*\.cloudseed\.json$`. These will indicate to Cloud Seed that there are entrypoints within those directories.

As an example, the above directory structure contains `src/{entrypoint}.cloudseed.json` and `src/{nested_dir}/{entrypoint}.cloudseed.json`. This means that there are two entrypoints in the source code, and each should be deployed as per their respective `{entrypoint}.cloudseed.json` config.

An `{entrypoint}` can be any appropriate source code file (such as an `index.js` or a `main.py` file), or a containerisable application.

# Generated IaC

Additionally, the `generated/` directory contains the following generated tree:

```text
generated/
|____ artefacts/
|     |____ aws/
|     |     |____ ...
|     |____ gcp/
|           |____ containers/
|           |     |____ {archive}
|           |     |____ ...
|           |____ functions/
|                 |____ {archive}
|                 |____ ...
|____ stacks/
|     |____ cloudseed/
|           |____ ...
|____ manifest.json
```

The `generated/stacks/` directory together with the `generated/manifest.json` will contain the output CDKTF files, which can be read by tools such as `cdktf` and `terraform` CLI.

The `generated/artefacts/` directory will contain prepared archives of each entrypoint source code directory, and their respective filepaths will be included in the CDKTF config. This way the archives are ready to be uploaded to the target cloud environment during a deployment.

# File Specifications

- Refer to [project_spec/](/docs/project_spec/) for the project-level config specification (`cloudseed.json`)
- Refer to [module_spec/](/docs/module_spec/) for the source code module specification (`{entrypoint}.cloudseed.json`)

# CLI Specification

## Build

```text
Usage:
  cloud-seed build [flags]

Flags:
  -e, --environment string   Set an environment
  -d, --project-dir string   Select a project directory
  -h, --help                 help for build
```

This command builds the HCL JSON-format IaC config using CDKTF and outputs it in the directory specified in `cloudseed.json`.

The `--environment` flag can be used to optionally override the `cloudseed.json` config's `default` values with `environmentOverrides` matching the environment name set in the same file.

The `--project-dir` flag can be used to optionally specify the location of the project root if it is not the current working directory.

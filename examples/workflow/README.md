# Workflows

This example shows how to build a GCP Workflow that implements switching and loops.

## Manifest

1. A subdirectory 'a' which contains:
    1. A `start` GCP Python HTTP Cloud Function, acting as the start state of the Workflow.
    1. A GCP Python HTTP Cloud Function which implements the loop counter decrement logic.
1. A subdirectory 'b' which contains:
    1. A GCP Python HTTP Cloud Function which implements a loop counter terminal check.
    1. An `end` GCP Python HTTP Cloud Function, which acts as the end state of the Workflow, and outputs the result.

## Details

Cloud Seed provides a way to easily build a Worklflow which orchestrates multiple compute services.

Each service that will be part of a workflow can be optionally provided an `orchestration.workflow` object as part of its config. This will contain all the input and output details that are related to that service as part of the Workflow.

The Workflow CDKTF Resource is then built dynamically using all the services which have this config defined.

This example shows how four GCP Cloud Functions can be connected as part of a Workflow.

The transition of states is as follows:

1. The Workflow begins when the start function is triggered, which will output a counter, captured as a GCP Workflow variable.
1. Another decrements the counter and passes it by jumping to the third function.
1. A third function implements a switch condition at its output, which checks if the counter has reached the terminal value. If it does not, then the loop continues by jumping to the previous function.
1. If the loop counter has reached its terminal value, then the fourth function is invoked, ending the workflow and outputting the value.

For more complex Workflow architectures, a cloud-platform specific file can be provided by pointing its filepath in the `buildConfig.files.orchestration` property of the project's `cloudseed.json` config. This will override any Workflows defined by other means.

package orchestration

import (
	"os"

	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder/gcp/service"
)

func NewWorkflow(
	scope *cdktf.TerraformStack,
	modules []*module.Module,
	endpoints *service.Endpoints,
	options *project.Config,
) *google.WorkflowsWorkflow {
	var sourceContents *string
	if len(options.BuildConfig.Files.Orchestration.Gcp.WorkflowPath) > 0 {
		contents, err := os.ReadFile(options.BuildConfig.Files.Orchestration.Gcp.WorkflowPath)
		if err != nil {
			panic(err)
		}
		contentsString := string(contents)
		sourceContents = &contentsString
	} else {
		sourceContents = generateWorkflowSourceContents(modules, endpoints)
	}
	workflowConfig := new(google.WorkflowsWorkflowConfig)
	(*workflowConfig) = options.Cloud.Gcp.Orchestration.Workflow
	if workflowConfig.Name == nil {
		workflowConfig.Name = jsii.String("workflow")
	}
	if workflowConfig.SourceContents == nil {
		workflowConfig.SourceContents = sourceContents
	}
	if workflowConfig.Region == nil {
		workflowConfig.Region = options.Cloud.Gcp.Provider.Region
	}
	workflow := google.NewWorkflowsWorkflow(*scope, workflowConfig.Name, workflowConfig)
	return &workflow
}

func IsWorkflow(o *module.Orchestration) bool {
	workflow := (module.Orchestration{}.Workflow)
	return o.Workflow.Start ||
		o.Workflow.End ||
		o.Workflow.Input != workflow.Input ||
		o.Workflow.Output != workflow.Output ||
		o.Workflow.Next.Jump != workflow.Next.Jump ||
		len(o.Workflow.Next.Condition) > 0
}

func generateWorkflowSourceContents(modules []*module.Module, endpoints *service.Endpoints) *string {
	var workflow workflow
	var steps stepsCollection
	for i := range modules {
		moduleConfig := modules[i]
		workflowConfig := &moduleConfig.Orchestration.Workflow
		var callStep call
		callStep.Call = "http.post"
		callStep.Args.Url = (*endpoints)[moduleConfig.Name].Uri
		if !moduleConfig.Security.NoAuthentication {
			callStep.Args.Auth = &callAuth{Type: jsii.String("OIDC")}
		}
		if len(workflowConfig.Input.Expression.Gcp) > 0 {
			callStep.Args.Body = cdktf.Fn_RawString(&workflowConfig.Input.Expression.Gcp)
		}
		if len(workflowConfig.Output.Expression.Gcp) > 0 {
			callStep.Result = cdktf.Fn_RawString(&workflowConfig.Output.Expression.Gcp)
		}
		if conditions := workflowConfig.Next.Condition; len(conditions) > 0 {
			var conditionStep condition
			conditionStep.Switch = make(conditionSwitch, len(conditions))
			for j := range conditions {
				conditionStep.Switch[j].Condition = *cdktf.Fn_RawString(&conditions[j].Expression.Gcp)
				conditionStep.Switch[j].Next = conditions[j].ServiceName
			}
			conditionMap := make(step[condition], 1)
			conditionMap[moduleConfig.Name+"-switch"] = conditionStep
			steps.StepsCondition.Steps = append(steps.StepsCondition.Steps, conditionMap)
			callStep.Next = moduleConfig.Name + "-switch"
		}
		if jump := &workflowConfig.Next.Jump; *jump != (module.Orchestration{}.Workflow.Next.Jump) {
			callStep.Next = jump.ServiceName
		}
		if workflowConfig.End {
			if len(workflowConfig.Output.Expression.Gcp) > 0 {
				var returnStep ret
				returnStep.Return = *cdktf.Fn_RawString(
					jsii.String("${" + workflowConfig.Output.Expression.Gcp + "}"),
				)
				returnMap := make(step[ret], 1)
				returnMap[moduleConfig.Name+"-return"] = returnStep
				steps.StepsReturn.Steps = append(steps.StepsReturn.Steps, returnMap)
				callStep.Next = moduleConfig.Name + "-return"
			} else {
				callStep.Next = "end"
			}
		}
		callMap := make(step[call], 1)
		callMap[moduleConfig.Name] = callStep
		if workflowConfig.Start {
			steps.StepsCall.Steps = prependStep(steps.StepsCall.Steps, &callMap)
			if len(workflowConfig.Input.Expression.Gcp) > 0 {
				workflow.Main.Params = []string{"args"}
			}
		} else {
			steps.StepsCall.Steps = append(steps.StepsCall.Steps, callMap)
		}
	}
	for i := range steps.StepsCall.Steps {
		workflow.Main.Steps = append(workflow.Main.Steps, steps.StepsCall.Steps[i])
	}
	for i := range steps.StepsCondition.Steps {
		workflow.Main.Steps = append(workflow.Main.Steps, steps.StepsCondition.Steps[i])
	}
	for i := range steps.StepsReturn.Steps {
		workflow.Main.Steps = append(workflow.Main.Steps, steps.StepsReturn.Steps[i])
	}
	return cdktf.Fn_Jsonencode(workflow)
}

func prependStep[T call | condition | ret](slice []step[T], element *step[T]) []step[T] {
	if len(slice) >= 1 {
		slice = append(slice, slice[len(slice)-1])
		copy(slice[1:len(slice)-1], slice[0:len(slice)-2])
		slice[0] = *element
	} else {
		slice = append(slice, *element)
	}
	return slice
}

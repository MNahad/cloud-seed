package gcp

import (
	"strconv"

	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

var eventSourceTopics = make(map[string]*google.PubsubTopic, 0)
var eventSourceQueues = make(map[string]*google.CloudTasksQueue, 0)
var eventSourceSchedulesCount int

const cloudSchedulerPlaceholderHttpTargetUri = "http://example.com/cloud-seed"

func newTopicEventSource(scope *cdktf.TerraformStack, eventSource *module.EventSource) *google.PubsubTopic {
	name := *eventSource.EventSpec.Gcp.Name
	if eventSourceTopics[name] == nil {
		topic := google.NewPubsubTopic(*scope, &name, &eventSource.EventSpec.Gcp)
		eventSourceTopics[name] = &topic
	}
	return eventSourceTopics[name]
}

func newQueueEventSource(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.CloudTasksQueue {
	name := *eventSource.QueueSpec.Gcp.Name
	if eventSourceQueues[name] == nil {
		if eventSource.QueueSpec.Gcp.Location == nil {
			eventSource.QueueSpec.Gcp.Location = &options.Cloud.Gcp.Region
		}
		queue := google.NewCloudTasksQueue(*scope, &name, &eventSource.QueueSpec.Gcp)
		eventSourceQueues[name] = &queue
	}
	return eventSourceQueues[name]
}

func newScheduleEventSource(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.CloudSchedulerJob {
	if eventSource.ScheduleSpec.Gcp.Name == nil {
		eventSource.ScheduleSpec.Gcp.Name = jsii.String("Scheduler" + strconv.Itoa(eventSourceSchedulesCount))
		eventSourceSchedulesCount += 1
	}
	if eventSource.ScheduleSpec.Gcp.Region == nil {
		eventSource.ScheduleSpec.Gcp.Region = &options.Cloud.Gcp.Region
	}
	if eventSource.ScheduleSpec.Gcp.HttpTarget != nil {
		if eventSource.ScheduleSpec.Gcp.HttpTarget.Uri == nil {
			eventSource.ScheduleSpec.Gcp.HttpTarget.Uri = jsii.String(cloudSchedulerPlaceholderHttpTargetUri)
		}
	}
	name := eventSource.ScheduleSpec.Gcp.Name
	job := google.NewCloudSchedulerJob(*scope, name, &eventSource.ScheduleSpec.Gcp)
	return &job
}

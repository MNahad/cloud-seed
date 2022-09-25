package eventsource

import (
	"strconv"

	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

const PlaceholderHttpTargetUri = "http://example.com/cloud-seed"

var topics = make(map[string]*google.PubsubTopic)
var queues = make(map[string]*google.CloudTasksQueue)
var schedulesCount uint64

func NewTopicEventSource(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.PubsubTopic {
	name := *eventSource.EventSpec.Gcp.Name
	if topics[name] == nil {
		if eventSource.EventSpec.Gcp.MessageStoragePolicy == nil {
			eventSource.EventSpec.Gcp.MessageStoragePolicy = &google.PubsubTopicMessageStoragePolicy{
				AllowedPersistenceRegions: &[]*string{options.Cloud.Gcp.Provider.Region},
			}
		}
		topic := google.NewPubsubTopic(*scope, &name, &eventSource.EventSpec.Gcp)
		topics[name] = &topic
	}
	return topics[name]
}

func NewQueueEventSource(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.CloudTasksQueue {
	name := *eventSource.QueueSpec.Gcp.Name
	if queues[name] == nil {
		if eventSource.QueueSpec.Gcp.Location == nil {
			eventSource.QueueSpec.Gcp.Location = options.Cloud.Gcp.Provider.Region
		}
		queue := google.NewCloudTasksQueue(*scope, &name, &eventSource.QueueSpec.Gcp)
		queues[name] = &queue
	}
	return queues[name]
}

func NewScheduleEventSource(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.CloudSchedulerJob {
	if eventSource.ScheduleSpec.Gcp.Name == nil {
		eventSource.ScheduleSpec.Gcp.Name = jsii.String("Scheduler" + strconv.FormatUint(schedulesCount, 10))
		schedulesCount += 1
	}
	if eventSource.ScheduleSpec.Gcp.Region == nil {
		eventSource.ScheduleSpec.Gcp.Region = options.Cloud.Gcp.Provider.Region
	}
	if eventSource.ScheduleSpec.Gcp.HttpTarget != nil {
		if eventSource.ScheduleSpec.Gcp.HttpTarget.Uri == nil {
			eventSource.ScheduleSpec.Gcp.HttpTarget.Uri = jsii.String(PlaceholderHttpTargetUri)
		}
	}
	name := eventSource.ScheduleSpec.Gcp.Name
	job := google.NewCloudSchedulerJob(*scope, name, &eventSource.ScheduleSpec.Gcp)
	return &job
}

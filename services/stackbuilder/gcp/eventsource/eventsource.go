package eventsource

import (
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

var topics = make(map[string]*google.PubsubTopic)
var queues = make(map[string]*google.CloudTasksQueue)

func NewTopicEventSource(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.PubsubTopic {
	name := *eventSource.TopicSpec.Gcp.Name
	if topics[name] == nil {
		topicConfig := new(google.PubsubTopicConfig)
		(*topicConfig) = eventSource.TopicSpec.Gcp
		if topicConfig.MessageStoragePolicy == nil {
			topicConfig.MessageStoragePolicy = &google.PubsubTopicMessageStoragePolicy{
				AllowedPersistenceRegions: &[]*string{options.Cloud.Gcp.Provider.Region},
			}
		}
		topic := google.NewPubsubTopic(*scope, &name, topicConfig)
		topics[name] = &topic
	}
	return topics[name]
}

func NewEventarcTrigger(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.EventarcTrigger {
	triggerConfig := new(google.EventarcTriggerConfig)
	(*triggerConfig) = eventSource.EventSpec.Gcp
	if triggerConfig.Location == nil {
		triggerConfig.Location = options.Cloud.Gcp.Provider.Region
	}
	if triggerConfig.Destination == nil {
		triggerConfig.Destination = &google.EventarcTriggerDestination{}
	}
	trigger := google.NewEventarcTrigger(*scope, triggerConfig.Name, triggerConfig)
	return &trigger
}

func NewQueueEventSource(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.CloudTasksQueue {
	name := *eventSource.QueueSpec.Gcp.Name
	if queues[name] == nil {
		queueConfig := new(google.CloudTasksQueueConfig)
		(*queueConfig) = eventSource.QueueSpec.Gcp
		if queueConfig.Location == nil {
			queueConfig.Location = options.Cloud.Gcp.Provider.Region
		}
		queue := google.NewCloudTasksQueue(*scope, &name, queueConfig)
		queues[name] = &queue
	}
	return queues[name]
}

func NewScheduleEventSource(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.CloudSchedulerJob {
	schedulerConfig := new(google.CloudSchedulerJobConfig)
	(*schedulerConfig) = eventSource.ScheduleSpec.Gcp
	if schedulerConfig.Region == nil {
		schedulerConfig.Region = options.Cloud.Gcp.Provider.Region
	}
	job := google.NewCloudSchedulerJob(*scope, schedulerConfig.Name, schedulerConfig)
	return &job
}

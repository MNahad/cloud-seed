package eventsource

import (
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

type eventSource struct {
	topics map[string]*google.PubsubTopic
	queues map[string]*google.CloudTasksQueue
}

func NewEventSource() *eventSource {
	return &eventSource{
		topics: make(map[string]*google.PubsubTopic),
		queues: make(map[string]*google.CloudTasksQueue),
	}
}

func (e *eventSource) NewTopicEventSource(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.PubsubTopic {
	name := *eventSource.TopicSpec.Gcp.Name
	if e.topics[name] == nil {
		topicConfig := new(google.PubsubTopicConfig)
		(*topicConfig) = eventSource.TopicSpec.Gcp
		if topicConfig.MessageStoragePolicy == nil {
			topicConfig.MessageStoragePolicy = &google.PubsubTopicMessageStoragePolicy{
				AllowedPersistenceRegions: &[]*string{options.Cloud.Gcp.Provider.Region},
			}
		}
		topic := google.NewPubsubTopic(*scope, &name, topicConfig)
		e.topics[name] = &topic
	}
	return e.topics[name]
}

func (e *eventSource) NewEventarcTrigger(
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

func (e *eventSource) NewQueueEventSource(
	scope *cdktf.TerraformStack,
	eventSource *module.EventSource,
	options *project.Config,
) *google.CloudTasksQueue {
	name := *eventSource.QueueSpec.Gcp.Name
	if e.queues[name] == nil {
		queueConfig := new(google.CloudTasksQueueConfig)
		(*queueConfig) = eventSource.QueueSpec.Gcp
		if queueConfig.Location == nil {
			queueConfig.Location = options.Cloud.Gcp.Provider.Region
		}
		queue := google.NewCloudTasksQueue(*scope, &name, queueConfig)
		e.queues[name] = &queue
	}
	return e.queues[name]
}

func (e *eventSource) NewScheduleEventSource(
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

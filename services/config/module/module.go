package module

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/mnahad/cloud-seed/services/config/project"
)

type Manifest struct {
	Path    string
	Modules []Module
}

func (m *Manifest) FilterModules(predicates []func(*Module) bool) [][]*Module {
	filtered := make([][]*Module, len(predicates))
	for i := range predicates {
		filtered[i] = make([]*Module, 0, len(m.Modules))
	}
	for i := range m.Modules {
		module := &m.Modules[i]
		for j := range predicates {
			if predicates[j](module) {
				filtered[j] = append(filtered[j], module)
			}
		}
	}
	return filtered
}

func (m *Manifest) UnmarshalJSON(b []byte) error {
	if b == nil {
		return nil
	}
	type Entrypoints map[string]Module
	modules := new(Entrypoints)
	var err error
	if err = json.Unmarshal(b, modules); err != nil {
		return err
	}
	for k, v := range *modules {
		if v.Name == "" {
			v.Name = k
		}
		m.Modules = append(m.Modules, v)
	}
	return err
}

func DetectManifests(config *project.Config) ([]Manifest, error) {
	manifests := make([]Manifest, 0, 100)
	err := filepath.WalkDir(config.BuildConfig.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".cloudseed.json") {
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			manifest := new(Manifest)
			err = json.Unmarshal(raw, manifest)
			if err != nil {
				return err
			}
			manifest.Path = path
			manifests = append(manifests, *manifest)
			return err
		}
		return nil
	})
	return manifests, err
}

type Module struct {
	Name          string        `json:"name"`
	EventSource   EventSource   `json:"eventSource"`
	Service       Service       `json:"service"`
	Networking    Networking    `json:"networking"`
	Security      Security      `json:"security"`
	Orchestration Orchestration `json:"orchestration"`
	Metadata      Metadata      `json:"metadata"`
}

type EventSource struct {
	EventSpec struct {
		Gcp google.EventarcTriggerConfig
		Aws any
	}
	TopicSpec struct {
		Gcp google.PubsubTopicConfig
		Aws any
	}
	QueueSpec struct {
		Gcp google.CloudTasksQueueConfig
		Aws any
	}
	ScheduleSpec struct {
		Gcp google.CloudSchedulerJobConfig
		Aws any
	}
}

func (c *EventSource) UnmarshalJSON(b []byte) error {
	if b == nil {
		return nil
	}
	type source struct {
		Kind    string          `json:"kind"`
		GcpSpec json.RawMessage `json:"gcp"`
		AwsSpec json.RawMessage `json:"aws"`
	}
	s := new(source)
	var err error
	if err = json.Unmarshal(b, s); err != nil {
		return err
	}
	switch s.Kind {
	case "event":
		{
			if err == nil && s.GcpSpec != nil {
				err = json.Unmarshal(s.GcpSpec, &c.EventSpec.Gcp)
			}
			if err == nil && s.AwsSpec != nil {
				err = json.Unmarshal(s.AwsSpec, &c.EventSpec.Aws)
			}
		}
	case "topic":
		{
			if err == nil && s.GcpSpec != nil {
				err = json.Unmarshal(s.GcpSpec, &c.TopicSpec.Gcp)
			}
			if err == nil && s.AwsSpec != nil {
				err = json.Unmarshal(s.AwsSpec, &c.TopicSpec.Aws)
			}
		}
	case "queue":
		{
			if err == nil && s.GcpSpec != nil {
				err = json.Unmarshal(s.GcpSpec, &c.QueueSpec.Gcp)
			}
			if err == nil && s.AwsSpec != nil {
				err = json.Unmarshal(s.AwsSpec, &c.QueueSpec.Aws)
			}
		}
	case "schedule":
		{
			if err == nil && s.GcpSpec != nil {
				err = json.Unmarshal(s.GcpSpec, &c.ScheduleSpec.Gcp)
			}
			if err == nil && s.AwsSpec != nil {
				err = json.Unmarshal(s.AwsSpec, &c.ScheduleSpec.Aws)
			}
		}
	case "http":
		fallthrough
	default:
	}
	return err
}

type Service struct {
	Function struct {
		Gcp google.Cloudfunctions2FunctionConfig `json:"gcp"`
		Aws any                                  `json:"aws"`
	} `json:"function"`
	Container struct {
		Gcp google.CloudRunServiceConfig `json:"gcp"`
		Aws any                          `json:"aws"`
	} `json:"container"`
}

func (c *Service) UnmarshalJSON(b []byte) error {
	if b == nil {
		return nil
	}
	type source struct {
		Kind    string          `json:"kind"`
		GcpSpec json.RawMessage `json:"gcp"`
		AwsSpec json.RawMessage `json:"aws"`
	}
	s := new(source)
	var err error
	if err = json.Unmarshal(b, s); err != nil {
		return err
	}
	switch s.Kind {
	case "function":
		{
			if err == nil && s.GcpSpec != nil {
				err = json.Unmarshal(s.GcpSpec, &c.Function.Gcp)
			}
			if err == nil && s.AwsSpec != nil {
				err = json.Unmarshal(s.AwsSpec, &c.Function.Aws)
			}
		}
	case "container":
		{
			if err == nil && s.GcpSpec != nil {
				err = json.Unmarshal(s.GcpSpec, &c.Container.Gcp)
			}
			if err == nil && s.AwsSpec != nil {
				err = json.Unmarshal(s.AwsSpec, &c.Container.Aws)
			}
		}
	default:
	}
	return err
}

type Networking struct {
	Internal bool `json:"internal"`
	Ingress  struct {
		Gateway struct {
			Paths map[string]map[string]struct {
				Parameters []struct {
					Name     string `json:"name"`
					In       string `json:"in"`
					Required bool   `json:"required"`
					gatewayContent
				} `json:"parameters"`
				RequestBody struct {
					gatewayContent
				} `json:"requestBody"`
				Responses map[string]struct {
					Description string `json:"description"`
					Headers     map[string]struct {
						gatewayContent
					} `json:"headers"`
					gatewayContent
				} `json:"responses"`
				Security []map[string][]string `json:"security"`
			} `json:"paths"`
			Components struct {
				SecuritySchemes map[string]struct {
					Type string `json:"type"`
					Name string `json:"name"`
					In   string `json:"in"`
				} `json:"securitySchemes"`
			} `json:"components"`
		} `json:"gateway"`
	} `json:"ingress"`
	Egress struct {
		StaticIp bool `json:"staticIp"`
	} `json:"egress"`
}

type gatewayContent struct {
	Content map[string]struct {
		Schema map[string]any `json:"schema"`
	} `json:"content"`
}

type Security struct {
	NoAuthentication bool `json:"noAuthentication"`
}

type Orchestration struct {
	Workflow struct {
		Start bool `json:"start"`
		End   bool `json:"end"`
		Input struct {
			Expression struct {
				Gcp string `json:"gcp"`
				Aws string `json:"aws"`
			} `json:"expression"`
		} `json:"input"`
		Output struct {
			Expression struct {
				Gcp string `json:"gcp"`
				Aws string `json:"aws"`
			} `json:"expression"`
		} `json:"output"`
		Next struct {
			Jump struct {
				ServiceName string `json:"serviceName"`
			} `json:"jump"`
			Condition []struct {
				Expression struct {
					Gcp string `json:"gcp"`
					Aws string `json:"aws"`
				} `json:"expression"`
				ServiceName string `json:"serviceName"`
			} `json:"condition"`
		} `json:"next"`
	} `json:"workflow"`
}

type Metadata struct {
	Metadata map[string]string `json:"metadata"`
}

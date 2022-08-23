package module

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mnahad/cloud-seed/generated/google"
	"github.com/mnahad/cloud-seed/services/config/project"
)

type Manifest struct {
	Path    string
	Modules []Module
}

func (m *Manifest) FilterModules(predicates []func(*Module) bool) [][]*Module {
	filtered := make([][]*Module, len(predicates))
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
	manifests := make([]Manifest, 0)
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

type Networking struct {
	Internal bool `json:"internal"`
	Ingress  struct {
		Gateway []struct {
			Path       string `json:"path"`
			Operations []struct {
				Verb      string `json:"verb"`
				Responses []struct {
					Code string `json:"code"`
				} `json:"responses"`
			} `json:"operations"`
		} `json:"gateway"`
	} `json:"ingress"`
	Egress struct {
		StaticIp bool `json:"staticIp"`
	} `json:"egress"`
}

type Security struct {
	NoAuthentication bool `json:"noAuthentication"`
}

type Orchestration struct {
	Workflow any `json:"workflow"`
}

type Metadata struct {
	Metadata map[string]string `json:"metadata"`
}

package module

import "github.com/mnahad/cloud-seed/generated/google"

type Manifest struct {
	Path    string
	Modules []Module
}

func (manifest *Manifest) FilterModules(predicates []func(*Module) bool) [][]*Module {
	filtered := make([][]*Module, len(predicates))
	for i := range manifest.Modules {
		module := &manifest.Modules[i]
		for j := range predicates {
			if predicates[j](module) {
				filtered[j] = append(filtered[j], module)
			}
		}
	}
	return filtered
}

type Module struct {
	Name          string
	EventSources  []EventSource
	Service       Service
	Networking    Networking
	Security      Security
	Orchestration Orchestration
}

type EventSource struct {
	Kind      string
	EventSpec struct {
		Topic    string
		Resource string
	}
	QueueSpec struct {
		Name string
	}
	ScheduleSpec struct {
		Schedule string
	}
}

type Service struct {
	Function struct {
		Gcp struct {
			Trigger struct {
				RetryOnFailure bool
			}
			Config google.CloudfunctionsFunctionConfig
		}
		Aws struct {
			Trigger struct {
				NumberOfRetries uint8
			}
			Config any
		}
	}
	Container struct {
		Gcp struct {
			Config google.CloudRunServiceConfig
		}
		Aws struct {
			Config any
		}
	}
}

type Networking struct {
	Internal bool
	Ingress  struct {
		Gateway []struct {
			Path       string
			Operations []struct {
				Verb      string
				Responses []struct {
					Code string
				}
			}
		}
	}
	Egress struct {
		StaticIp bool
	}
}

type Security struct {
	Authentication bool
}

type Orchestration struct {
	Workflow any
}

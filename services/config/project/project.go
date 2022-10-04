package project

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type Config struct {
	cloudConfig
	tfConfig
	buildConfig
	environmentConfig
	orchestrationConfig
	metadata
}

type ConfigFile struct {
	Default              Config            `json:"default"`
	EnvironmentOverrides map[string]Config `json:"environmentOverrides"`
}

func (projectConfig *ConfigFile) MergeConfig(env *string) Config {
	conf := new(Config)
	*conf = projectConfig.Default
	envConf := projectConfig.EnvironmentOverrides[*env]
	conf.cloudConfig.merge(&envConf.cloudConfig)
	conf.tfConfig.merge(&envConf.tfConfig)
	conf.buildConfig.merge(&envConf.buildConfig)
	conf.environmentConfig.merge(&envConf.environmentConfig)
	conf.orchestrationConfig.merge(&envConf.orchestrationConfig)
	conf.metadata.merge(&envConf.metadata)
	return *conf
}

func DetectConfig() (*ConfigFile, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(pwd, "cloudseed.json"))
	if err != nil {
		return nil, err
	}
	config := new(ConfigFile)
	err = json.Unmarshal(raw, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

type cloudConfig struct {
	Cloud struct {
		Gcp struct {
			Provider   google.GoogleProviderConfig `json:"provider"`
			Networking struct {
				StaticIpNetwork struct {
					AccessConnector google.VpcAccessConnectorConfig `json:"accessConnector"`
					Address         google.ComputeAddressConfig     `json:"address"`
					Network         google.ComputeNetworkConfig     `json:"network"`
					Router          google.ComputeRouterConfig      `json:"router"`
					RouterNat       google.ComputeRouterNatConfig   `json:"routerNat"`
				} `json:"staticIpNetwork"`
			} `json:"networking"`
			Security struct {
				RuntimeServiceAccount google.ServiceAccountConfig      `json:"runtimeServiceAccount"`
				SecretManagerSecret   google.SecretManagerSecretConfig `json:"secretManagerSecret"`
			} `json:"security"`
			Service struct {
				SourceCodeStorage struct {
					Bucket                     google.StorageBucketConfig              `json:"bucket"`
					ArtifactRegistryRepository google.ArtifactRegistryRepositoryConfig `json:"artifactRegistryRepository"`
					StagingBucket              google.StorageBucketConfig              `json:"stagingBucket"`
				} `json:"sourceCodeStorage"`
			} `json:"service"`
		} `json:"gcp"`
	} `json:"cloud"`
}

func (c *cloudConfig) merge(other *cloudConfig) {
	if other.Cloud.Gcp.Provider != (google.GoogleProviderConfig{}) {
		c.Cloud.Gcp.Provider = other.Cloud.Gcp.Provider
	}
	if other.Cloud.Gcp.Networking.StaticIpNetwork.AccessConnector != (google.VpcAccessConnectorConfig{}) {
		c.Cloud.Gcp.Networking.StaticIpNetwork.AccessConnector = other.Cloud.Gcp.Networking.StaticIpNetwork.AccessConnector
	}
	if other.Cloud.Gcp.Networking.StaticIpNetwork.Address != (google.ComputeAddressConfig{}) {
		c.Cloud.Gcp.Networking.StaticIpNetwork.Address = other.Cloud.Gcp.Networking.StaticIpNetwork.Address
	}
	if other.Cloud.Gcp.Networking.StaticIpNetwork.Network != (google.ComputeNetworkConfig{}) {
		c.Cloud.Gcp.Networking.StaticIpNetwork.Network = other.Cloud.Gcp.Networking.StaticIpNetwork.Network
	}
	if other.Cloud.Gcp.Networking.StaticIpNetwork.Router != (google.ComputeRouterConfig{}) {
		c.Cloud.Gcp.Networking.StaticIpNetwork.Router = other.Cloud.Gcp.Networking.StaticIpNetwork.Router
	}
	if other.Cloud.Gcp.Networking.StaticIpNetwork.RouterNat != (google.ComputeRouterNatConfig{}) {
		c.Cloud.Gcp.Networking.StaticIpNetwork.RouterNat = other.Cloud.Gcp.Networking.StaticIpNetwork.RouterNat
	}
	if other.Cloud.Gcp.Security.RuntimeServiceAccount != (google.ServiceAccountConfig{}) {
		c.Cloud.Gcp.Security.RuntimeServiceAccount = other.Cloud.Gcp.Security.RuntimeServiceAccount
	}
	if other.Cloud.Gcp.Security.SecretManagerSecret != (google.SecretManagerSecretConfig{}) {
		c.Cloud.Gcp.Security.SecretManagerSecret = other.Cloud.Gcp.Security.SecretManagerSecret
	}
	if other.Cloud.Gcp.Service.SourceCodeStorage.Bucket != (google.StorageBucketConfig{}) {
		c.Cloud.Gcp.Service.SourceCodeStorage.Bucket = other.Cloud.Gcp.Service.SourceCodeStorage.Bucket
	}
	if other.Cloud.Gcp.Service.SourceCodeStorage.ArtifactRegistryRepository != (google.ArtifactRegistryRepositoryConfig{}) {
		c.Cloud.Gcp.Service.SourceCodeStorage.ArtifactRegistryRepository =
			other.Cloud.Gcp.Service.SourceCodeStorage.ArtifactRegistryRepository
	}
	if other.Cloud.Gcp.Service.SourceCodeStorage.StagingBucket != (google.StorageBucketConfig{}) {
		c.Cloud.Gcp.Service.SourceCodeStorage.StagingBucket = other.Cloud.Gcp.Service.SourceCodeStorage.StagingBucket
	}
}

type tfConfig struct {
	TfConfig struct {
		Backend tfBackendConfig `json:"backend"`
		Cdktf   string          `json:"cdktf"`
	} `json:"tfConfig"`
}

func (c *tfConfig) merge(other *tfConfig) {
	if other.TfConfig.Backend.Gcs != (cdktf.GcsBackendProps{}) {
		c.TfConfig.Backend.Gcs = other.TfConfig.Backend.Gcs
	}
	if other.TfConfig.Backend.S3 != (cdktf.S3BackendProps{}) {
		c.TfConfig.Backend.S3 = other.TfConfig.Backend.S3
	}
	if other.TfConfig.Backend.Local != (cdktf.LocalBackendProps{}) {
		c.TfConfig.Backend.Local = other.TfConfig.Backend.Local
	}
	if len(other.TfConfig.Cdktf) > 0 {
		c.TfConfig.Cdktf = other.TfConfig.Cdktf
	}
}

type tfBackendConfig struct {
	Gcs   cdktf.GcsBackendProps
	S3    cdktf.S3BackendProps
	Local cdktf.LocalBackendProps
}

func (c *tfBackendConfig) UnmarshalJSON(b []byte) error {
	if b == nil {
		return nil
	}
	type opts struct {
		Type    string          `json:"type"`
		Options json.RawMessage `json:"options"`
	}
	o := new(opts)
	var err error
	if err = json.Unmarshal(b, o); err != nil {
		return err
	}
	if o.Options == nil {
		return err
	}
	switch o.Type {
	case "gcs":
		err = json.Unmarshal(o.Options, &c.Gcs)
	case "s3":
		err = json.Unmarshal(o.Options, &c.S3)
	case "local":
		fallthrough
	default:
		err = json.Unmarshal(o.Options, &c.Local)
	}
	return err
}

type buildConfig struct {
	BuildConfig struct {
		Dir    string `json:"dir"`
		OutDir string `json:"outDir"`
	} `json:"buildConfig"`
}

func (c *buildConfig) merge(other *buildConfig) {
	if len(other.BuildConfig.Dir) > 0 {
		c.BuildConfig.Dir = other.BuildConfig.Dir
	}
	if len(other.BuildConfig.OutDir) > 0 {
		c.BuildConfig.OutDir = other.BuildConfig.OutDir
	}
}

type environmentConfig struct {
	EnvironmentConfig struct {
		RuntimeEnvironmentVariables map[string]string `json:"runtimeEnvironmentVariables"`
		SecretVariableNames         []string          `json:"secretVariableNames"`
	} `json:"environmentConfig"`
}

func (c *environmentConfig) merge(other *environmentConfig) {
	if c.EnvironmentConfig.RuntimeEnvironmentVariables == nil {
		c.EnvironmentConfig.RuntimeEnvironmentVariables =
			make(map[string]string, len(other.EnvironmentConfig.RuntimeEnvironmentVariables))
	}
	for k, v := range other.EnvironmentConfig.RuntimeEnvironmentVariables {
		c.EnvironmentConfig.RuntimeEnvironmentVariables[k] = v
	}
	if c.EnvironmentConfig.SecretVariableNames == nil {
		c.EnvironmentConfig.SecretVariableNames =
			make([]string, len(other.EnvironmentConfig.SecretVariableNames))
	}
	for i := range other.EnvironmentConfig.SecretVariableNames {
		c.EnvironmentConfig.SecretVariableNames[i] = other.EnvironmentConfig.SecretVariableNames[i]
	}
}

type orchestrationConfig struct {
	OrchestrationConfig struct {
		Gcp struct {
			FilePath string                         `json:"filePath"`
			Config   google.WorkflowsWorkflowConfig `json:"config"`
		} `json:"gcp"`
	} `json:"orchestrationConfig"`
}

func (c *orchestrationConfig) merge(other *orchestrationConfig) {
	if len(other.OrchestrationConfig.Gcp.FilePath) > 0 {
		c.OrchestrationConfig.Gcp.FilePath = other.OrchestrationConfig.Gcp.FilePath
	}
	if other.OrchestrationConfig.Gcp.Config != (google.WorkflowsWorkflowConfig{}) {
		c.OrchestrationConfig.Gcp.Config = other.OrchestrationConfig.Gcp.Config
	}
}

type metadata struct {
	Metadata map[string]string `json:"metadata"`
}

func (c *metadata) merge(other *metadata) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]string, len(other.Metadata))
	}
	for k, v := range other.Metadata {
		c.Metadata[k] = v
	}
}

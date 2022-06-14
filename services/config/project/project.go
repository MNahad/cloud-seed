package project

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
)

type Config struct {
	cloudConfig
	tfConfig
	buildConfig
	environmentConfig
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
			Project           string `json:"project"`
			Region            string `json:"region"`
			SourceCodeStorage struct {
				Bucket google.StorageBucketConfig `json:"bucket"`
			} `json:"sourceCodeStorage"`
		} `json:"gcp"`
	} `json:"cloud"`
}

func (c *cloudConfig) merge(other *cloudConfig) {
	if other.Cloud.Gcp.Project != "" {
		c.Cloud.Gcp.Project = other.Cloud.Gcp.Project
	}
	if other.Cloud.Gcp.Region != "" {
		c.Cloud.Gcp.Region = other.Cloud.Gcp.Region
	}
	if other.Cloud.Gcp.SourceCodeStorage.Bucket != (google.StorageBucketConfig{}) {
		c.Cloud.Gcp.SourceCodeStorage.Bucket = other.Cloud.Gcp.SourceCodeStorage.Bucket
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
	if other.TfConfig.Cdktf != "" {
		c.TfConfig.Cdktf = other.TfConfig.Cdktf
	}
}

type tfBackendConfig struct {
	Gcs   cdktf.GcsBackendProps
	S3    cdktf.S3BackendProps
	Local cdktf.LocalBackendProps
}

func (c *tfBackendConfig) UnmarshalJSON(b []byte) error {
	type opts struct {
		Type    string          `json:"type"`
		Options json.RawMessage `json:"options"`
	}
	o := new(opts)
	var err error
	if err = json.Unmarshal(b, o); err != nil {
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
	if other.BuildConfig.Dir != "" {
		c.BuildConfig.Dir = other.BuildConfig.Dir
	}
	if other.BuildConfig.OutDir != "" {
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
		c.EnvironmentConfig.RuntimeEnvironmentVariables = make(map[string]string)
	}
	for k, v := range other.EnvironmentConfig.RuntimeEnvironmentVariables {
		c.EnvironmentConfig.RuntimeEnvironmentVariables[k] = v
	}
	if c.EnvironmentConfig.SecretVariableNames == nil {
		c.EnvironmentConfig.SecretVariableNames = make([]string, 0)
	}
	if other.EnvironmentConfig.SecretVariableNames != nil {
		c.EnvironmentConfig.SecretVariableNames =
			append(c.EnvironmentConfig.SecretVariableNames, other.EnvironmentConfig.SecretVariableNames...)
	}
}

type metadata struct {
	Metadata map[string]string `json:"metadata"`
}

func (c *metadata) merge(other *metadata) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]string)
	}
	for k, v := range other.Metadata {
		c.Metadata[k] = v
	}
}

package project

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type Config struct {
	cloudConfig
	tfConfig
	buildConfig
	runtimeConfig
}

type ProjectConfig struct {
	Default              Config            `json:"default"`
	EnvironmentOverrides map[string]Config `json:"environmentOverrides"`
}

func (projectConfig *ProjectConfig) MergeConfig(env *string) Config {
	conf := new(Config)
	*conf = projectConfig.Default
	envConf := projectConfig.EnvironmentOverrides[*env]
	if envConf.Cloud.Gcp.Project != "" {
		conf.Cloud.Gcp.Project = envConf.Cloud.Gcp.Project
	}
	if envConf.Cloud.Gcp.Region != "" {
		conf.Cloud.Gcp.Region = envConf.Cloud.Gcp.Region
	}
	if envConf.TfConfig.Backend.Type != "" {
		conf.TfConfig.Backend.Type = envConf.TfConfig.Backend.Type
	}
	if envConf.TfConfig.Backend.Gcs != (cdktf.GcsBackendProps{}) {
		conf.TfConfig.Backend.Gcs = envConf.TfConfig.Backend.Gcs
	}
	if envConf.TfConfig.Backend.S3 != (cdktf.S3BackendProps{}) {
		conf.TfConfig.Backend.S3 = envConf.TfConfig.Backend.S3
	}
	if envConf.TfConfig.Backend.Local != (cdktf.LocalBackendProps{}) {
		conf.TfConfig.Backend.Local = envConf.TfConfig.Backend.Local
	}
	if envConf.TfConfig.Cdktf != "" {
		conf.TfConfig.Cdktf = envConf.TfConfig.Cdktf
	}
	if envConf.BuildConfig.Dir != "" {
		conf.BuildConfig.Dir = envConf.BuildConfig.Dir
	}
	if envConf.BuildConfig.OutDir != "" {
		conf.BuildConfig.OutDir = envConf.BuildConfig.OutDir
	}
	if conf.RuntimeEnvironmentVariables == nil {
		conf.RuntimeEnvironmentVariables = make(map[string]string)
	}
	for k, v := range envConf.RuntimeEnvironmentVariables {
		conf.RuntimeEnvironmentVariables[k] = v
	}
	if conf.SecretVariableNames == nil {
		conf.SecretVariableNames = make([]string, 0)
	}
	conf.SecretVariableNames = append(conf.SecretVariableNames, envConf.SecretVariableNames...)
	return *conf
}

func DetectConfig() *ProjectConfig {
	config := new(ProjectConfig)
	pwd, err := os.Getwd()
	terminateOnErr(err)
	raw, err := ioutil.ReadFile(filepath.Join(pwd, "cloudseed.json"))
	terminateOnErr(err)
	err = json.Unmarshal(raw, config)
	terminateOnErr(err)
	return config
}

func terminateOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type cloudConfig struct {
	Cloud struct {
		Gcp struct {
			Project string `json:"project"`
			Region  string `json:"region"`
		} `json:"gcp"`
	} `json:"cloud"`
}

type tfConfig struct {
	TfConfig struct {
		Backend tfBackendConfig `json:"backend"`
		Cdktf   string          `json:"cdktf"`
	} `json:"tfConfig"`
}

type tfBackendConfig struct {
	Type  string
	Gcs   cdktf.GcsBackendProps
	S3    cdktf.S3BackendProps
	Local cdktf.LocalBackendProps
}

func (c *tfBackendConfig) UnmarshalJSON(b []byte) error {
	type S struct {
		Type    string          `json:"type"`
		Options json.RawMessage `json:"options"`
	}
	s := new(S)
	var err error
	if err = json.Unmarshal(b, s); err != nil {
		return err
	}
	c.Type = s.Type
	switch s.Type {
	case "gcs":
		err = json.Unmarshal(s.Options, &c.Gcs)
	case "s3":
		err = json.Unmarshal(s.Options, &c.S3)
	case "local":
	default:
		err = json.Unmarshal(s.Options, &c.Local)
	}
	return err
}

type buildConfig struct {
	BuildConfig struct {
		Dir    string `json:"dir"`
		OutDir string `json:"outDir"`
	} `json:"buildConfig"`
}

type runtimeConfig struct {
	RuntimeEnvironmentVariables map[string]string `json:"runtimeEnvironmentVariables"`
	SecretVariableNames         []string          `json:"secretVariableNames"`
}

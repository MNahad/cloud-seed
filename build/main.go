package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/stacks"
)

var config Config

func Build(env string) {
	getConfig()
	conf := mergeConfig(env)
	app := cdktf.NewApp(nil)
	stacks.NewGcpStack(app, conf.Cloud.Gcp.Project, stacks.GcpStackOptions{
		Project: conf.Cloud.Gcp.Project,
		Region:  conf.Cloud.Gcp.Region,
	})
	app.Synth()
}

func getConfig() {
	pwd, err := os.Getwd()
	terminateOnErr(err)
	raw, err := ioutil.ReadFile(filepath.Join(pwd, "cloudseed.json"))
	terminateOnErr(err)
	err = json.Unmarshal(raw, &config)
	terminateOnErr(err)
}

func mergeConfig(env string) BaseConfig {
	conf := new(BaseConfig)
	*conf = config.Default
	envConf := config.EnvironmentOverrides[env]
	if envConf.Cloud.Gcp.Project != "" {
		conf.Cloud.Gcp.Project = envConf.Cloud.Gcp.Project
	}
	if envConf.Cloud.Gcp.Region != "" {
		conf.Cloud.Gcp.Region = envConf.Cloud.Gcp.Region
	}
	if envConf.TfConfig.Backend.Type != "" {
		conf.TfConfig.Backend.Type = envConf.TfConfig.Backend.Type
	}
	if envConf.TfConfig.Backend.BackendOptions != nil {
		conf.TfConfig.Backend.BackendOptions = envConf.TfConfig.Backend.BackendOptions
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

func terminateOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type BaseConfig struct {
	Cloud struct {
		Gcp struct {
			Project string `json:"project"`
			Region  string `json:"region"`
		} `json:"gcp"`
	} `json:"cloud"`
	TfConfig struct {
		Backend struct {
			Type           string      `json:"type"`
			BackendOptions interface{} `json:"backendOptions"`
		} `json:"backend"`
		Cdktf string `json:"cdktf"`
	} `json:"tfConfig"`
	BuildConfig struct {
		Dir    string `json:"dir"`
		OutDir string `json:"outDir"`
	} `json:"buildConfig"`
	RuntimeEnvironmentVariables map[string]string `json:"runtimeEnvironmentVariables"`
	SecretVariableNames         []string          `json:"secretVariableNames"`
}

type Config struct {
	Default              BaseConfig            `json:"default"`
	EnvironmentOverrides map[string]BaseConfig `json:"environmentOverrides"`
}

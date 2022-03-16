package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

var workspaceConfig Config

func Build(env string) {
	getWorkspaceConfig()
	app := cdktf.NewApp(nil)
	app.Synth()
}

func getWorkspaceConfig() {
	pwd, err := os.Getwd()
	handleErr(err)
	raw, err := ioutil.ReadFile(filepath.Join(pwd, "cloudseed.json"))
	handleErr(err)
	err = json.Unmarshal(raw, &workspaceConfig)
	handleErr(err)
}

func handleErr(err error) {
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
	} `json:"tfConfig"`
	BuildConfig struct {
		Dir    string `json:"dir"`
		OutDir string `json:"outDir"`
	} `json:"buildConfig"`
	EnvironmentVariables map[string]string `json:"environmentVariables"`
	SecretVariableNames  []string          `json:"secretVariableNames"`
}

type Config struct {
	Default              BaseConfig            `json:"default"`
	EnvironmentOverrides map[string]BaseConfig `json:"environmentOverrides"`
}

package build

import (
	"os"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder"
)

func Build(env *string) (*project.Config, *cdktf.App) {
	projectConf, err := project.DetectConfig()
	if err != nil {
		os.Exit(1)
	}
	conf := projectConf.MergeConfig(env)
	app := cdktf.NewApp(&cdktf.AppOptions{Outdir: &conf.BuildConfig.OutDir})
	manifests, err := module.DetectManifests(&conf)
	if err != nil {
		os.Exit(1)
	}
	stackbuilder.NewStack(&app, "CloudSeed", env, &manifests, &conf)
	app.Synth()
	return &conf, &app
}

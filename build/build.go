package build

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder"
)

func Build(env *string) (*project.Config, *cdktf.App) {
	conf := project.DetectConfig().MergeConfig(env)
	app := cdktf.NewApp(&cdktf.AppOptions{Outdir: &conf.BuildConfig.OutDir})
	manifests := module.DetectManifests(&conf)
	stackbuilder.NewStack(&app, "CloudSeed", env, &manifests, &conf)
	app.Synth()
	return &conf, &app
}

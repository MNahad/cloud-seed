package build

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/artefactgenerator"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder"
)

func Build(env string, dir string) (*project.Config, *cdktf.App) {
	projectConf, err := project.DetectConfig(dir)
	if err != nil {
		panic(err)
	}
	conf := projectConf.MergeConfig(env)
	manifests, err := module.DetectManifests(&conf)
	if err != nil {
		panic(err)
	}
	for i := range manifests {
		err = artefactgenerator.Generate(&manifests[i], &conf)
		if err != nil {
			panic(err)
		}
	}
	app := cdktf.NewApp(&cdktf.AppOptions{Outdir: &conf.BuildConfig.OutDir})
	stackbuilder.NewStack(
		&app,
		"CloudSeed",
		&stackbuilder.StackConfig{Environment: &env, Manifests: manifests, Config: &conf},
	)
	app.Synth()
	return &conf, &app
}

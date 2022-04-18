package build

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/services/config/project"
	"github.com/mnahad/cloud-seed/services/stackbuilder"
)

func Build(env *string) (*project.Config, *cdktf.App) {
	conf := project.MergeConfig(project.DetectConfig(), env)
	app := cdktf.NewApp(nil)
	stackbuilder.NewStack(&app, "CloudSeed", env, conf)
	app.Synth()
	return conf, &app
}

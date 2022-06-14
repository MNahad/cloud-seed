package artefactgenerator

import (
	"github.com/mnahad/cloud-seed/services/artefactgenerator/gcp"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

func Generate(manifest *module.Manifest, config *project.Config) error {
	err := gcp.GenerateArtefacts(manifest, config)
	if err != nil {
		return err
	}
	return nil
}

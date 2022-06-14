package gcp

import (
	"path/filepath"

	"github.com/mnahad/cloud-seed/services/archiver"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

var outPrefix [2]string = [...]string{"artefacts", "gcp"}

type artefact string

const (
	FunctionArtefact  artefact = "functions"
	ContainerArtefact artefact = "images"
)

func GetArtefactPrefix(artefactType artefact) string {
	prefix := ""
	for i := range outPrefix {
		prefix = filepath.Join(prefix, outPrefix[i])
	}
	switch artefactType {
	case FunctionArtefact:
		fallthrough
	case ContainerArtefact:
		prefix = filepath.Join(prefix, string(artefactType))
	}
	return prefix
}

func GenerateArtefacts(manifest *module.Manifest, config *project.Config) error {
	predicates := []func(*module.Module) bool{
		func(m *module.Module) bool {
			return m.Service.Function.Gcp != (module.Service{}.Function.Gcp)
		},
	}
	gcpModules := manifest.FilterModules(predicates)
	for i := range gcpModules[0] {
		err := generateGcpFunctionArtefact(gcpModules[0][i], manifest, config)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateGcpFunctionArtefact(module *module.Module, manifest *module.Manifest, config *project.Config) error {
	archivePath := filepath.Join(config.BuildConfig.OutDir, GetArtefactPrefix(FunctionArtefact), module.Name) + ".zip"
	err := archiver.Archive(archivePath, filepath.Dir(manifest.Path))
	if err != nil {
		return err
	}
	return nil
}

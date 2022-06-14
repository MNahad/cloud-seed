package gcp

import (
	"path/filepath"

	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
	"github.com/mnahad/cloud-seed/generated/google_beta"
	gcpArtefactGenerator "github.com/mnahad/cloud-seed/services/artefactgenerator/gcp"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

func newFunction(
	scope *cdktf.TerraformStack,
	config *module.Module,
	supportInfrastructure *supportInfrastructure,
	manifest *module.Manifest,
	options *project.Config,
) *google_beta.GoogleCloudfunctions2Function {
	archivePath, _ := filepath.Abs(filepath.Join(
		options.BuildConfig.OutDir,
		gcpArtefactGenerator.GetArtefactPrefix(gcpArtefactGenerator.FunctionArtefact),
		config.Name,
	) + ".zip")

	archiveObject := google.NewStorageBucketObject(
		*scope,
		jsii.String(config.Name+"-sourceArchive"),
		&google.StorageBucketObjectConfig{
			Bucket: (*supportInfrastructure.function.archiveBucket).Name(),
			Name:   jsii.String(config.Name + "-source"),
			Source: &archivePath,
		},
	)
	functionConfig := new(google_beta.GoogleCloudfunctions2FunctionConfig)
	(*functionConfig) = config.Service.Function.Gcp
	if functionConfig.Name == nil {
		functionConfig.Name = &config.Name
	}
	if functionConfig.BuildConfig == nil {
		functionConfig.BuildConfig = &google_beta.GoogleCloudfunctions2FunctionBuildConfig{}
	}
	if functionConfig.ServiceConfig == nil {
		functionConfig.ServiceConfig = &google_beta.GoogleCloudfunctions2FunctionServiceConfig{}
	}
	if functionConfig.BuildConfig.EntryPoint == nil {
		functionConfig.BuildConfig.EntryPoint = &config.Name
	}
	if functionConfig.BuildConfig.Source == nil {
		functionConfig.BuildConfig.Source = &google_beta.GoogleCloudfunctions2FunctionBuildConfigSource{}
	}
	if functionConfig.BuildConfig.Source.StorageSource == nil {
		functionConfig.BuildConfig.Source.StorageSource =
			&google_beta.GoogleCloudfunctions2FunctionBuildConfigSourceStorageSource{}
	}
	if functionConfig.BuildConfig.Source.StorageSource.Bucket == nil {
		functionConfig.BuildConfig.Source.StorageSource.Bucket = archiveObject.Bucket()
	}
	if functionConfig.BuildConfig.Source.StorageSource.Object == nil {
		functionConfig.BuildConfig.Source.StorageSource.Object = archiveObject.Name()
	}
	if functionConfig.ServiceConfig.EnvironmentVariables == nil {
		functionConfig.ServiceConfig.EnvironmentVariables = &map[string]*string{}
	}
	for k, v := range options.EnvironmentConfig.RuntimeEnvironmentVariables {
		if (*functionConfig.ServiceConfig.EnvironmentVariables)[k] == nil {
			(*functionConfig.ServiceConfig.EnvironmentVariables)[k] = &v
		}
	}
	function := google_beta.NewGoogleCloudfunctions2Function(*scope, &config.Name, functionConfig)
	return &function
}

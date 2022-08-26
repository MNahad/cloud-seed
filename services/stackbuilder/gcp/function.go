package gcp

import (
	"path/filepath"

	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/mnahad/cloud-seed/generated/google"
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
) *google.Cloudfunctions2Function {
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
			Name:   jsii.String(config.Name + "-source-" + *cdktf.Fn_Urlencode(cdktf.Fn_Filebase64sha512(&archivePath)) + ".zip"),
			Source: &archivePath,
		},
	)
	functionConfig := new(google.Cloudfunctions2FunctionConfig)
	(*functionConfig) = config.Service.Function.Gcp
	if functionConfig.Name == nil {
		functionConfig.Name = &config.Name
	}
	if functionConfig.Location == nil {
		functionConfig.Location = &options.Cloud.Gcp.Region
	}
	if functionConfig.BuildConfig == nil {
		functionConfig.BuildConfig = &google.Cloudfunctions2FunctionBuildConfig{}
	}
	if functionConfig.ServiceConfig == nil {
		functionConfig.ServiceConfig = &google.Cloudfunctions2FunctionServiceConfig{}
	}
	if functionConfig.BuildConfig.EntryPoint == nil {
		functionConfig.BuildConfig.EntryPoint = &config.Name
	}
	if functionConfig.BuildConfig.Source == nil {
		functionConfig.BuildConfig.Source = &google.Cloudfunctions2FunctionBuildConfigSource{}
	}
	if functionConfig.BuildConfig.Source.StorageSource == nil {
		functionConfig.BuildConfig.Source.StorageSource =
			&google.Cloudfunctions2FunctionBuildConfigSourceStorageSource{}
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
	for k := range options.EnvironmentConfig.RuntimeEnvironmentVariables {
		if (*functionConfig.ServiceConfig.EnvironmentVariables)[k] == nil {
			(*functionConfig.ServiceConfig.EnvironmentVariables)[k] = jsii.String(
				options.EnvironmentConfig.RuntimeEnvironmentVariables[k],
			)
		}
	}
	function := google.NewCloudfunctions2Function(*scope, &config.Name, functionConfig)
	return &function
}

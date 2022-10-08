package service

import (
	"path/filepath"

	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	gcpartefactgenerator "github.com/mnahad/cloud-seed/services/artefactgenerator/gcp"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

func (s *service) NewFunction(
	scope *cdktf.TerraformStack,
	config *module.Module,
	options *project.Config,
) *google.Cloudfunctions2Function {
	if s.archiveBucket == nil {
		s.archiveBucket = newArchiveBucket(scope, options)
	}
	archivePath, err := filepath.Abs(filepath.Join(
		options.Path,
		options.BuildConfig.OutDir,
		gcpartefactgenerator.GetArtefactPrefix(gcpartefactgenerator.FunctionArtefact),
		config.Name,
	) + ".zip")
	if err != nil {
		panic(err)
	}
	archiveObject := google.NewStorageBucketObject(
		*scope,
		jsii.String(config.Name+"-sourceArchive"),
		&google.StorageBucketObjectConfig{
			Bucket: (*s.archiveBucket).Name(),
			Name:   jsii.String(config.Name + ":" + *cdktf.Fn_Urlencode(cdktf.Fn_Filebase64sha512(&archivePath)) + ".zip"),
			Source: &archivePath,
		},
	)
	functionConfig := new(google.Cloudfunctions2FunctionConfig)
	(*functionConfig) = config.Service.Function.Gcp
	if functionConfig.Name == nil {
		functionConfig.Name = &config.Name
	}
	if functionConfig.Location == nil {
		functionConfig.Location = options.Cloud.Gcp.Provider.Region
	}
	if functionConfig.BuildConfig == nil {
		functionConfig.BuildConfig = &google.Cloudfunctions2FunctionBuildConfig{}
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
	if functionConfig.ServiceConfig == nil {
		functionConfig.ServiceConfig = &google.Cloudfunctions2FunctionServiceConfig{}
	}
	if functionConfig.ServiceConfig.EnvironmentVariables == nil {
		envVars := make(map[string]*string, len(options.EnvironmentConfig.RuntimeEnvironmentVariables))
		functionConfig.ServiceConfig.EnvironmentVariables = &envVars
	}
	for k := range options.EnvironmentConfig.RuntimeEnvironmentVariables {
		if (*functionConfig.ServiceConfig.EnvironmentVariables)[k] == nil {
			(*functionConfig.ServiceConfig.EnvironmentVariables)[k] = jsii.String(
				options.EnvironmentConfig.RuntimeEnvironmentVariables[k],
			)
		}
	}
	if functionConfig.ServiceConfig.AvailableMemory == nil {
		functionConfig.ServiceConfig.AvailableMemory = jsii.String("256M")
	}
	if functionConfig.ServiceConfig.MaxInstanceCount == nil {
		functionConfig.ServiceConfig.MaxInstanceCount = jsii.Number(100)
	}
	if functionConfig.ServiceConfig.TimeoutSeconds == nil {
		functionConfig.ServiceConfig.TimeoutSeconds = jsii.Number(60)
	}
	if functionConfig.Lifecycle == nil {
		functionConfig.Lifecycle = &cdktf.TerraformResourceLifecycle{}
	}
	if functionConfig.Lifecycle.IgnoreChanges == nil {
		functionConfig.Lifecycle.IgnoreChanges =
			&[]*string{jsii.String("build_config[0].source[0].storage_source[0].generation")}
	}
	function := google.NewCloudfunctions2Function(*scope, functionConfig.Name, functionConfig)
	return &function
}

func newArchiveBucket(scope *cdktf.TerraformStack, options *project.Config) *google.StorageBucket {
	archiveBucketConfig := new(google.StorageBucketConfig)
	(*archiveBucketConfig) = options.Cloud.Gcp.Service.SourceCodeStorage.Bucket
	if archiveBucketConfig.Name == nil {
		archiveBucketConfig.Name = jsii.String(*options.Cloud.Gcp.Provider.Project + "-functions-sources")
	}
	if archiveBucketConfig.Location == nil {
		archiveBucketConfig.Location = options.Cloud.Gcp.Provider.Region
	}
	if archiveBucketConfig.UniformBucketLevelAccess == nil {
		archiveBucketConfig.UniformBucketLevelAccess = jsii.Bool(true)
	}
	archiveBucket := google.NewStorageBucket(*scope, jsii.String("ArchiveBucket"), archiveBucketConfig)
	return &archiveBucket
}

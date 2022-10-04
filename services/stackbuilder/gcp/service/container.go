package service

import (
	"path/filepath"

	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
	"github.com/hashicorp/cdktf-provider-null-go/null/v2"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	gcpArtefactGenerator "github.com/mnahad/cloud-seed/services/artefactgenerator/gcp"
	"github.com/mnahad/cloud-seed/services/config/module"
	"github.com/mnahad/cloud-seed/services/config/project"
)

var imageRepository *google.ArtifactRegistryRepository
var stagingBucket *google.StorageBucket

func NewRunService(scope *cdktf.TerraformStack, config *module.Module, options *project.Config) *google.CloudRunService {
	if imageRepository == nil {
		imageRepository = newArtifactRegistryRepository(scope, config, options)
	}
	if stagingBucket == nil {
		stagingBucket = newStagingBucket(scope, options)
	}
	archivePath, err := filepath.Abs(filepath.Join(
		options.BuildConfig.OutDir,
		gcpArtefactGenerator.GetArtefactPrefix(gcpArtefactGenerator.ContainerArtefact),
		config.Name,
	) + ".zip")
	if err != nil {
		panic(err)
	}
	image, imageUrl := newImage(scope, &config.Name, &archivePath, options)
	runConfig := new(google.CloudRunServiceConfig)
	(*runConfig) = config.Service.Container.Gcp
	if runConfig.Name == nil {
		runConfig.Name = &config.Name
	}
	if runConfig.Location == nil {
		runConfig.Location = options.Cloud.Gcp.Provider.Region
	}
	if runConfig.AutogenerateRevisionName == nil {
		runConfig.AutogenerateRevisionName = jsii.Bool(true)
	}
	if runConfig.Template == nil {
		runConfig.Template = &google.CloudRunServiceTemplate{}
	}
	if runConfig.Template.Spec == nil {
		runConfig.Template.Spec = &google.CloudRunServiceTemplateSpec{}
	}
	env := make(
		[]*google.CloudRunServiceTemplateSpecContainersEnv,
		0,
		len(options.EnvironmentConfig.RuntimeEnvironmentVariables),
	)
	for k, v := range options.EnvironmentConfig.RuntimeEnvironmentVariables {
		env = append(env, &google.CloudRunServiceTemplateSpecContainersEnv{Name: jsii.String(k), Value: jsii.String(v)})
	}
	if containers, ok := runConfig.Template.Spec.Containers.([]any); ok {
		if container, ok := containers[0].(map[string]any); ok {
			if _, ok := container["image"].(string); !ok {
				container["image"] = imageUrl
			}
			if existingEnvs, ok := container["env"].([]any); ok {
				existingEnvNames := make(map[string]bool, len(existingEnvs))
				for i := range existingEnvs {
					if e, ok := existingEnvs[i].(map[string]any); ok {
						if name, ok := e["name"].(string); ok {
							existingEnvNames[name] = true
						}
					}
				}
				targetEnvs := make([]any, 0, len(env))
				for i := range env {
					if !existingEnvNames[*env[i].Name] {
						targetEnvs = append(targetEnvs, map[string]any{"name": *env[i].Name, "value": *env[i].Value})
					}
				}
				if len(targetEnvs) > 0 {
					existingEnvs = append(existingEnvs, targetEnvs...)
					container["env"] = existingEnvs
				}
			} else {
				container["env"] = env
			}
		}
	} else {
		runConfig.Template.Spec.Containers = &[]*google.CloudRunServiceTemplateSpecContainers{{Image: &imageUrl, Env: &env}}
	}
	if runConfig.Lifecycle == nil {
		runConfig.Lifecycle = &cdktf.TerraformResourceLifecycle{}
	}
	if runConfig.Lifecycle.IgnoreChanges == nil {
		runConfig.Lifecycle.IgnoreChanges = &[]*string{
			jsii.String("metadata[0].annotations[\"run.googleapis.com/operation-id\"]"),
			jsii.String("template[0].metadata[0].labels[\"run.googleapis.com/startupProbeType\"]"),
		}
	}
	if runConfig.DependsOn == nil {
		runConfig.DependsOn = &[]cdktf.ITerraformDependable{image}
	}
	runService := google.NewCloudRunService(*scope, runConfig.Name, runConfig)
	return &runService
}

func newArtifactRegistryRepository(
	scope *cdktf.TerraformStack,
	config *module.Module,
	options *project.Config,
) *google.ArtifactRegistryRepository {
	repoConfig := new(google.ArtifactRegistryRepositoryConfig)
	(*repoConfig) = options.Cloud.Gcp.Service.SourceCodeStorage.ArtifactRegistryRepository
	if repoConfig.RepositoryId == nil {
		repoConfig.RepositoryId = jsii.String("containers-sources")
	}
	if repoConfig.Location == nil {
		repoConfig.Location = options.Cloud.Gcp.Provider.Region
	}
	if repoConfig.Format == nil {
		repoConfig.Format = jsii.String("DOCKER")
	}
	repo := google.NewArtifactRegistryRepository(*scope, jsii.String("ImageRepository"), repoConfig)
	return &repo
}

func newStagingBucket(scope *cdktf.TerraformStack, options *project.Config) *google.StorageBucket {
	stagingBucketConfig := new(google.StorageBucketConfig)
	(*stagingBucketConfig) = options.Cloud.Gcp.Service.SourceCodeStorage.StagingBucket
	if stagingBucketConfig.Name == nil {
		stagingBucketConfig.Name = jsii.String(*options.Cloud.Gcp.Provider.Project + "_cloudbuild")
	}
	if stagingBucketConfig.Location == nil {
		stagingBucketConfig.Location = options.Cloud.Gcp.Provider.Region
	}
	if stagingBucketConfig.UniformBucketLevelAccess == nil {
		stagingBucketConfig.UniformBucketLevelAccess = jsii.Bool(true)
	}
	if stagingBucketConfig.ForceDestroy == nil {
		stagingBucketConfig.ForceDestroy = jsii.Bool(true)
	}
	stagingBucket := google.NewStorageBucket(*scope, jsii.String("StagingBucket"), stagingBucketConfig)
	return &stagingBucket
}

func getImageUrl(project *string, region *string, repositoryName *string, sourceName *string, tag *string) string {
	return *region + "-docker.pkg.dev/" + *project + "/" + *repositoryName + "/" + *sourceName + ":" + *tag
}

func newImage(
	scope *cdktf.TerraformStack,
	archiveName *string,
	archivePath *string,
	options *project.Config,
) (null.Resource, string) {
	archiveHash := cdktf.Fn_Filesha1(archivePath)
	imageUrl := getImageUrl(
		options.Cloud.Gcp.Provider.Project,
		options.Cloud.Gcp.Provider.Region,
		(*imageRepository).Name(),
		archiveName,
		archiveHash,
	)
	nullProvisionerCommand := jsii.Strings(
		"gcloud",
		"builds",
		"submit",
		"'"+*archivePath+"'",
		"--project "+*options.Cloud.Gcp.Provider.Project,
		"--region "+*options.Cloud.Gcp.Provider.Region,
		"--tag "+imageUrl,
		"--gcs-source-staging-dir gs://"+*(*stagingBucket).Id()+"/source",
		"--gcs-log-dir gs://"+*(*stagingBucket).Id()+"/log",
		"--suppress-logs",
	)
	nullResource := null.NewResource(*scope, jsii.String(*archiveName+"sourceImage"), &null.ResourceConfig{
		Triggers: &map[string]*string{
			"tag": archiveHash,
		},
		Provisioners: &[]any{
			&cdktf.LocalExecProvisioner{
				Type:    jsii.String("local-exec"),
				Command: cdktf.Fn_Join(jsii.String(" "), nullProvisionerCommand),
			},
		},
	})
	return nullResource, imageUrl
}

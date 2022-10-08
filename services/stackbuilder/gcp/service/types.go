package service

import (
	"github.com/hashicorp/cdktf-provider-google-go/google/v2"
)

type service struct {
	archiveBucket   *google.StorageBucket
	imageRepository *google.ArtifactRegistryRepository
	stagingBucket   *google.StorageBucket
}

func NewService() *service {
	return &service{}
}

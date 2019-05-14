package services

import (
	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"testing"
)

func TestArtifactoryRepositoryKey(t *testing.T) {
	fieldsArtifactory := &types.ArtifactoryInformation{
		Tenant:      utils.ArtifactoryTestTenant,
		ProjectName: utils.ArtifactoryTestProjectName,
		Stages:      []string{utils.ArtifactoryStageScratch, utils.ArtifactoryStageStaging, utils.ArtifactoryStageStable},
		Location:    utils.ArtifactoryLocationIntranet,
		Description: utils.ArtifactoryDescription,
	}

	stage := utils.ArtifactoryStageScratch

	keyToTest := artifactoryRepositoryKey(fieldsArtifactory, stage)

	keyShouldBe := utils.XrayTestRepoScratch

	if keyToTest != keyShouldBe {
		t.Errorf("Artifactory Repository Key was incorrect, got: %v, want: %v.", keyToTest, keyShouldBe)
	}

}

func TestPermissionNameRO(t *testing.T) {
	fieldsArtifactory := &types.ArtifactoryInformation{
		Tenant:      utils.ArtifactoryTestTenant,
		ProjectName: utils.ArtifactoryTestProjectName,
		Stages:      []string{utils.ArtifactoryStageScratch, utils.ArtifactoryStageStaging, utils.ArtifactoryStageStable},
		Location:    utils.ArtifactoryLocationIntranet,
		Description: utils.ArtifactoryDescription,
	}

	permission := "ro"

	permissionToTest := permissionName(permission, fieldsArtifactory)

	if permissionToTest != utils.ArtifactoryTestPermissionRO {
		t.Errorf("Artifactory permission name was incorrect, got: %v, want: %v.", permissionToTest, utils.ArtifactoryTestPermissionRO)
	}

}

func TestPermissionNameRW(t *testing.T) {
	fieldsArtifactory := &types.ArtifactoryInformation{
		Tenant:      utils.ArtifactoryTestTenant,
		ProjectName: utils.ArtifactoryTestProjectName,
		Stages:      []string{utils.ArtifactoryStageScratch, utils.ArtifactoryStageStaging, utils.ArtifactoryStageStable},
		Location:    utils.ArtifactoryLocationIntranet,
		Description: utils.ArtifactoryDescription,
	}

	permission := "rw"

	permissionToTest := permissionName(permission, fieldsArtifactory)

	if permissionToTest != utils.ArtifactoryTestPermissionRW {
		t.Errorf("Artifactory permission name was incorrect, got: %v, want: %v.", permissionToTest, utils.ArtifactoryTestPermissionRW)
	}

}

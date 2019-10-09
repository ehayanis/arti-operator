package services

import (
	"fmt"
	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	kubiv1 "github.com/ca-gip/kubi/pkg/apis/ca-gip/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_generateArtifactoryFields_EmptyProject(t *testing.T) {
	config := types.ArtifactoryOperatorConfig{
		ClusterLocation: utils.ArtifactoryLocationIntranet,
	}
	fmt.Println("TOTOTOTOTO")

	service, _ := NewProjectService(&config, nil, nil, nil, nil)

	input := &kubiv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.ArtifactoryTestProjectMetaNameDevelopment,
		},
		Spec: kubiv1.ProjectSpec{
			Tenant:      utils.ArtifactoryTestTenant,
			Environment: utils.ArtifactoryProjectSpecEnvironment,
			Project:     "",
			Stages: []string{
				utils.ArtifactoryStageScratch,
			},
		},
	}

	result, err := service.generateArtifactoryFields(input)

	assert.Nil(t, result, "Result returned, expected none.")
	assert.Error(t, err, "Expected error.")
}

func Test_generateArtifactoryFields_TenantDuplicatedInProjectName(t *testing.T) {
	config := types.ArtifactoryOperatorConfig{
		ClusterLocation: utils.ArtifactoryLocationIntranet,
	}

	service, _ := NewProjectService(&config, nil, nil, nil, nil)

	input := &kubiv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.ArtifactoryTestProjectMetaNameDevelopment,
		},
		Spec: kubiv1.ProjectSpec{
			Tenant:      utils.ArtifactoryTestTenant,
			Environment: utils.ArtifactoryProjectSpecEnvironment,
			Project:     "aug-e4",
			Stages: []string{
				utils.ArtifactoryStageScratch,
			},
		},
	}

	expected := &types.ArtifactoryInformation{
		Tenant:      utils.ArtifactoryTestTenant,
		ProjectName: utils.ArtifactoryTestProjectName,
		Stages: []string{
			utils.ArtifactoryStageScratch,
		},
		Location:    utils.ArtifactoryLocationIntranet,
		Description: utils.ArtifactoryDescription,
		Environment: utils.ArtifactoryProjectSpecEnvironment,
	}

	result, err := service.generateArtifactoryFields(input)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected result %v, got %v.", expected, result)
	}

}

func Test_generateArtifactoryFields_TenantMissing(t *testing.T) {
	config := types.ArtifactoryOperatorConfig{
		ClusterLocation: utils.ArtifactoryLocationIntranet,
	}

	service, _ := NewProjectService(&config, nil, nil, nil, nil)

	input := &kubiv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: "defaulttenant-e4-development",
		},
		Spec: kubiv1.ProjectSpec{
			Tenant:      "",
			Environment: utils.ArtifactoryProjectSpecEnvironment,
			Project:     "defaulttenant-e4",
			Stages: []string{
				utils.ArtifactoryStageScratch,
			},
		},
	}

	result, err := service.generateArtifactoryFields(input)

	assert.Error(t, err)
	assert.Nil(t, result)
}
func Test_generateArtifactoryFields_CleanProjectName(t *testing.T) {
	config := types.ArtifactoryOperatorConfig{
		ClusterLocation: utils.ArtifactoryLocationIntranet,
	}

	service, _ := NewProjectService(&config, nil, nil, nil, nil)

	input := &kubiv1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.ArtifactoryTestProjectMetaNameDevelopment,
		},
		Spec: kubiv1.ProjectSpec{
			Tenant:      utils.ArtifactoryTestTenant,
			Environment: utils.ArtifactoryProjectSpecEnvironment,
			Project:     utils.ArtifactoryTestProjectName,
			Stages: []string{
				utils.ArtifactoryStageScratch,
			},
		},
	}

	expected := &types.ArtifactoryInformation{
		Tenant:      utils.ArtifactoryTestTenant,
		ProjectName: utils.ArtifactoryTestProjectName,
		Stages: []string{
			utils.ArtifactoryStageScratch,
		},
		Location:    utils.ArtifactoryLocationIntranet,
		Description: utils.ArtifactoryDescription,
		Environment: utils.ArtifactoryProjectSpecEnvironment,
	}

	result, err := service.generateArtifactoryFields(input)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected result %v, got %v.", expected, result)
	}
}

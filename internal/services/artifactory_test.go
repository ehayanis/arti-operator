package services

import (
	"context"
	"errors"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type mockArtifactorySecurity struct {
	getGroupWithError          bool
	getGroupWithStatusNotFound bool
}

func NewMockArtifactorySecurityWithStatusNotFound() ArtifactorySecurity {
	return &mockArtifactorySecurity{
		getGroupWithError:          false,
		getGroupWithStatusNotFound: true,
	}
}

func NewMockArtifactorySecurityWithError() ArtifactorySecurity {
	return &mockArtifactorySecurity{
		getGroupWithError:          true,
		getGroupWithStatusNotFound: false,
	}
}

func NewMockArtifactorySecurityWithErrorAndNotFound() ArtifactorySecurity {
	return &mockArtifactorySecurity{
		getGroupWithError:          true,
		getGroupWithStatusNotFound: true,
	}
}

func NewMockArtifactorySecurityWithoutError() ArtifactorySecurity {
	return &mockArtifactorySecurity{
		getGroupWithError:          false,
		getGroupWithStatusNotFound: false,
	}
}

func (m *mockArtifactorySecurity) GetGroup(ctx context.Context, groupName string) (*artifactory.Group, *http.Response, error) {

	if m.getGroupWithStatusNotFound == false && m.getGroupWithError == true {
		return nil, nil, errors.New("dummy error")
	}

	if m.getGroupWithStatusNotFound == true && m.getGroupWithError == false {
		return nil, &http.Response{StatusCode: http.StatusNotFound}, nil
	}

	if m.getGroupWithStatusNotFound == true && m.getGroupWithError == true {
		return nil, &http.Response{StatusCode: http.StatusNotFound}, errors.New("dummy error")
	}

	group := &artifactory.Group{
		Name: artifactory.String("fake group created"),
	}

	resp := &http.Response{
		StatusCode: 200,
	}

	return group, resp, nil
}

func (m *mockArtifactorySecurity) CreateOrReplaceGroup(ctx context.Context, groupName string, group *artifactory.Group) (*http.Response, error) {

	resp := &http.Response{
		StatusCode: http.StatusCreated,
	}

	return resp, nil
}

func prepareArtifactoryService(fakeError bool, groupNotFound bool) *ArtifactoryService {

	if fakeError == true && groupNotFound == true {
		return &ArtifactoryService{
			logger:              zerolog.Logger{},
			artifactoryClient:   &artifactory.Client{},
			artifactoryUrl:      "https://dummy-arti.example.fr",
			clusterDNSSubdomain: "devops-test",
			Security:            NewMockArtifactorySecurityWithErrorAndNotFound(),
		}
	}

	if fakeError == true && groupNotFound == false {
		return &ArtifactoryService{
			logger:              zerolog.Logger{},
			artifactoryClient:   &artifactory.Client{},
			artifactoryUrl:      "https://dummy-arti.example.fr",
			clusterDNSSubdomain: "devops-test",
			Security:            NewMockArtifactorySecurityWithError(),
		}
	}

	if groupNotFound == true && fakeError == false {
		return &ArtifactoryService{
			logger:              zerolog.Logger{},
			artifactoryClient:   &artifactory.Client{},
			artifactoryUrl:      "https://dummy-arti.example.fr",
			clusterDNSSubdomain: "devops-test",
			Security:            NewMockArtifactorySecurityWithStatusNotFound(),
		}
	}

	return &ArtifactoryService{
		logger:              zerolog.Logger{},
		artifactoryClient:   &artifactory.Client{},
		artifactoryUrl:      "https://dummy-arti.example.fr",
		clusterDNSSubdomain: "devops-test",
		Security:            NewMockArtifactorySecurityWithoutError(),
	}
}

func TestCreateArtifactoryGroup(t *testing.T) {
	t.Run("Error retrieving the security group should be catch and returned", func(t *testing.T) {
		//Prepare
		artifactoryService := prepareArtifactoryService(true, false)
		group := &artifactory.Group{
			Name:            nil,
			Description:     nil,
			AutoJoin:        nil,
			AdminPrivileges: nil,
			Realm:           nil,
			RealmAttributes: nil,
		}

		groupName := "dl_kub_entity_ops"

		expectedError := errors.New("dummy error")

		//Test
		_, err := artifactoryService.createArtifactoryGroup(group, groupName)

		//Assert
		assert.Equal(t, expectedError, err)
	})

	t.Run("Error retrieving the security group and resp status 404 should led to create a new security group", func(t *testing.T) {
		//Prepare
		artifactoryService := prepareArtifactoryService(true, true)
		group := &artifactory.Group{
			Name:            nil,
			Description:     nil,
			AutoJoin:        nil,
			AdminPrivileges: nil,
			Realm:           nil,
			RealmAttributes: nil,
		}

		groupName := "dl_kub_entity_ops"

		expectedResp := &http.Response{StatusCode: http.StatusCreated}

		//Test
		resp, _ := artifactoryService.createArtifactoryGroup(group, groupName)

		//Assert
		assert.Equal(t, expectedResp.StatusCode, resp.StatusCode)
	})

	t.Run("if security group is not found, then it must create it and return a 201 created", func(t *testing.T) {
		//Prepare
		artifactoryService := prepareArtifactoryService(false, true)
		group := &artifactory.Group{
			Name:            nil,
			Description:     nil,
			AutoJoin:        nil,
			AdminPrivileges: nil,
			Realm:           nil,
			RealmAttributes: nil,
		}

		groupName := "dl_kub_entity_ops"

		expectedResp := &http.Response{StatusCode: http.StatusCreated}

		//Test
		resp, _ := artifactoryService.createArtifactoryGroup(group, groupName)

		//Assert
		assert.Equal(t, expectedResp.StatusCode, resp.StatusCode)
	})
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

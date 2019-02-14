// Retrieves list of all repositories for an artifactory instance
package services

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"log"
	"net/http"
)

type ArtifactoryInformation struct {
	Owner       string
	ProjectName string
	Stage       string
	Location    string
	Description string
}

func permissionName(permission string, fields *ArtifactoryInformation) string {
	return fmt.Sprintf("%s-%s-docker-allrepos-%s-%s",
		fields.Owner,
		fields.ProjectName,
		fields.Location,
		permission)
}

func artifactoryRepositoryKey(fields *ArtifactoryInformation) string {
	return fmt.Sprintf("%s-%s-docker-%s-%s",
		fields.Owner,
		fields.ProjectName,
		fields.Stage,
		fields.Location)
}

func userName(permission string, fields *ArtifactoryInformation) string {
	return fmt.Sprintf("%s_%s_%s_%s", fields.Owner, fields.ProjectName, fields.Location, permission)
}

func (fields *ArtifactoryInformation) ArtifactoryRepositoryCreate(client *artifactory.Client) {
	artifactoryRepositoryName := artifactoryRepositoryKey(fields)

	repo := artifactory.LocalRepository{
		Key:             artifactory.String(artifactoryRepositoryName),
		RClass:          artifactory.String("local"),
		PackageType:     artifactory.String("docker"),
		HandleSnapshots: artifactory.Bool(false),
		Description:     artifactory.String(fields.Description),
	}

	existingRepo, response, err := client.Repositories.GetLocal(context.Background(), artifactoryRepositoryName)

	if response.StatusCode == http.StatusBadRequest && err != nil {
		resp, error := client.Repositories.CreateLocal(context.Background(), &repo)
		if error != nil {
			log.Printf(err.Error())
		} else if resp.StatusCode == http.StatusCreated {
			log.Printf("Creation of Repository %s", artifactoryRepositoryName)
		}
	} else if response.StatusCode == http.StatusUnauthorized {
		log.Println("Unauthorized access to Artifactory")
	} else if response.StatusCode == http.StatusOK && existingRepo != nil {
		resp, error := client.Repositories.UpdateLocal(context.Background(), artifactoryRepositoryName, &repo)
		if error != nil {
			log.Printf(err.Error())
		} else if resp.StatusCode == http.StatusOK {
			log.Printf("Update of Repository %s", artifactoryRepositoryName)
		}
	}
}

func (fields *ArtifactoryInformation) CreateArtifactoryGroup(client *artifactory.Client) {

	groupName := fmt.Sprintf("dl_artifactory_%s_%s", fields.Owner, fields.ProjectName)
	group := artifactory.Group{
		Name:        artifactory.String(groupName),
		Description: artifactory.String(fields.Description),
	}

	resp, err := client.Security.CreateOrReplaceGroup(context.Background(), groupName, &group)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("%d: Group %s created or replaced", resp.StatusCode, groupName)
	}
}

func createArtifactoryUsers(client *artifactory.Client, userName string, email string, password string, groups *[]string) {

	user := artifactory.User{
		Name:            artifactory.String(userName),
		Email:           artifactory.String(email),
		Password:        artifactory.String(password),
		DisableUIAccess: artifactory.Bool(true),
		Groups:          groups,
	}

	resp, err := client.Security.CreateOrReplaceUser(context.Background(), userName, &user)
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("%d: Users %s created or replaced", resp.StatusCode, userName)
	}
}

func (fields *ArtifactoryInformation) CreateArtifactoryUsers(client *artifactory.Client) {

	userNameRO := fmt.Sprintf("%s_%s_%s_k8s_reader", fields.Owner, fields.ProjectName, fields.Location)
	userEmailRO := fmt.Sprintf("%s@notanadress.ca.example.com", userNameRO)
	groupsRO := &[]string{"readers"}
	passwordRO := "toto"
	createArtifactoryUsers(client, userNameRO, userEmailRO, passwordRO, groupsRO)

	userNameRW := fmt.Sprintf("%s_%s_%s_jenkins_writer", fields.Owner, fields.ProjectName, fields.Location)
	userEmailRW := fmt.Sprintf("%s@notanadress.ca.example.com", userNameRW)
	groupsRW := &[]string{"readers"}
	passwordRW := "toto"
	createArtifactoryUsers(client, userNameRW, userEmailRW, passwordRW, groupsRW)
}

func createArtifactoryPermissions(
	client *artifactory.Client,
	permissionName string,
	artifactoryRepositoryName string,
	user *map[string][]string,
	group *map[string][]string,
) {

	permissions := artifactory.PermissionTargets{
		Name:         artifactory.String(permissionName),
		Repositories: &[]string{artifactoryRepositoryName},
		Principals: &artifactory.Principals{
			Users:  user,
			Groups: group,
		},
	}

	resp, err := client.Security.CreateOrReplacePermissionTargets(context.Background(), permissionName, &permissions)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("%d: Permission %s created or replaced", resp.StatusCode, permissionName)
	}

}

func (fields *ArtifactoryInformation) CreateArtifactoryPermissions(client *artifactory.Client) {

	permissionNameRO := permissionName("ro", fields)
	artifactoryRepositoryName := artifactoryRepositoryKey(fields)
	userNameRO := userName("k8s_reader", fields)
	userPermissions := []string{"r"}
	userRO := &map[string][]string{userNameRO: userPermissions}
	createArtifactoryPermissions(
		client,
		permissionNameRO,
		artifactoryRepositoryName,
		userRO,
		nil,
	)

	permissionNameRW := permissionName("rw", fields)
	userNameRW := userName("jenkins_writer", fields)
	userPermissionsRW := []string{"d", "w", "n", "r"}
	userRW := &map[string][]string{userNameRW: userPermissionsRW}
	groupNameRW := fmt.Sprintf("dl_artifactory_%s_%s", fields.Owner, fields.ProjectName)
	groupPermissionsRW := []string{"d", "w", "n", "r"}
	groupRW := &map[string][]string{groupNameRW: groupPermissionsRW}

	createArtifactoryPermissions(
		client,
		permissionNameRW,
		artifactoryRepositoryName,
		userRW,
		groupRW,
	)
}

// Retrieves list of all repositories for an artifactory instance
package main

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"log"
	"net/http"
	"os"
)

type ArtifactoryInformation struct {
	Owner       string
	ProjectName string
	Stage       string
	Location    string
	Description string
}

func (fields *ArtifactoryInformation) ArtifactoryRepositoryCreate(client *artifactory.Client) {
	artifactoryRepositoryKey := fmt.Sprintf("%s-%s-docker-%s-%s",
		fields.Owner,
		fields.ProjectName,
		fields.Stage,
		fields.Location)

	repo := artifactory.LocalRepository{
		Key:             artifactory.String(artifactoryRepositoryKey),
		RClass:          artifactory.String("local"),
		PackageType:     artifactory.String("docker"),
		HandleSnapshots: artifactory.Bool(false),
		Description:     artifactory.String(fields.Description),
	}

	existingRepo,resp, error :=client.Repositories.GetLocal(context.Background(), artifactoryRepositoryKey)

	if resp.StatusCode == http.StatusBadRequest && error != nil {
		client.Repositories.CreateLocal(context.Background(), &repo)
		log.Printf("Creation of Repository %s", artifactoryRepositoryKey)
	} else if resp.StatusCode == http.StatusUnauthorized {
		log.Println("Unauthorized access to Artifactory")
	} else if resp.StatusCode == http.StatusOK && existingRepo != nil {
		client.Repositories.UpdateLocal(context.Background(), artifactoryRepositoryKey, &repo)
		log.Printf("Update of Repository %s", artifactoryRepositoryKey)
	}
}

func (fields *ArtifactoryInformation) CreateArtifactoryGroup(client *artifactory.Client) {

	nameOfGroup := fmt.Sprintf("dl_artifactory_%s_%s", fields.Owner, fields.ProjectName)
	group := artifactory.Group{
		Name:            artifactory.String(nameOfGroup),
		Description:     artifactory.String(fields.Description),
	}

	resp, err := client.Security.CreateOrReplaceGroup(context.Background(), nameOfGroup, &group )
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("%d: Group %s created or replaced",resp.StatusCode, nameOfGroup)
	}
}

func (fields *ArtifactoryInformation) CreateArtifactoryUsers(client *artifactory.Client) {

	userNameRO := fmt.Sprintf("%s_%s_%s_k8s_reader", fields.Owner, fields.ProjectName, fields.Location)
	userNameRW := fmt.Sprintf("%s_%s_%s_jenkins_writer", fields.Owner, fields.ProjectName, fields.Location)
	userEmailRO := fmt.Sprintf("%s@notanadress.ca.example.com", userNameRO)
	userEmailRW := fmt.Sprintf("%s@notanadress.ca.example.com", userNameRW)

	userRO := artifactory.User{
		Name:                     artifactory.String(userNameRO),
		Email:                    artifactory.String(userEmailRO),
		Password:                 artifactory.String("toto"),
		Admin:                    nil,
		DisableUIAccess:          artifactory.Bool(true),
		InternalPasswordDisabled: nil,
		Realm:                    nil,
		Groups:                   &[]string{"readers"},
	}


	resp, err := client.Security.CreateOrReplaceUser(context.Background(), userNameRO, &userRO )
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("%d: Group %s created or replaced",resp.StatusCode, userNameRO)
	}

	userRW := artifactory.User{
		Name:                     artifactory.String(userNameRW),
		Email:                    artifactory.String(userEmailRW),
		Password:                 artifactory.String("toto"),
		Admin:                    nil,
		DisableUIAccess:          artifactory.Bool(true),
		InternalPasswordDisabled: nil,
		Realm:                    nil,
		Groups:                   &[]string{"readers"},
	}


	resp, err = client.Security.CreateOrReplaceUser(context.Background(), userNameRW, &userRW )
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("%d: Group %s created or replaced",resp.StatusCode, userNameRW)
	}

}

func main() {

	// TODO objet config, sortie en exit 1 si configuration invalidation
	tp := artifactory.BasicAuthTransport{
		Username: os.Getenv("ARTIFACTORY_USERNAME"),
		Password: os.Getenv("ARTIFACTORY_PASSWORD"),
	}

	client, err := artifactory.NewClient(os.Getenv("ARTIFACTORY_URL"), tp.Client())
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	fieldsArtifactory := &ArtifactoryInformation{
		Owner:       "aug",
		ProjectName: "e4",
		Stage:       "scratch",
		Location:    "intranet",
		Description: "Test fait par Aurélien Gabet, à supprimer ",
	}

	fieldsArtifactory.ArtifactoryRepositoryCreate(client)
	fieldsArtifactory.CreateArtifactoryGroup(client)
	fieldsArtifactory.CreateArtifactoryUsers(client)

}

package main

import (
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/ca-gip/artifactory-operator/internal/services"
	"os"
)

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

	fieldsArtifactory := &services.ArtifactoryInformation{
		Owner:       "aug",
		ProjectName: "e4",
		Stage:       "scratch",
		Location:    "intranet",
		Description: "Test fait par Aurélien Gabet, à supprimer ",
	}
	/*
		fieldsArtifactory.ArtifactoryRepositoryCreate(client)
		fieldsArtifactory.CreateArtifactoryGroup(client)
		fieldsArtifactory.CreateArtifactoryUsers(client)*/
	fieldsArtifactory.CreateArtifactoryPermissions(client)

}

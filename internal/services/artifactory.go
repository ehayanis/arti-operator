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

func (fields *ArtifactoryInformation) ArtifactoryRepositoryCreate(client *artifactory.Client, listOfRepositories []string) {
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

	if isRepositoryExist(listOfRepositories, artifactoryRepositoryKey) == true {
		client.Repositories.UpdateLocal(context.Background(), artifactoryRepositoryKey, &repo)
		fmt.Println("Update")
	} else {
		client.Repositories.CreateLocal(context.Background(), &repo)
		fmt.Println("Create")
	}
}

func isRepositoryExist(list []string, item string) bool {
	for i := 0; i < len(list); i++ {
		if list[i] == item {
			return true
		}
	}

	return false
}

func ListOfExistingRepositories(client *artifactory.Client) []string {
	var listOfRepo []string
	opts := artifactory.RepositoryListOptions{
		Type: "local",
	}

	repo,resp,error :=client.Repositories.GetLocal(context.Background(), "ok")

	// traitement
	if resp.StatusCode == http.StatusNotFound {

	} else if resp.StatusCode == http.StatusUnauthorized {

	}
	// traiter 200

	// update utiliser repo

	//

	//



	repos, _, err := client.Repositories.ListRepositories(context.Background(), &opts)
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return nil
	} else if repos == nil {
		fmt.Printf("\nerror: repos cannot be nil\n")
		return nil
	}

	for _, repo := range *repos {
		listOfRepo = append(listOfRepo,*repo.Key)
	}

	return listOfRepo
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
		log.Println(resp)
	}
}

func (fields *ArtifactoryInformation) CreateArtifactoryUsers(client *artifactory.Client) {

	userNameRO := fmt.Sprintf("%s_%s_k8s_reader", fields.Owner, fields.ProjectName, fields.Location)
	userNameRW := fmt.Sprintf("%s_%s_jenkins_writer", fields.Owner, fields.ProjectName, fields.Location)
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
		log.Println(resp)
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
		log.Println(resp)
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

	list := ListOfExistingRepositories(client)
	fieldsArtifactory.ArtifactoryRepositoryCreate(client, list)
	fieldsArtifactory.CreateArtifactoryGroup(client)
	fieldsArtifactory.CreateArtifactoryUsers(client)

}

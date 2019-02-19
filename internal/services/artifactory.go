package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/ca-gip/artifactory-operator/internal/config"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/rs/zerolog"
)

type ArtifactoryInformation struct {
	Tenant      string
	ProjectName string   // cr
	Stages      []string // cr
	Location    string   //param confmap
	Description string
}

type ArtifactoryService struct {
	logger               zerolog.Logger
	PasswordStoreService *PasswordStoreService
	artifactoryClient    *artifactory.Client
	artifactoryUrl       string
}

func NewArtifactoryService(operatorConfig *config.ArtifactoryOperatorConfig, PasswordStoreService *PasswordStoreService) (*ArtifactoryService, error) {
	logger := utils.Log.With().Str("service", "artifactory").Logger()

	tp := artifactory.BasicAuthTransport{
		Username: operatorConfig.ArtifactoryServerUser,
		Password: operatorConfig.ArtifactoryServerPassword,
	}

	client, err := artifactory.NewClient(operatorConfig.ArtifactoryServerUrl, tp.Client())
	if err != nil {
		logger.Info().Msgf("Couldn't create Artifactory client: %v", err)
		return nil, err
	}

	result := &ArtifactoryService{
		artifactoryClient:    client,
		artifactoryUrl:       operatorConfig.ArtifactoryServerUrl,
		PasswordStoreService: PasswordStoreService,
		logger:               logger,
	}

	return result, nil
}

func permissionName(permission string, fields *ArtifactoryInformation) string {
	return fmt.Sprintf("%s-%s-docker-allrepos-%s-%s",
		fields.Tenant,
		fields.ProjectName,
		fields.Location,
		permission)
}

func artifactoryRepositoryKey(fields *ArtifactoryInformation, stage string) string {
	tenantPrefix := fields.Tenant + "-"
	projectNameWithoutTenant := strings.TrimPrefix(fields.ProjectName, tenantPrefix)

	return fmt.Sprintf("%s-%s-docker-%s-%s",
		fields.Tenant,
		projectNameWithoutTenant,
		stage,
		fields.Location)
}

func userName(permission string, fields *ArtifactoryInformation) string {
	return fmt.Sprintf("%s_%s_%s_%s", fields.Tenant, fields.ProjectName, fields.Location, permission)
}

func (s *ArtifactoryService) ArtifactoryRepositoryCreate(fields *ArtifactoryInformation) ([]string, error) {
	result := []string{}

	for _, stage := range fields.Stages {

		artifactoryRepositoryName := artifactoryRepositoryKey(fields, stage)
		result = append(result, artifactoryRepositoryName)

		repo := artifactory.LocalRepository{
			Key:             artifactory.String(artifactoryRepositoryName),
			RClass:          artifactory.String("local"),
			PackageType:     artifactory.String("docker"),
			HandleSnapshots: artifactory.Bool(false),
			Description:     artifactory.String(fields.Description),
		}

		existingRepo, response, err := s.artifactoryClient.Repositories.GetLocal(context.Background(), artifactoryRepositoryName)

		if err != nil {
			if response == nil {
				s.logger.Error().Msgf("Unidentified error creating repo: %v", err)
				return result, err
			}
			if response.StatusCode == http.StatusBadRequest {
				resp, err := s.artifactoryClient.Repositories.CreateLocal(context.Background(), &repo)
				if err != nil {
					s.logger.Error().Msgf("Could not create artifactory repo %v: %v", artifactoryRepositoryName, err)
					return result, err
				} else if resp.StatusCode == http.StatusOK {
					s.logger.Info().Msgf("Creation of Repository %s successful.", artifactoryRepositoryName)
					continue
				}
			} else if response.StatusCode == http.StatusUnauthorized {
				s.logger.Error().Msgf("Could not access artifactory to create repo %v, unauthorized: %v", artifactoryRepositoryName, err)
				return result, err
			} else if response.StatusCode == http.StatusOK && existingRepo != nil {
				resp, err := s.artifactoryClient.Repositories.UpdateLocal(context.Background(), artifactoryRepositoryName, &repo)
				if err != nil {
					s.logger.Error().Msgf("Could not update artifactory repo %v: %v", artifactoryRepositoryName, err)
					return result, err
				} else if resp.StatusCode == http.StatusOK {
					s.logger.Info().Msgf("Update of Repository %s successful.", artifactoryRepositoryName)
					continue
				}
			}
		}
	}

	return result, nil
}

func (s *ArtifactoryService) CreateArtifactoryGroup(fields *ArtifactoryInformation) {

	groupName := fmt.Sprintf("dl_artifactory_%s_%s", fields.Tenant, fields.ProjectName)
	group := artifactory.Group{
		Name:        artifactory.String(groupName),
		Description: artifactory.String(fields.Description),
	}

	resp, err := s.artifactoryClient.Security.CreateOrReplaceGroup(context.Background(), groupName, &group)
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

type ArtifactoryRepoUsers struct {
	UserNameRO string
	PasswordRO string
	UserNameRW string
	PasswordRW string
}

func (s *ArtifactoryService) CreateArtifactoryUsers(fields *ArtifactoryInformation) (*ArtifactoryRepoUsers, error) {

	userNameRO := fmt.Sprintf("%s_%s_%s_k8s_reader", fields.Tenant, fields.ProjectName, fields.Location)
	userEmailRO := fmt.Sprintf("%s@notanadress.ca.example.com", userNameRO)
	groupsRO := &[]string{"readers"}
	passwordRO, err := s.PasswordStoreService.GetUserPassword(userNameRO)
	if err != nil {
		s.logger.Error().Msgf("Couldn't generate password for user %v: %v", userNameRO, err)
		return nil, err
	}
	s.logger.Debug().Msgf("Creating RO user %v.", userNameRO)
	createArtifactoryUsers(s.artifactoryClient, userNameRO, userEmailRO, passwordRO, groupsRO)

	userNameRW := fmt.Sprintf("%s_%s_%s_jenkins_writer", fields.Tenant, fields.ProjectName, fields.Location)
	userEmailRW := fmt.Sprintf("%s@notanadress.ca.example.com", userNameRW)
	groupsRW := &[]string{"readers"}
	passwordRW, err := s.PasswordStoreService.GetUserPassword(userNameRW)
	if err != nil {
		s.logger.Error().Msgf("Couldn't generate password for user %v: %v", userNameRW, err)
		return nil, err
	}
	s.logger.Debug().Msgf("Creating RW user %v.", userNameRW)
	createArtifactoryUsers(s.artifactoryClient, userNameRW, userEmailRW, passwordRW, groupsRW)

	result := &ArtifactoryRepoUsers{
		UserNameRO: userNameRO,
		PasswordRO: passwordRO,
		UserNameRW: userNameRW,
		PasswordRW: passwordRW,
	}

	return result, nil
}

func createArtifactoryPermissions(
	client *artifactory.Client,
	permissionName string,
	artifactoryRepositoryNames []string,
	user *map[string][]string,
	group *map[string][]string,
) {

	permissions := artifactory.PermissionTargets{
		Name:         artifactory.String(permissionName),
		Repositories: &artifactoryRepositoryNames,
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

func (s *ArtifactoryService) CreateArtifactoryPermissions(fields *ArtifactoryInformation) {
	repositories := []string{}
	for _, stage := range fields.Stages {
		repositories = append(repositories, artifactoryRepositoryKey(fields, stage))
	}

	permissionNameRO := permissionName("ro", fields)
	userNameRO := userName("k8s_reader", fields)
	userPermissions := []string{"r"}
	userRO := &map[string][]string{userNameRO: userPermissions}
	createArtifactoryPermissions(
		s.artifactoryClient,
		permissionNameRO,
		repositories,
		userRO,
		nil,
	)

	permissionNameRW := permissionName("rw", fields)
	userNameRW := userName("jenkins_writer", fields)
	userPermissionsRW := []string{"d", "w", "n", "r"}
	userRW := &map[string][]string{userNameRW: userPermissionsRW}
	groupNameRW := fmt.Sprintf("dl_artifactory_%s_%s", fields.Tenant, fields.ProjectName)
	groupPermissionsRW := []string{"d", "w", "n", "r"}
	groupRW := &map[string][]string{groupNameRW: groupPermissionsRW}

	createArtifactoryPermissions(
		s.artifactoryClient,
		permissionNameRW,
		repositories,
		userRW,
		groupRW,
	)
}

package services

import (
	"context"
	"fmt"
	"github.com/ca-gip/artifactory-operator/internal/types"
	"net/http"

	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/rs/zerolog"
)

type ArtifactoryService struct {
	logger               zerolog.Logger
	PasswordStoreService *PasswordStoreService
	artifactoryClient    *artifactory.Client
	artifactoryUrl       string
	clusterDNSSubdomain  string
}

func NewArtifactoryService(operatorConfig *types.ArtifactoryOperatorConfig, PasswordStoreService *PasswordStoreService) (*ArtifactoryService, error) {
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
		clusterDNSSubdomain:  operatorConfig.ClusterDNSSubdomain,
	}

	return result, nil
}

func permissionName(permission string, fields *types.ArtifactoryInformation, stage string) string {
	return fmt.Sprintf("%s-%s-docker-%s-%s-%s",
		fields.Tenant,
		fields.ProjectName,
		stage,
		fields.Location,
		permission)
}

func artifactoryRepositoryKey(fields *types.ArtifactoryInformation, stage string) string {
	return fmt.Sprintf("%s-%s-docker-%s-%s",
		fields.Tenant,
		fields.ProjectName,
		stage,
		fields.Location)
}

func stageInFields(stage string, fields *types.ArtifactoryInformation) bool {
	for _, fieldsStage := range fields.Stages {
		if stage == fieldsStage {
			return true
		}
	}

	return false
}

func (s *ArtifactoryService) ArtifactoryRepositoryCreate(fields *types.ArtifactoryInformation) ([]string, error) {
	result := []string{}

	for _, stage := range fields.Stages {
		artifactoryRepositoryName := artifactoryRepositoryKey(fields, stage)
		s.logger.Info().Msgf("Creating/updating repo %v", artifactoryRepositoryName)

		if stageInFields(stage, fields) {
			result = append(result, artifactoryRepositoryName)
		}

		repo := artifactory.LocalRepository{
			Key:             artifactory.String(artifactoryRepositoryName),
			RClass:          artifactory.String("local"),
			PackageType:     artifactory.String("docker"),
			HandleSnapshots: artifactory.Bool(false),
			Description:     artifactory.String(fields.Description),
			XrayIndex:       artifactory.Bool(true),
		}

		existingRepo, response, err := s.artifactoryClient.Repositories.GetLocal(context.Background(), artifactoryRepositoryName)
		//AUG: Attention, l'API Artifactory renvoie un 400 bad request si la ressource n'existe pas.
		if err != nil {
			resp, err := s.artifactoryClient.Repositories.CreateLocal(context.Background(), &repo)
			if err != nil {
				s.logger.Error().Msgf("Could not create artifactory repo %v: %v", artifactoryRepositoryName, err)
				return result, err
			} else if resp.StatusCode == http.StatusOK {
				s.logger.Info().Msgf("Creation of Repository %s successful.", artifactoryRepositoryName)
				continue
			}
		} else if response.StatusCode != http.StatusNotFound && existingRepo != nil {
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

	return result, nil
}

func (s *ArtifactoryService) CreateArtifactoryGroup(fields *types.ArtifactoryInformation) {

	groupName := fields.SourceEntity
	group := artifactory.Group{
		Name:        artifactory.String(groupName),
		Description: artifactory.String("created by new version"), //artifactory.String(fields.Description),
		Realm:       artifactory.String("ldap"),
		RealmAttributes: artifactory.String(fmt.Sprintf("ldapGroupName=%s", groupName)),
	}

	resp, err := s.artifactoryClient.Security.CreateOrReplaceGroup(context.Background(), groupName, &group)
	if err != nil {
		s.logger.Error().Msgf("Error creating or replacing group: %v", err)
	} else {
		s.logger.Info().Msgf("%d: Group %s created or replaced", resp.StatusCode, groupName)
	}
}

func (s *ArtifactoryService) createArtifactoryUsers(client *artifactory.Client, userName string, email string, password string, groups *[]string) {

	user := artifactory.User{
		Name:            artifactory.String(userName),
		Email:           artifactory.String(email),
		Password:        artifactory.String(password),
		DisableUIAccess: artifactory.Bool(true),
		Groups:          groups,
	}

	resp, err := client.Security.CreateOrReplaceUser(context.Background(), userName, &user)
	if err != nil {
		s.logger.Error().Msgf("Error creating or replacing user: %v", err)
	} else {
		s.logger.Info().Msgf("%d: Users %s created or replaced", resp.StatusCode, userName)
	}
}

func (s *ArtifactoryService) getVaultSecret(pathVault string, userNameRW string) (string, error) {

	passwordRW, err := s.PasswordStoreService.GetUserPassword(userNameRW)
	secretData := map[string]interface{}{
		"data": map[string]interface{}{
			userNameRW: passwordRW,
		},
	}
	_, err = VaultWriteSecret(s.PasswordStoreService.clientVault, secretData, pathVault)
	if err != nil {
		s.logger.Error().Msgf("Couldn't write secret in to vault for user %v: %v", userNameRW, err)
		return "", nil
	}

	s.logger.Info().Msgf("Password created and stored in Vault server for user %v", userNameRW)
	return passwordRW, nil
}


func (s *ArtifactoryService) CreateArtifactoryUsers(fields *types.ArtifactoryInformation) (*types.ArtifactoryRepoUsers, error) {

	//Generate Read Only users (used for k8s only)
	userNameRO, userEmailRO, groupsRO := s.generateUserFields(fields, "RO")
	passwordRO, err := s.PasswordStoreService.GetUserPassword(userNameRO)
	if err != nil {
		s.logger.Error().Msgf("Couldn't generate password for user %v: %v", userNameRO, err)
		return nil, err
	}
	s.logger.Info().Msgf("Creating RO user %v.", userNameRO)
	s.createArtifactoryUsers(s.artifactoryClient, userNameRO, userEmailRO, passwordRO, groupsRO)

	//Generate Read Write users (used as a service account for CI)
	userNameRW, userEmailRW, groupsRW := s.generateUserFields(fields, "RW")
	pathVault := fmt.Sprintf("%s/%s/%s/k8s/%s-%s/artifactory", utils.VaultStore, fields.Tenant, fields.ProjectName, s.clusterDNSSubdomain, fields.Environment)
	passwordRW, err := s.getVaultSecret(pathVault, userNameRW)

	if err != nil {
		s.logger.Error().Msgf("Couldn't generate password for user %v: %v", userNameRW, err)
		return nil, err
	}

	//Create or Update user each time the operator pass
	s.createArtifactoryUsers(s.artifactoryClient, userNameRW, userEmailRW, passwordRW, groupsRW)

	result := &types.ArtifactoryRepoUsers{
		UserNameRO: userNameRO,
		PasswordRO: passwordRO,
		UserNameRW: userNameRW,
		PasswordRW: passwordRW,
	}

	return result, nil
}


func (s *ArtifactoryService) generateUserFields(fields *types.ArtifactoryInformation, mode string) (string, string, *[]string) {

	suffix := ""
	if mode == "RW" {
		suffix = utils.ArtifactoryUserRWSuffix
	} else if mode == "RO" {
		if fields.Environment == "production" {
			suffix = utils.ArtifactoryUserROSuffixProduction
		} else {
			suffix = utils.ArtifactoryUserROSuffixNonProduction
		}
	}
	userName := fmt.Sprintf("%s_%s_%s_%s", fields.Tenant, fields.ProjectName, fields.Location, suffix)
	userEmail := fmt.Sprintf("%s@notanadress.ca.example.com", userName)
	groups := &[]string{"readers"}
	return userName, userEmail, groups
}

func (s *ArtifactoryService) createArtifactoryPermissions(
	client *artifactory.Client,
	permissionName string,
	artifactoryRepositoryNames []string,
	user *map[string][]string,
	group *map[string][]string,
) {

	repositories := &artifactoryRepositoryNames

	//check if permissions already exists
	existingPermission, resp, err := client.Security.GetPermissionTargets(context.Background(), permissionName)
	if err != nil {
		s.logger.Error().Msgf("Error listing existing permissions : %v", err)
	} else if existingPermission != nil {
		*repositories = append(*repositories, *existingPermission.Repositories...)
		*repositories = utils.Uniq(*repositories)
	}

	permissions := artifactory.PermissionTargets{
		Name:         artifactory.String(permissionName),
		Repositories: repositories,
		Principals: &artifactory.Principals{
			Users:  user,
			Groups: group,
		},
	}

	resp, err = client.Security.CreateOrReplacePermissionTargets(context.Background(), permissionName, &permissions)
	if err != nil {
		s.logger.Error().Msgf("Error creating or replacing Permission : %v", err)
	} else {
		s.logger.Info().Msgf("%d: Permission %s created or replaced", resp.StatusCode, permissionName)
	}

}

func (s *ArtifactoryService) CreateArtifactoryPermissions(fields *types.ArtifactoryInformation, users *types.ArtifactoryRepoUsers) {

	s.logger.Info().Msgf("START CreateArtifactoryPermissions")
	userPermissionsRO := []string{"r"}
	userPermissionsRW := []string{"d", "w", "n", "r"}
	groupPermissionsRW := []string{"d", "w", "n", "r"}
	userRO := &map[string][]string{users.UserNameRO: userPermissionsRO}
	userRW := &map[string][]string{users.UserNameRW: userPermissionsRW}
	
	groupNameRW := fields.SourceEntity
	groupRW := &map[string][]string{groupNameRW: groupPermissionsRW}

	for _, stage := range fields.Stages {
		permissionNameRO := permissionName("ro", fields, stage)
		repos := artifactoryRepositoryKey(fields, stage)
		s.logger.Info().Msgf("createArtifactoryPermissions(%s, %s, %s, %s)", permissionNameRO, repos, userRO, groupRW)
		s.createArtifactoryPermissions(
			s.artifactoryClient,
			permissionNameRO,
			[]string{repos},
			userRO,
			groupRW,
		)
		if stage == utils.ArtifactoryStageScratch || stage == utils.ArtifactoryStageStaging {
			permissionNameRW := permissionName("rw", fields, stage)
			s.logger.Info().Msgf("createArtifactoryPermissions(%s, %s, %s, %s)", permissionNameRW, repos, userRW, groupRW)
			s.createArtifactoryPermissions(
				s.artifactoryClient,
				permissionNameRW,
				[]string{repos},
				userRW,
				groupRW,
			)
		}
	}
	s.logger.Info().Msgf("FINISH CreateArtifactoryPermissions")
}

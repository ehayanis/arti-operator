package services

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/rs/zerolog"
	"net/http"
)

type ArtifactoryService struct {
	logger               zerolog.Logger
	PasswordStoreService *PasswordStoreService
	artifactoryClient    *artifactory.Client
	artifactoryUrl       string
	clusterDNSSubdomain  string
	LDAPGroups			 types.LDAPGroups
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
		LDAPGroups:           operatorConfig.LDAPGroups,
	}

	return result, nil
}

func permissionName(permission string, fields *types.ArtifactoryInformation) string {
	return fmt.Sprintf("%s-%s-docker-allrepos-%s-%s",
		fields.Tenant,
		fields.ProjectName,
		fields.Location,
		permission)

}

func permissionEnvName(permission string, fields *types.ArtifactoryInformation) string {
	return fmt.Sprintf("%s-%s-docker-%s-%s-%s",
		fields.Tenant,
		fields.ProjectName,
		fields.Environment,
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

		if stageInFields(stage, fields) {
			result = append(result, artifactoryRepositoryName)
		}

		repo := artifactory.LocalRepository{
			Key:             artifactory.String(artifactoryRepositoryName),
			RClass:          artifactory.String("local"),
			PackageType:     artifactory.String("docker"),
			HandleSnapshots: artifactory.Bool(false),
			Description:     artifactory.String(fields.Description),
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
			if *repo.Description == *existingRepo.Description && *repo.HandleSnapshots == *existingRepo.HandleSnapshots && *repo.PackageType == *existingRepo.PackageType && *repo.RClass == *existingRepo.RClass {
				s.logger.Debug().Msgf("Update not necessary, skipping the repository %v", artifactoryRepositoryName)
				continue
			}

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

	//Create Group function to Source Entity read on project
	groupName := fields.SourceEntity
	group := &artifactory.Group{
		Name:            artifactory.String(groupName),
		Description:     artifactory.String("created by Artifactory Operator"),
		Realm:           artifactory.String("ldap"),
		RealmAttributes: artifactory.String(fmt.Sprintf("ldapGroupName=%s", groupName)),
	}

	s.createArtifactoryGroup(group, groupName)

	//Create Group for Customer OPS
	customerOPSGroup := &artifactory.Group{
		Name:            artifactory.String(s.LDAPGroups.CustomerOPS),
		Description:     artifactory.String("created by Artifactory Operator"),
		Realm:           artifactory.String("ldap"),
		RealmAttributes: artifactory.String(fmt.Sprintf("ldapGroupName=%s", s.LDAPGroups.CustomerOPS)),
	}

	s.createArtifactoryGroup(customerOPSGroup, s.LDAPGroups.CustomerOPS)
}

func (s *ArtifactoryService) createArtifactoryGroup(group *artifactory.Group, groupName string) {
	group, resp, err := s.artifactoryClient.Security.GetGroup(context.Background(), groupName)

	if resp.StatusCode == http.StatusNotFound {
		s.logger.Info().Msgf("Group %v doesn't exist. Will be created", groupName)
		resp, err = s.artifactoryClient.Security.CreateOrReplaceGroup(context.Background(), groupName, group)
	}

	if err != nil {
		s.logger.Error().Msgf("Technical error occured during replacing group: %v", err)
	}
}

func (s *ArtifactoryService) createArtifactoryUsers(client *artifactory.Client, userName string, email string, password string) {

	user := artifactory.User{
		Name:            artifactory.String(userName),
		Email:           artifactory.String(email),
		Password:        artifactory.String(password),
		DisableUIAccess: artifactory.Bool(true),
		Groups:          nil,
	}

	existingUser, resp, err := client.Security.GetUser(context.Background(), userName)

	if err != nil {
		s.logger.Error().Msgf("Technical error occured during user creation/update: %s", err.Error())
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		resp, err = client.Security.CreateOrReplaceUser(context.Background(), userName, &user)
		if err == nil {
			s.logger.Info().Msgf("Users %s created", userName)
		}
	} else if *existingUser.DisableUIAccess != *user.DisableUIAccess || *existingUser.Email != *user.Email {
		existingUser.Email = user.Email
		existingUser.DisableUIAccess = user.DisableUIAccess
		existingUser.Password = user.Password
		resp, err = client.Security.CreateOrReplaceUser(context.Background(), userName, existingUser)
		if err == nil {
			s.logger.Info().Msgf("Users %s updated", userName)
		}
	} else {
		s.logger.Debug().Msgf("Skipping User %s, up to date.", userName)
	}

	if err != nil {
		s.logger.Error().Msgf("Technical error occured during user creation/update: %s", err.Error())
	}

}

func (s *ArtifactoryService) getVaultSecret(pathVault string, userNameRW string) (string, error) {

	passwordRW, err := s.PasswordStoreService.GetUserPassword(userNameRW)
	secretData := map[string]interface{}{
		"data": map[string]interface{}{
			userNameRW: passwordRW,
		},
	}
	_, err, changed := VaultWriteSecret(s.PasswordStoreService.clientVault, secretData, pathVault)
	if err != nil {
		s.logger.Error().Msgf("Couldn't write secret in to vault for user %v: %v", userNameRW, err)
		return "", nil
	} else if changed {
		s.logger.Info().Msgf("Password created and stored in Vault server for user %v", userNameRW)
	} else if !changed {
		s.logger.Debug().Msgf("Update for password is not necessary %v", userNameRW)
	}
	return passwordRW, nil
}

func (s *ArtifactoryService) CreateArtifactoryUsers(fields *types.ArtifactoryInformation) (*types.ArtifactoryRepoUsers, error) {

	//Generate Read Only users (used for k8s only)
	userNameRO, userEmailRO := s.generateUserFields(fields, "RO")
	passwordRO, err := s.PasswordStoreService.GetUserPassword(userNameRO)
	if err != nil {
		s.logger.Error().Msgf("Couldn't generate password for user %v: %v", userNameRO, err)
		return nil, err
	}

	s.createArtifactoryUsers(s.artifactoryClient, userNameRO, userEmailRO, passwordRO)

	//Generate Read Write users (used as a service account for CI)
	userNameRW, userEmailRW := s.generateUserFields(fields, "RW")
	pathVault := fmt.Sprintf("%s/%s/%s/k8s/%s-%s/artifactory", utils.VaultStore, fields.Tenant, fields.ProjectName, s.clusterDNSSubdomain, fields.Environment)
	passwordRW, err := s.getVaultSecret(pathVault, userNameRW)

	if err != nil {
		s.logger.Error().Msgf("Couldn't generate password for user %v: %v", userNameRW, err)
		return nil, err
	}

	//Create or Update user each time the operator pass
	s.createArtifactoryUsers(s.artifactoryClient, userNameRW, userEmailRW, passwordRW)

	result := &types.ArtifactoryRepoUsers{
		UserNameRO: userNameRO,
		PasswordRO: passwordRO,
		UserNameRW: userNameRW,
		PasswordRW: passwordRW,
	}

	return result, nil
}

func (s *ArtifactoryService) generateUserFields(fields *types.ArtifactoryInformation, mode string) (string, string) {

	suffix := ""
	if mode == "RW" {
		suffix = utils.ArtifactoryUserRWSuffix
	} else if mode == "RO" {
		suffix = utils.ArtifactoryUserROSuffixNonProduction
	}

	userName := fmt.Sprintf("%s_%s_%s_%s", fields.Tenant, fields.ProjectName, fields.Location, suffix)
	userEmail := fmt.Sprintf("%s@notanadress.ca.example.com", userName)
	return userName, userEmail
}

// TODO Elie, vérifier si possibilité de désactiver l'update quand les groups sont equivalents ainsi que les repositories
func (s *ArtifactoryService) createArtifactoryPermissions(permissionName string, artifactoryRepositoryNames []string, user *map[string][]string, group *map[string][]string) {

	// Initializing default struct
	permissions := artifactory.PermissionTargets{
		Name:         artifactory.String(permissionName),
		Repositories: &artifactoryRepositoryNames,
		Principals: &artifactory.Principals{
			Users:  user,
			Groups: group,
		},
	}

	existingPermissions, resp, err := s.artifactoryClient.Security.GetPermissionTargets(context.Background(), permissionName)

	if resp == nil && err != nil {
		s.logger.Error().Msgf("Technical Error occured during GetPermissionTargets reponse is empty and got error : '%s'", err.Error())
		return
	}

	if err != nil {
		s.logger.Error().Msgf("Technical Error occured during permission creation/update: '%s'", err.Error())
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		resp, err = s.artifactoryClient.Security.CreateOrReplacePermissionTargets(context.Background(), permissionName, &permissions)
		if err == nil {
			s.logger.Info().Msgf("Permission %s created", resp.StatusCode, permissionName)
		}
	} else {
		repositories := &artifactoryRepositoryNames
		*repositories = append(*repositories, *existingPermissions.Repositories...)
		*repositories = utils.Uniq(*repositories)

		expectedGroups := utils.MapConcat(existingPermissions.Principals.Groups, group)
		if (!utils.Equal(existingPermissions.Repositories, repositories)) || !utils.MapEquals(existingPermissions.Principals.Groups, expectedGroups) {
			permissions.Principals.Groups = expectedGroups
			resp, err = s.artifactoryClient.Security.CreateOrReplacePermissionTargets(context.Background(), permissionName, &permissions)
			if err == nil {
				s.logger.Info().Msgf("Permission %s replaced for repository %v", resp.StatusCode, permissionName, *repositories)
			}
		}
	}

	if err != nil {
		s.logger.Error().Msgf("Technical Error occured during permission creation/update: '%s'", err.Error())
	}

}

func (s *ArtifactoryService) CreateArtifactoryPermissions(fields *types.ArtifactoryInformation, users *types.ArtifactoryRepoUsers) {
	readOnlyRepositories := []string{}
	readWriteRepositories := []string{}

	for _, stage := range fields.Stages {
		if stage == utils.ArtifactoryStageScratch || stage == utils.ArtifactoryStageStaging {
			readWriteRepositories = append(readWriteRepositories, artifactoryRepositoryKey(fields, stage))
		}
		readOnlyRepositories = append(readOnlyRepositories, artifactoryRepositoryKey(fields, stage))
	}

	permissionsRO := []string{"r"}
	s.createServiceAccountPermissions(fields, users, permissionsRO, readOnlyRepositories, "ro")
	s.createLDAPPermissions(fields, permissionsRO, readOnlyRepositories, "ro")

	// No need to create RW permission for Production environment
	if len(readWriteRepositories) != 0 {
		permissionsRW := []string{"d", "w", "n", "r"}
		s.createServiceAccountPermissions(fields, users, permissionsRW, readWriteRepositories, "rw")
		s.createLDAPPermissions(fields, permissionsRW, readWriteRepositories, "rw")
	}

}

func (s *ArtifactoryService) createLDAPPermissions(fields *types.ArtifactoryInformation, permissions []string, repositories []string, role string) {
	permissionEnvName := permissionEnvName(role, fields)
	group := &map[string][]string{fields.SourceEntity: permissions, s.LDAPGroups.CustomerOPS: permissions}
	s.createArtifactoryPermissions(permissionEnvName, repositories, nil, group)
}

func (s *ArtifactoryService) createServiceAccountPermissions(fields *types.ArtifactoryInformation, users *types.ArtifactoryRepoUsers, permissions []string, repositories []string, role string) {
	var userRole *map[string][]string
	userRole = setUserRole(role, userRole, users, permissions)
	permissionName := permissionName(role, fields)
	s.createArtifactoryPermissions(permissionName, repositories, userRole, nil)
}

func setUserRole(role string, userRole *map[string][]string, users *types.ArtifactoryRepoUsers, permissions []string) *map[string][]string {
	if role == "ro" {
		userRole = &map[string][]string{users.UserNameRO: permissions}
	} else if role == "rw" {
		userRole = &map[string][]string{users.UserNameRW: permissions}
	}
	return userRole
}

package services

import (
	"errors"
	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	kubiv1 "github.com/ca-gip/kubi/pkg/apis/ca-gip/v1"
	"github.com/rs/zerolog"
	"net/url"
	"strings"
)

type ProjectService struct {
	logger                     zerolog.Logger
	dockerConfigSecretsService *DockerConfigSecretsService
	artifactoryService         *ArtifactoryService
	artifactoryHostBase        string
	clusterLocation            string
	xrayService                *XrayService
}

func NewProjectService(operatorConfig *types.ArtifactoryOperatorConfig,
	dockerConfigSecretsService *DockerConfigSecretsService,
	artifactoryService *ArtifactoryService,
	xrayService *XrayService) (*ProjectService, error) {
	logger := utils.Log.With().Str("service", "project").Logger()

	parsedArtifactoryUri, err := url.Parse(operatorConfig.ArtifactoryServerUrl)
	if err != nil {
		logger.Error().Msgf("Couldn't parse Artifactory URI: %v", err)
		return nil, err
	}

	result := &ProjectService{
		logger: logger,
		dockerConfigSecretsService: dockerConfigSecretsService,
		artifactoryService:         artifactoryService,
		clusterLocation:            operatorConfig.ClusterLocation,
		artifactoryHostBase:        parsedArtifactoryUri.Host,
		xrayService:                xrayService,
	}

	return result, nil
}

func (s *ProjectService) HandleProject(project *kubiv1.Project) error {
	repos, users, err := s.createArtifactoryResources(project)
	if err != nil {
		return err
	}

	if len(repos) > 0 {
		s.createXrayResources(repos)
	}

	err = s.createDockerSecret(project, repos, users)
	if err != nil {
		return err
	}

	return nil
}

func (s *ProjectService) createXrayResources(repos []string) {
	s.xrayService.createXrayPolicy(repos)
	s.xrayService.createXrayWatch(repos)
}

func (s *ProjectService) createDockerSecret(project *kubiv1.Project, repos []string, users *types.ArtifactoryRepoUsers) error {
	registries := []types.DockerConfigRegistryInfo{}

	for _, reg := range repos {
		regElt := types.DockerConfigRegistryInfo{
			Username: users.UserNameRO,
			Password: users.PasswordRO,
			Url:      registryUrl(s.artifactoryHostBase, reg),
		}

		registries = append(registries, regElt)
	}
	secretContents := &types.DockerConfigSecret{
		Name:       "project-registries",
		Registries: registries,
	}

	_, err := s.dockerConfigSecretsService.CreateOrUpdateDockerConfigSecret(project.Name, secretContents)
	if err != nil {
		s.logger.Error().Msgf("Couldn't create DockerConfig secret: %v", err)
		return err
	}

	return nil
}

func registryUrl(artifactoryHostBase, registryName string) string {
	return registryName + "." + artifactoryHostBase
}

func (s *ProjectService) createArtifactoryResources(project *kubiv1.Project) ([]string, *types.ArtifactoryRepoUsers, error) {
	fieldsArtifactory, err := s.generateArtifactoryFields(project)
	if err != nil {
		s.logger.Error().Msgf("Couldn't read Project fields for project %v: %v", project.Spec, err)
		return nil, nil, err
	}

	repos, err := s.artifactoryService.ArtifactoryRepositoryCreate(fieldsArtifactory)
	if err != nil {
		s.logger.Error().Msgf("Couldn't create Artifactory repos for project %v: %v", project.Spec, err)
		return nil, nil, err
	}

	s.artifactoryService.CreateArtifactoryGroup(fieldsArtifactory)

	users, err := s.artifactoryService.CreateArtifactoryUsers(fieldsArtifactory)
	if err != nil {
		s.logger.Error().Msgf("Couldn't create Artifactory users for project %v: %v", project.Spec, err)
		return nil, nil, err
	}

	s.artifactoryService.CreateArtifactoryPermissions(fieldsArtifactory, users)

	return repos, users, nil
}

func (s *ProjectService) generateArtifactoryFields(project *kubiv1.Project) (*types.ArtifactoryInformation, error) {
	if strings.TrimSpace(project.Spec.Project) == "" {
		return nil, errors.New("Spec.Project empty.")
	}

	if strings.TrimSpace(project.Spec.Tenant) == "" {
		return nil, errors.New("Spec.Tenant empty.")
	}

	if len(project.Spec.Stages) == 0 {
		return nil, errors.New("Spec.Stages empty.")
	}

	projectNameWithoutTenant := strings.TrimPrefix(project.Spec.Project, project.Spec.Tenant+"-")
	fieldsArtifactory := &types.ArtifactoryInformation{
		Tenant:      project.Spec.Tenant,
		ProjectName: projectNameWithoutTenant,
		Stages:      project.Spec.Stages,
		Location:    s.clusterLocation,
		Description: utils.ArtifactoryDescription,
		Environment: project.Spec.Environment,
	}

	return fieldsArtifactory, nil
}

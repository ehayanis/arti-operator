package services

import (
	"errors"
	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	kubiv1 "github.com/ca-gip/kubi/pkg/apis/ca-gip/v1"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	kClient                    *kubernetes.Clientset
}

func NewProjectService(operatorConfig *types.ArtifactoryOperatorConfig,
	dockerConfigSecretsService *DockerConfigSecretsService,
	artifactoryService *ArtifactoryService,
	xrayService *XrayService,
	kconfig *rest.Config) (*ProjectService, error) {
	logger := utils.Log.With().Str("service", "project").Logger()

	parsedArtifactoryUri, err := url.Parse(operatorConfig.ArtifactoryServerUrl)
	if err != nil {
		logger.Error().Msgf("Couldn't parse Artifactory URI: %v", err)
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		logger.Error().Msgf("Could not create k8s client: %v", err)
		return nil, err
	}

	result := &ProjectService{
		logger: logger,
		dockerConfigSecretsService: dockerConfigSecretsService,
		artifactoryService:         artifactoryService,
		clusterLocation:            operatorConfig.ClusterLocation,
		artifactoryHostBase:        parsedArtifactoryUri.Host,
		xrayService:                xrayService,
		kClient:                    kubeClient,
	}

	return result, nil
}

func (s *ProjectService) HandleProject(project *kubiv1.Project) error {
	repos, users, err := s.createArtifactoryResources(project)
	if err != nil {
		return err
	}

/*	if len(repos) > 0 {
		s.createXrayResources(repos)
	}*/

	err = s.createDockerSecret(project, repos, users)
	if err != nil {
		return err
	}

	err = s.annotateNamespace(project, repos)
	if err != nil {
		return err
	}

	return nil
}

func (s *ProjectService) annotateNamespace(project *kubiv1.Project, repos []string) error {
	currentNamespace, err := s.kClient.CoreV1().Namespaces().Get(project.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	var fqdnRepos []string

	for _, repo := range repos {
		fqdnRepos = append(fqdnRepos, repo+utils.Domain)
		fqdnRepos = append(fqdnRepos, repo+utils.DomainGroup)
	}

	if utils.ArtifactoryProjectSpecEnvironment == project.Spec.Environment {
		fqdnRepos = append(fqdnRepos, utils.DockerRemote)
		fqdnRepos = append(fqdnRepos, utils.DockerRemoteGroup)
	}

	annotations := make(map[string]string)
	annotations[utils.WhitelistKey] = strings.Join(fqdnRepos[:], ",")
	currentNamespace.SetAnnotations(annotations)

	_, err = s.kClient.CoreV1().Namespaces().Update(currentNamespace)
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
		SourceEntity: project.Spec.SourceEntity,
	}

	return fieldsArtifactory, nil
}

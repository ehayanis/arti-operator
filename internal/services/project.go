package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	kubiv1 "github.com/ca-gip/kubi/pkg/apis/cagip/v1"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ProjectSkip struct {
	password_name string
	creation_time time.Time
}

var skipped_projects []ProjectSkip

// var ResyncPeriod = time.Second * 10

type ProjectService struct {
	logger                     zerolog.Logger
	dockerConfigSecretsService *DockerConfigSecretsService
	artifactoryService         *ArtifactoryService
	artifactoryHostBase        string
	clusterLocation            string
	kClient                    *kubernetes.Clientset
	resyncPeriod               time.Duration
}

func NewProjectService(operatorConfig *types.ArtifactoryOperatorConfig,
	dockerConfigSecretsService *DockerConfigSecretsService,
	artifactoryService *ArtifactoryService,
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
		logger:                     logger,
		dockerConfigSecretsService: dockerConfigSecretsService,
		artifactoryService:         artifactoryService,
		clusterLocation:            operatorConfig.ClusterLocation,
		artifactoryHostBase:        parsedArtifactoryUri.Host,
		kClient:                    kubeClient,
		resyncPeriod:               operatorConfig.ProjectResyncPeriod,
	}

	return result, nil
}

func (s *ProjectService) HandleProject(project *kubiv1.Project) error {
	// Check if this is a v2 project
	isV2 := utils.IsV2Project(project)

	repos, users, err := s.createArtifactoryResources(project, isV2)
	if err != nil {
		return err
	}

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
	currentNamespace, err := s.kClient.CoreV1().Namespaces().Get(context.TODO(), project.Name, metav1.GetOptions{})
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

	_, err = s.kClient.CoreV1().Namespaces().Update(context.TODO(), currentNamespace, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
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

func (s *ProjectService) createArtifactoryResources(project *kubiv1.Project, isV2 bool) ([]string, *types.ArtifactoryRepoUsers, error) {

	curTime := time.Now()
	coreClient, err := kubernetes.NewForConfig(s.artifactoryService.PasswordStoreService.clientConfig)

	if err != nil {
		s.logger.Info().Err(err).Msg("Cannot get kubernetes core API client.")
		return nil, nil, err
	}

	fieldsArtifactory, err := s.generateArtifactoryFields(project, isV2)
	if err != nil {
		s.logger.Error().Msgf("Couldn't read Project fields for project %v: %v", project.Spec, err)
		return nil, nil, err
	}

	// s.logger.Error().Msgf("createArtifactoryResources(%v)", project.Name)

	userNameRO, _ := s.artifactoryService.generateUserFields(fieldsArtifactory, "RO")
	secretName := s.artifactoryService.PasswordStoreService.getSecretNameForUser(userNameRO)

	skip_found := false
	if len(skipped_projects) > 0 {
		for i, sk := range skipped_projects {
			if sk.password_name == secretName {
				newdate := sk.creation_time.Add(s.resyncPeriod)
				if newdate.Before(time.Now()) {
					if len(skipped_projects) == 1 {
						skipped_projects = nil
					} else {
						skipped_projects[i] = skipped_projects[len(skipped_projects)-1]
						skipped_projects = skipped_projects[:len(skipped_projects)-1]
					}
					break
				} else {
					skip_found = true
				}
			}
			if skip_found {
				break
			}
		}
	}

	if !skip_found {
		secret, err := coreClient.CoreV1().Secrets(s.artifactoryService.PasswordStoreService.secretsNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err == nil {
			_, keyExists := secret.Data["password"]
			if keyExists {
				repositoryNames := []string{}
				for _, stage := range fieldsArtifactory.Stages {
					artifactoryRepositoryName := artifactoryRepositoryKey(fieldsArtifactory, stage)
					if stageInFields(stage, fieldsArtifactory) {
						repositoryNames = append(repositoryNames, artifactoryRepositoryName)
					}
				}

				if s.artifactoryService.SharedRepository == "true" {
					sharedRepoName := fmt.Sprintf("%s-shared-docker-stable-%s", fieldsArtifactory.Tenant, fieldsArtifactory.Location)
					repositoryNames = append(repositoryNames, sharedRepoName)
				}

				passwordRO, err := s.artifactoryService.PasswordStoreService.GetUserPassword(userNameRO)
				userNameRW, _ := s.artifactoryService.generateUserFields(fieldsArtifactory, "RW")
				pathVault := fmt.Sprintf("%s/%s/%s/k8s/%s-%s/artifactory", utils.VaultStore, fieldsArtifactory.Tenant, fieldsArtifactory.ProjectName, s.artifactoryService.clusterDNSSubdomain, fieldsArtifactory.Environment)
				passwordRW, err := s.artifactoryService.getVaultSecret(pathVault, userNameRW)

				if err != nil {
					s.logger.Error().Msgf("Couldn't generate password for user %v: %v", userNameRW, err)
				} else {
					users := &types.ArtifactoryRepoUsers{
						UserNameRO: userNameRO,
						PasswordRO: passwordRO,
						UserNameRW: userNameRW,
						PasswordRW: passwordRW,
					}
					s.logger.Debug().Msgf("Secret %v already exist, skipping project : %v", secretName, project.Name)
					return repositoryNames, users, nil
				}
			}
		}
	}

	skip_found = false
	if len(skipped_projects) > 0 {
		for _, p := range skipped_projects {
			if p.password_name == secretName {
				skip_found = true
				break
			}
		}
	}

	if !skip_found {
		pskip := ProjectSkip{
			password_name: secretName,
			creation_time: curTime,
		}
		skipped_projects = append(skipped_projects, pskip)
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

func (s *ProjectService) generateArtifactoryFields(project *kubiv1.Project, isV2 bool) (*types.ArtifactoryInformation, error) {
	if strings.TrimSpace(project.Spec.Project) == "" {
		return nil, errors.New("Spec.Project empty.")
	}

	if strings.TrimSpace(project.Spec.Tenant) == "" {
		return nil, errors.New("Spec.Tenant empty.")
	}

	// For v2 resources, we don't require stages
	if !isV2 && len(project.Spec.Stages) == 0 {
		return nil, errors.New("Spec.Stages empty.")
	}

	// Set default stages for v2 resources if they don't have any
	stages := project.Spec.Stages
	if isV2 && len(stages) == 0 {
		s.logger.Info().Msgf("Setting default stages for v2 project %s", project.Name)
		stages = []string{"stable"}
	}

	projectNameWithoutTenant := strings.TrimPrefix(project.Spec.Project, project.Spec.Tenant+"-")
	fieldsArtifactory := &types.ArtifactoryInformation{
		Tenant:      project.Spec.Tenant,
		ProjectName: projectNameWithoutTenant,
		Stages:      stages,
		Location:    s.clusterLocation,
		Description: utils.ArtifactoryDescription,
		Environment: project.Spec.Environment,
		SourceDN:    project.Spec.SourceDN,
	}

	return fieldsArtifactory, nil
}

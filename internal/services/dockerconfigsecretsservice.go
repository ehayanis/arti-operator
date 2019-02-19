package services

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/ca-gip/artifactory-operator/internal/config"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type DockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Auth     string `json:"auth,omitempty"`
}

type DockerConfigSecretsService struct {
	logger       zerolog.Logger
	clientConfig *rest.Config
}

type DockerConfigRegistryInfo struct {
	Url      string
	Username string
	Password string
}

type DockerConfigSecret struct {
	Name       string
	Registries []DockerConfigRegistryInfo
}

func NewDockerConfigSecretsService(kconfig *rest.Config, operatorConfig *config.ArtifactoryOperatorConfig) *DockerConfigSecretsService {
	result := &DockerConfigSecretsService{
		logger:       utils.Log.With().Str("service", "imagepullsecrets").Logger(),
		clientConfig: kconfig,
	}

	return result
}

func (s *DockerConfigSecretsService) CreateOrUpdateDockerConfigSecret(namespace string, ips *DockerConfigSecret) (*v1.Secret, error) {
	secret, err := generateSecretObject(ips)

	if err != nil {
		return nil, err
	}

	coreClient, err := kubernetes.NewForConfig(s.clientConfig)

	if err != nil {
		s.logger.Info().Err(err).Msg("Cannot get kubernetes core API client.")
		return nil, err
	}

	secretsClient := coreClient.CoreV1().Secrets(namespace)

	createdSecret, err := secretsClient.Create(secret)

	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			// Secret already exists, replace it with our version
			s.logger.Debug().Msgf("Updating secret for namespace %v.", namespace)
			updatedSecret, err := secretsClient.Update(secret)

			if err != nil {
				return nil, err
			}

			return updatedSecret, nil
		} else {
			return nil, err
		}
	}

	return createdSecret, nil
}

func generateSecretObject(ips *DockerConfigSecret) (*v1.Secret, error) {
	registriesBlocks := map[string]DockerConfigEntry{}

	for _, elt := range ips.Registries {
		registriesBlocks[elt.Url] = getDockerConfigRegistryBlock(&elt)
	}

	dockerConfigStruct := map[string]map[string]DockerConfigEntry{
		"auths": registriesBlocks,
	}

	cleartextDockerConfig, err := json.Marshal(dockerConfigStruct)

	if err != nil {
		return nil, err
	}

	secret := &v1.Secret{
		Type: v1.SecretTypeDockerConfigJson,
		ObjectMeta: metav1.ObjectMeta{
			Name: ips.Name,
			Labels: map[string]string{
				"creator": "artifactory-operator",
			},
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte(cleartextDockerConfig),
		},
	}

	return secret, nil
}

func getAuthStringFromUsernamePassword(username, password string) string {
	cleartextAuthString := strings.Join([]string{username, password}, ":")
	base64AuthString := base64.StdEncoding.EncodeToString([]byte(cleartextAuthString))

	return base64AuthString
}

func getDockerConfigRegistryBlock(registry *DockerConfigRegistryInfo) DockerConfigEntry {
	entry := DockerConfigEntry{
		Username: registry.Username,
		Password: registry.Password,
		Auth:     getAuthStringFromUsernamePassword(registry.Username, registry.Password),
	}

	return entry
}

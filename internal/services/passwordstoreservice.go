package services

import (
	"crypto/md5"
	"encoding/hex"
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

type PasswordStoreService struct {
	secretsNamespace string
	secretNamePrefix string
	logger           zerolog.Logger
	clientConfig     *rest.Config
}

func NewPasswordStoreService(kconfig *rest.Config, operatorConfig *config.ArtifactoryOperatorConfig) *PasswordStoreService {
	result := &PasswordStoreService{
		secretsNamespace: operatorConfig.PasswordStoreBackendNamespace,
		secretNamePrefix: operatorConfig.PasswordStoreSecretNamePrefix,
		logger:           utils.Log.With().Str("service", "passwordstore").Logger(),
		clientConfig:     kconfig,
	}
	return result
}

func (s *PasswordStoreService) GetUserPassword(username string) (password string, err error) {

	coreClient, err := kubernetes.NewForConfig(s.clientConfig)

	if err != nil {
		s.logger.Info().Err(err).Msg("Cannot get kubernetes core API client.")
		return "", err
	}

	secret := s.newPasswordSecretForUser(username)

	_, err = coreClient.CoreV1().Secrets(s.secretsNamespace).Create(secret)

	if err != nil {
		// If the object already exists, we get it to extract the password from it
		if k8serrors.IsAlreadyExists(err) {
			secret, err = coreClient.CoreV1().Secrets(s.secretsNamespace).Get(secret.ObjectMeta.Name, metav1.GetOptions{})
			if err != nil {
				s.logger.Info().Err(err).Str("namespace", s.secretsNamespace).Msgf("Cannot get existing user secret : %s.", secret.ObjectMeta.Name)
				return "", err
			}
		} else {
			s.logger.Info().Err(err).Str("namespace", s.secretsNamespace).Msgf("Cannot create user secret : %s.", secret.ObjectMeta.Name)
			return "", err
		}
	}

	password, err = s.extractPasswordFromSecret(secret)

	return password, nil
}

func (s *PasswordStoreService) extractPasswordFromSecret(secret *v1.Secret) (password string, err error) {
	passwordBytes, keyExists := secret.Data["password"]
	if !keyExists {
		s.logger.Info().Err(err).Str("namespace", s.secretsNamespace).Msgf("Invalid user secret '%s', does not contain password.", secret.ObjectMeta.Name)
		return "", err
	}

	return string(passwordBytes[:]), nil

}

func (s *PasswordStoreService) getSecretNameForUser(username string) string {
	hashedUsername := md5.Sum([]byte(username))
	usernameSuffix := hex.EncodeToString(hashedUsername[:])

	result := strings.Join(
		[]string{
			s.secretNamePrefix,
			usernameSuffix,
		},
		"-")

	return result
}

func (s *PasswordStoreService) newPasswordSecretForUser(username string) (secret *v1.Secret) {
	secretName := s.getSecretNameForUser(username)
	password := utils.GenerateRandomPassword(12)

	secret = &v1.Secret{
		Type: v1.SecretTypeOpaque,

		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Labels: map[string]string{
				"creator": "artifactory-operator",
				"user":    username,
			},
		},

		Data: map[string][]byte{
			"password": []byte(password),
		},
	}

	return secret
}

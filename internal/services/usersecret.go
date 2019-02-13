package services

import (
	"strings"

	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type UserSecretService struct {
	secretsNamespace string
	secretNamePrefix string
	logger           zerolog.Logger
}

func NewUserSecretService() *UserSecretService {
	return &UserSecretService{
		secretsNamespace: "kube-system",
		secretNamePrefix: "artifactory-user",
		logger:           utils.Log.With().Str("service", "usersecret").Logger(),
	}
}

func (s *UserSecretService) GetUserPassword(username string) (Password string, err error) {
	kconfig, err := rest.InClusterConfig()

	if err != nil {
		s.logger.Info().Err(err).Msg("Cannot get service account.")
		return "", err
	}

	coreClient, err := kubernetes.NewForConfig(kconfig)

	if err != nil {
		s.logger.Info().Err(err).Msg("Cannot get kubernetes core API client.")
		return "", err
	}

	secret := s.newPasswordSecretForUser(username)

	secret, err = coreClient.CoreV1().Secrets(s.secretsNamespace).Create(secret)

	if err != nil {
		// If the object already exists, we get it to extract the password from it
		if k8serrors.IsAlreadyExists(err) {
			secret, err := coreClient.CoreV1().Secrets(s.secretsNamespace).Get(secret.ObjectMeta.Name, metav1.GetOptions{})
			if err != nil {
				s.logger.Info().Err(err).Str("namespace", s.secretsNamespace).Msgf("Cannot get existing user secret : %s.", secret.ObjectMeta.Name)
				return "", err
			}
		} else {
			s.logger.Info().Err(err).Str("namespace", s.secretsNamespace).Msgf("Cannot create user secret : %s.", secret.ObjectMeta.Name)
			return "", err
		}
	}

	password, err := s.extractPasswordFromSecret(secret)

	return password, nil
}

func (s *UserSecretService) extractPasswordFromSecret(secret *v1.Secret) (password string, err error) {
	passwordBytes, keyExists := secret.Data["password"]
	if !keyExists {
		s.logger.Info().Err(err).Str("namespace", s.secretsNamespace).Msgf("Invalid user secret '%s', does not contain password.", secret.ObjectMeta.Name)
		return "", err
	}

	return string(passwordBytes[:]), nil

}

func (s *UserSecretService) getSecretNameForUser(username string) string {
	result := strings.Join(
		[]string{
			s.secretNamePrefix,
			username,
		},
		"_")

	return result
}

func (s *UserSecretService) newPasswordSecretForUser(username string) (secret *v1.Secret) {
	secretName := s.getSecretNameForUser(username)
	password := s.generatePassword()

	secret = &v1.Secret{
		Type: v1.SecretTypeOpaque,

		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Labels: map[string]string{
				"creator": "artifactory-operator",
			},
		},

		Data: map[string][]byte{
			"password": []byte(password),
		},
	}

	return secret
}

func (s *UserSecretService) generatePassword() string {
	// FIXME: just to test the secret creation ;). See how to create a random string later
	return "toto"
}

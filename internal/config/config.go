package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type ArtifactoryOperatorConfig struct {
	ClusterLocation               string
	PasswordStoreBackendNamespace string
	PasswordStoreSecretNamePrefix string
	ArtifactoryServerUrl          string
	ArtifactoryServerUser         string
	ArtifactoryServerPassword     string
}

var AllowedClusterLocations = []string{"intranet", "extranet"}

const (
	DefaultPasswordStoreBackendNamespace = "kube-system"
	DefaultPasswordStoreSecretNamePrefix = "artifactory-user"
)

func LoadConfig() (*ArtifactoryOperatorConfig, error) {
	result := &ArtifactoryOperatorConfig{
		PasswordStoreBackendNamespace: DefaultPasswordStoreBackendNamespace,
		PasswordStoreSecretNamePrefix: DefaultPasswordStoreSecretNamePrefix,
	}

	// TODO : passer par la lib https://github.com/go-ozzo/ozzo-validation pour valider la conf

	clusterLocation, ok := os.LookupEnv("ARTI_OP_CLUSTER_LOCATION")
	if ok {
		clusterLocation, err := validateClusterLocation(clusterLocation)
		if err != nil {
			return nil, err
		}

		result.ClusterLocation = clusterLocation
	} else {
		err := errors.New("ARTI_OP_CLUSTER_LOCATION is required.")
		return nil, err
	}

	passwordStoreBackendNamespace, ok := os.LookupEnv("ARTI_OP_PASSWORDSTORE_BACKEND_NAMESPACE")
	if ok {
		result.PasswordStoreBackendNamespace = passwordStoreBackendNamespace
	}

	passwordStoreSecretNamePrefix, ok := os.LookupEnv("ARTI_OP_PASSWORDSTORE_SECRET_NAME_PREFIX")
	if ok {
		result.PasswordStoreSecretNamePrefix = passwordStoreSecretNamePrefix
	}

	artifactoryServerUrl, ok := os.LookupEnv("ARTI_OP_ARTIFACTORY_SERVER_URL")
	if ok {
		if strings.TrimSpace(artifactoryServerUrl) == "" {
			err := errors.New("ARTI_OP_ARTIFACTORY_SERVER_URL is required, but empty.")
			return nil, err
		}

		result.ArtifactoryServerUrl = artifactoryServerUrl
	} else {
		err := errors.New("ARTI_OP_ARTIFACTORY_SERVER_URL is required.")
		return nil, err
	}

	artifactoryServerUser, ok := os.LookupEnv("ARTI_OP_ARTIFACTORY_SERVER_USER")
	if ok {
		if strings.TrimSpace(artifactoryServerUser) == "" {
			err := errors.New("ARTI_OP_ARTIFACTORY_SERVER_USER is required, but empty.")
			return nil, err
		}

		result.ArtifactoryServerUser = artifactoryServerUser
	} else {
		err := errors.New("ARTI_OP_ARTIFACTORY_SERVER_USER is required.")
		return nil, err
	}

	artifactoryServerPassword, ok := os.LookupEnv("ARTI_OP_ARTIFACTORY_SERVER_PASSWORD")
	if ok {
		if strings.TrimSpace(artifactoryServerPassword) == "" {
			err := errors.New("ARTI_OP_ARTIFACTORY_SERVER_PASSWORD is required, but empty.")
			return nil, err
		}

		result.ArtifactoryServerPassword = artifactoryServerPassword
	} else {
		err := errors.New("ARTI_OP_ARTIFACTORY_SERVER_PASSWORD is required.")
		return nil, err
	}

	return result, nil
}

func validateClusterLocation(location string) (string, error) {
	trimmedClusterLocation := strings.TrimSpace(location)

	for _, candidateLoc := range AllowedClusterLocations {
		if trimmedClusterLocation == candidateLoc {
			return trimmedClusterLocation, nil
		}
	}
	return "", fmt.Errorf("ARTI_OP_CLUSTER_LOCATION is required and must be one of %v.", AllowedClusterLocations)
}

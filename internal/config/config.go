package config

import (
	"errors"
	"os"
	"strings"
)

type ArtifactoryOperatorConfig struct {
	ClusterLocation               string
	PasswordStoreBackendNamespace string
	PasswordStoreSecretNamePrefix string
	ClusterTenant                 *string
	ArtifactoryServerUrl          string
}

const (
	DefaultClusterLocation               = "intranet"
	DefaultPasswordStoreBackendNamespace = "kube-system"
	DefaultPasswordStoreSecretNamePrefix = "artifactory-user"
)

func LoadConfig() (*ArtifactoryOperatorConfig, error) {
	result := &ArtifactoryOperatorConfig{
		ClusterLocation:               DefaultClusterLocation,
		PasswordStoreBackendNamespace: DefaultPasswordStoreBackendNamespace,
		PasswordStoreSecretNamePrefix: DefaultPasswordStoreSecretNamePrefix,
	}

	clusterLocation, ok := os.LookupEnv("ARTI_OP_CLUSTER_LOCATION")
	if ok {
		result.ClusterLocation = clusterLocation
	}

	passwordStoreBackendNamespace, ok := os.LookupEnv("ARTI_OP_PASSWORDSTORE_BACKEND_NAMESPACE")
	if ok {
		result.PasswordStoreBackendNamespace = passwordStoreBackendNamespace
	}

	passwordStoreSecretNamePrefix, ok := os.LookupEnv("ARTI_OP_PASSWORDSTORE_SECRET_NAME_PREFIX")
	if ok {
		result.PasswordStoreSecretNamePrefix = passwordStoreSecretNamePrefix
	}

	clusterTenant, ok := os.LookupEnv("ARTI_OP_CLUSTER_TENANT")
	if ok {
		if strings.TrimSpace(clusterTenant) == "" {
			err := errors.New("ARTI_OP_CLUSTER_TENANT is empty. Undefine the variable if you want it unset.")
			return nil, err
		}

		result.ClusterTenant = &clusterTenant
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

	return result, nil
}

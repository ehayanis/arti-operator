package config

import (
	"errors"
	"fmt"
	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"os"
	"strings"
)

func LoadConfig() (*types.ArtifactoryOperatorConfig, error) {
	result := &types.ArtifactoryOperatorConfig{
		PasswordStoreBackendNamespace: utils.DefaultPasswordStoreBackendNamespace,
		PasswordStoreSecretNamePrefix: utils.DefaultPasswordStoreSecretNamePrefix,
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

	vaultServerUrl, ok := os.LookupEnv("ARTI_OP_VAULT_SERVER_URL")
	if ok {
		if strings.TrimSpace(vaultServerUrl) == "" {
			err := errors.New("ARTI_OP_VAULT_SERVER_URL is required, but empty.")
			return nil, err
		}

		result.VaultServerUrl = vaultServerUrl
	} else {
		err := errors.New("ARTI_OP_VAULT_SERVER_URL is required.")
		return nil, err
	}

	vaultToken, ok := os.LookupEnv("ARTI_OP_VAULT_TOKEN")
	if ok {
		result.VaultServerToken = vaultToken
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

	xrayServerUrl, ok := os.LookupEnv("ARTI_OP_XRAY_SERVER_URL")
	if ok {
		if strings.TrimSpace(xrayServerUrl) == "" {
			err := errors.New("ARTI_OP_XRAY_SERVER_URL is required, but empty.")
			return nil, err
		}

		result.XrayServerUrl = xrayServerUrl
	} else {
		err := errors.New("ARTI_OP_XRAY_SERVER_URL is required.")
		return nil, err
	}

	xrayServerUser, ok := os.LookupEnv("ARTI_OP_XRAY_SERVER_USER")
	if ok {
		if strings.TrimSpace(xrayServerUser) == "" {
			err := errors.New("ARTI_OP_XRAY_SERVER_USER is required, but empty.")
			return nil, err
		}

		result.XrayServerUser = xrayServerUser
	} else {
		err := errors.New("ARTI_OP_XRAY_SERVER_USER is required.")
		return nil, err
	}

	xrayServerPassword, ok := os.LookupEnv("ARTI_OP_XRAY_SERVER_PASSWORD")
	if ok {
		if strings.TrimSpace(xrayServerPassword) == "" {
			err := errors.New("ARTI_OP_XRAY_SERVER_PASSWORD is required, but empty.")
			return nil, err
		}

		result.XrayServerPassword = xrayServerPassword
	} else {
		err := errors.New("ARTI_OP_XRAY_SERVER_PASSWORD is required.")
		return nil, err
	}

	xrayBinMgrID, ok := os.LookupEnv("ARTI_OP_XRAY_BINMGRID")
	if ok {
		if strings.TrimSpace(xrayServerPassword) == "" {
			err := errors.New("ARTI_OP_XRAY_BINMGRID is required, but empty.")
			return nil, err
		}

		result.XrayBinMgrID = xrayBinMgrID
	} else {
		err := errors.New("ARTI_OP_XRAY_BINMGRID is required.")
		return nil, err
	}

	return result, nil
}

func validateClusterLocation(location string) (string, error) {
	trimmedClusterLocation := strings.TrimSpace(location)

	for _, candidateLoc := range utils.AllowedClusterLocations {
		if trimmedClusterLocation == candidateLoc {
			return trimmedClusterLocation, nil
		}
	}
	return "", fmt.Errorf("ARTI_OP_CLUSTER_LOCATION is required and must be one of %v.", utils.AllowedClusterLocations)
}

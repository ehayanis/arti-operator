package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
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
	} else {
		err := errors.New("ARTI_OP_VAULT_TOKEN is required.")
		return nil, err
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

	artifactoryServerToken, ok := os.LookupEnv("ARTI_OP_ARTIFACTORY_SERVER_TOKEN")
	if ok {
		if strings.TrimSpace(artifactoryServerToken) == "" {
			err := errors.New("ARTI_OP_ARTIFACTORY_SERVER_TOKEN is required, but empty.")
			return nil, err
		}

		result.ArtifactoryServerToken = artifactoryServerToken
	} else {
		err := errors.New("ARTI_OP_ARTIFACTORY_SERVER_TOKEN is required.")
		return nil, err
	}

	clusterDNSSubdomain, ok := os.LookupEnv("ARTI_OP_CLUSTER_DNSSUBDOMAIN")
	if ok {
		if strings.TrimSpace(clusterDNSSubdomain) == "" {
			err := errors.New("ARTI_OP_CLUSTER_DNSSUBDOMAIN is required, but empty.")
			return nil, err
		}

		result.ClusterDNSSubdomain = clusterDNSSubdomain
	} else {
		err := errors.New("ARTI_OP_CLUSTER_DNSSUBDOMAIN is required.")
		return nil, err
	}

	appOPS, ok := os.LookupEnv("LDAP_APP_OPS_GROUPBASE")
	if ok {
		if strings.TrimSpace(appOPS) == "" {
			err := errors.New("LDAP_APP_OPS_GROUPBASE is required, but empty.")
			return nil, err
		}

		result.LDAPGroups.AppOPS = appOPS
	} else {
		err := errors.New("LDAP_APP_OPS_GROUPBASE is required.")
		return nil, err
	}

	customerOPS, ok := os.LookupEnv("LDAP_CUSTOMER_OPS_GROUPBASE")
	if ok {
		if strings.TrimSpace(customerOPS) == "" {
			err := errors.New("LDAP_CUSTOMER_OPS_GROUPBASE is required, but empty.")
			return nil, err
		}

		result.LDAPGroups.CustomerOPS = customerOPS
	} else {
		err := errors.New("LDAP_CUSTOMER_OPS_GROUPBASE is required.")
		return nil, err
	}

	viewer, ok := os.LookupEnv("LDAP_VIEWER_GROUPBASE")
	if ok {
		if strings.TrimSpace(viewer) == "" {
			err := errors.New("LDAP_VIEWER_GROUPBASE is required, but empty.")
			return nil, err
		}

		result.LDAPGroups.Viewer = viewer
	} else {
		err := errors.New("LDAP_VIEWER_GROUPBASE is required.")
		return nil, err
	}

	ShareRepository, ok := os.LookupEnv("SHARED_REPOSITORY")
	if ok {
		if strings.TrimSpace(ShareRepository) == "" {
			err := errors.New("SKIP_SHARED_REPOSITORY (bool) is required, but empty.")
			return nil, err
		}

		result.SharedRepository = ShareRepository
	} else {
		err := errors.New("SKIP_SHARED_REPOSITORY is required.")
		return nil, err
	}

	logger := utils.Log.With().Str("service", "project").Logger()
	result.ProjectResyncPeriod = time.Hour * 4
	projectResyncPeriod, ok := os.LookupEnv("ARTI_OP_RESYNC_PERIOD_SECONDS")
	if !ok || strings.TrimSpace(projectResyncPeriod) == "" {
		logger.Warn().Msgf("ARTI_OP_RESYNC_PERIOD_SECONDS is empty, using default value : 14400 (4 hours)")
	} else {
		intProjectResyncPeriod, err := strconv.Atoi(projectResyncPeriod)
		if err != nil {
			logger.Warn().Msgf("ARTI_OP_RESYNC_PERIOD_SECONDS is not valid, using default value : 14400 (4 hours)")
		} else {
			result.ProjectResyncPeriod = time.Second * time.Duration(intProjectResyncPeriod)
		}
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

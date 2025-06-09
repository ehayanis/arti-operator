package config

import (
	"fmt"
	"os"

	"github.com/ca-gip/artifactory-operator/internal/services"
	"github.com/ca-gip/artifactory-operator/internal/types"
	v2 "github.com/ca-gip/artifactory-operator/internal/types/v2"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/ca-gip/kubi/pkg/generated/clientset/versioned"
	"github.com/rs/zerolog"
	"k8s.io/client-go/rest"
)

// InstanciateKubernetesClients creates Kubernetes and Kubi clients
func InstanciateKubernetesClients() (*rest.Config, *versioned.Clientset) {
	kconf, err := utils.GetClientConfig()
	if err != nil {
		fmt.Println("Couldn't load K8S client config:", err)
		os.Exit(1)
	}

	v3, err := versioned.NewForConfig(kconf)
	if err != nil {
		fmt.Println("Couldn't create Kubi client:", err)
		os.Exit(1)
	}
	return kconf, v3
}

// InstanciateServices creates and wires together the various services for v1
func InstanciateServices(kconfig *rest.Config, operatorConfig *types.ArtifactoryOperatorConfig, logger zerolog.Logger) *services.ProjectService {
	passwordStoreService := services.NewPasswordStoreService(kconfig, operatorConfig)

	artifactoryService, err := services.NewArtifactoryService(operatorConfig, passwordStoreService)
	if err != nil {
		logger.Error().Msgf("Couldn't create Artifactory service: %v", err)
	}

	dockerConfigSecretsService := services.NewDockerConfigSecretsService(kconfig)

	projectService, err := services.NewProjectService(operatorConfig, dockerConfigSecretsService, artifactoryService, kconfig)
	if err != nil {
		logger.Error().Msgf("Couldn't create Project service: %v", err)
	}

	return projectService
}

// InstanciateServicesV2 creates and wires together the various services for v2
func InstanciateServicesV2(kconfig *rest.Config, operatorConfig *v2.ArtifactoryOperatorConfigV2, logger zerolog.Logger) (*services.ProjectService, *services.ExternalAPIService) {
	// Convert v2 config to v1 config for backward compatibility
	v1Config := &types.ArtifactoryOperatorConfig{
		ClusterLocation:               operatorConfig.ClusterLocation,
		PasswordStoreBackendNamespace: operatorConfig.PasswordStoreBackendNamespace,
		PasswordStoreSecretNamePrefix: operatorConfig.PasswordStoreSecretNamePrefix,
		ArtifactoryServerUrl:          operatorConfig.ArtifactoryServerUrl,
		ArtifactoryServerUser:         operatorConfig.ArtifactoryServerUser,
		VaultServerToken:              operatorConfig.VaultServerToken,
		VaultServerUrl:                operatorConfig.VaultServerUrl,
		ClusterDNSSubdomain:           operatorConfig.ClusterDNSSubdomain,
		LDAPGroups:                    types.LDAPGroups(operatorConfig.LDAPGroups),
		SharedRepository:              operatorConfig.SharedRepository,
		ArtifactoryServerToken:        operatorConfig.ArtifactoryServerToken,
		ProjectResyncPeriod:           operatorConfig.ProjectResyncPeriod,
	}

	// Create v1 services
	passwordStoreService := services.NewPasswordStoreService(kconfig, v1Config)

	artifactoryService, err := services.NewArtifactoryService(v1Config, passwordStoreService)
	if err != nil {
		logger.Error().Msgf("Couldn't create Artifactory service: %v", err)
	}

	dockerConfigSecretsService := services.NewDockerConfigSecretsService(kconfig)

	projectService, err := services.NewProjectService(v1Config, dockerConfigSecretsService, artifactoryService, kconfig)
	if err != nil {
		logger.Error().Msgf("Couldn't create Project service: %v", err)
	}

	// Create v2-specific services
	var externalAPIService *services.ExternalAPIService
	if operatorConfig.ExternalAPI.Enabled {
		externalAPIService = services.NewExternalAPIService(operatorConfig.ExternalAPI.Endpoint)
		logger.Info().Msgf("External API service created with endpoint: %s", operatorConfig.ExternalAPI.Endpoint)
	} else {
		logger.Info().Msgf("External API service disabled")
	}

	return projectService, externalAPIService
}
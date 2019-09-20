package config

import (
	"fmt"
	"github.com/ca-gip/artifactory-operator/internal/services"
	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/ca-gip/kubi/pkg/client/clientset/versioned"
	"github.com/rs/zerolog"
	"k8s.io/client-go/rest"
	"os"
)

func InstanciateKubernetesClients() (*rest.Config, *versioned.Clientset) {
	kconfig, err := utils.GetClientConfig()
	if err != nil {
		fmt.Println("Couldn't load K8S client config:", err)
		os.Exit(1)
	}
	v3, err := versioned.NewForConfig(kconfig)
	if err != nil {
		fmt.Println("Couldn't create Kubi client:", err)
		os.Exit(1)
	}
	return kconfig, v3
}

func InstanciateServices(kconfig *rest.Config, operatorConfig *types.ArtifactoryOperatorConfig, logger zerolog.Logger) *services.ProjectService {

	passwordStoreService := services.NewPasswordStoreService(kconfig, operatorConfig)

	artifactoryService, err := services.NewArtifactoryService(operatorConfig, passwordStoreService)
	if err != nil {
		logger.Error().Msgf("Couldn't create Artifactory service: %v", err)
	}

	xrayService, err := services.NewXrayService(operatorConfig)
	if err != nil {
		logger.Error().Msgf("Couldn't create Xray service: %v", err)
	}

	dockerConfigSecretsService := services.NewDockerConfigSecretsService(kconfig)

	projectService, err := services.NewProjectService(operatorConfig, dockerConfigSecretsService, artifactoryService, xrayService, kconfig)
	if err != nil {
		logger.Error().Msgf("Couldn't create Project service: %v", err)
	}

	return projectService
}

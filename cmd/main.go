package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ca-gip/artifactory-operator/internal/config"
	"github.com/ca-gip/artifactory-operator/internal/services"

	"github.com/ca-gip/artifactory-operator/internal/utils"
	v1 "github.com/ca-gip/kubi/pkg/apis/ca-gip/v1"
	"github.com/ca-gip/kubi/pkg/client/clientset/versioned"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

func main() {

	WatchProjects()
}

type operatorServices struct {
	passwordStoreService       *services.PasswordStoreService
	dockerConfigSecretsService *services.DockerConfigSecretsService
}

// Watch NetworkPolicyConfig, which is a config object for namespace network bubble
// This CRD allow user to deploy global configuration for network configuration
// for update, the default network config is update
// for deletion, it is automatically recreated
// for create, just create it
func WatchProjects() cache.Store {
	logger := utils.Log.With().Str("service", "watcher").Logger()

	operatorConfig, err := config.LoadConfig()

	if err != nil {
		fmt.Println("Couldn't load operator configuration:", err)
		os.Exit(1)
	}

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

	passwordStoreService := services.NewPasswordStoreService(kconfig, operatorConfig)
	artifactoryService, err := services.NewArtifactoryService(operatorConfig, passwordStoreService)
	if err != nil {
		logger.Error().Msgf("Couldn't create Artifactory service: %v", err)
	}

	dockerConfigSecretsService := services.NewDockerConfigSecretsService(kconfig, operatorConfig)

	projectService, err := services.NewProjectService(operatorConfig, dockerConfigSecretsService, artifactoryService)
	if err != nil {
		logger.Error().Msgf("Couldn't create Project service: %v", err)
	}

	watchlist := cache.NewListWatchFromClient(v3.CagipV1().RESTClient(), "projects", v12.NamespaceAll, fields.Everything())
	resyncPeriod := 30 * time.Minute

	store, controller := cache.NewInformer(watchlist, &v1.Project{}, resyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			projectCreated(obj, projectService)
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			projectUpdate(old, new, projectService)
		},
	})

	logger.Info().Msgf("Operator configuration: %v", operatorConfig)
	controller.Run(wait.NeverStop)
	logger.Info().Msgf("Controller exited, terminating.")

	return store
}

func projectUpdate(old interface{}, new interface{}, projectService *services.ProjectService) {
	newProject := new.(*v1.Project)

	err := projectService.HandleProject(newProject)
	if err != nil {
		utils.Log.Error().Msgf("Error when creating assets in artifactory: %v", err)
	} else {
		utils.Log.Info().Msgf("Operator: the project %v has been updated, updating associated resources: artifactory repositories, users, groups and permissions.", newProject.Name)
	}

}

func projectCreated(obj interface{}, projectService *services.ProjectService) {
	project := obj.(*v1.Project)
	err := projectService.HandleProject(project)
	if err != nil {
		utils.Log.Error().Msgf("Error when creating assets in artifactory: %v", err)
	} else {
		utils.Log.Info().Msgf("Operator: the project %v has been created, generating associated resources: artifactory repositories, users, groups and permissions.", project.Name)
	}
}

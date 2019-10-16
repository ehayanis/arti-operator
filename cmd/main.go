package main

import (
	"fmt"
	"github.com/ca-gip/artifactory-operator/internal/types"
	"goji.io"
	"goji.io/pat"
	"net/http"
	"os"
	"time"

	"github.com/ca-gip/artifactory-operator/internal/config"
	"github.com/ca-gip/artifactory-operator/pkg/route"
	"github.com/ca-gip/artifactory-operator/internal/services"

	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/ca-gip/kubi/pkg/apis/ca-gip/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

func debugHandler(next http.Handler) http.Handler {
	logger := utils.Log.With().Str("service", "watcher").Logger()

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger.Info().Msgf("%s %s", r.Method, r.URL)
			next.ServeHTTP(w, r)
		})
}

func main() {

	operatorConfig, err := config.LoadConfig()

	if err != nil {
		fmt.Println("Couldn't load operator configuration:", err)
		os.Exit(1)
	}

	mux := goji.NewMux()
	mux.Use(debugHandler)
	mux.HandleFunc(pat.Get("/healthz"), route.Healthz)

	go http.ListenAndServe(":8080", mux)

	WatchProjects(operatorConfig)
}

// WatchProjects is going to instanciate Kubernetes Client and Services (Project, Artifactory and Xray).
// It is going to listen the Projects CRD (on creation and updates) and create or update resources
// (artifactory client, vault secrets, xray policies and watches) through their services
func WatchProjects(operatorConfig *types.ArtifactoryOperatorConfig) cache.Store {

	logger := utils.Log.With().Str("service", "watcher").Logger()

	kconfig, v3 := config.InstanciateKubernetesClients()

	projectService := config.InstanciateServices(kconfig, operatorConfig, logger)

	watchlist := cache.NewListWatchFromClient(v3.CagipV1().RESTClient(), "projects", v12.NamespaceAll, fields.Everything())
	resyncPeriod := 30 * time.Minute

	store, controller := cache.NewInformer(watchlist, &v1.Project{}, resyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			projectCreated(obj, projectService)
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			projectUpdate(new, projectService)
		},
	})

	logger.Info().Msgf("[Operator configuration] ClusterLocation:%v,"+
		"PasswordBackendNamespace:%v,"+
		"ArtifactoryServerUrl:%v,"+
		"ArtifactoryServerUser:%v,"+
		"VaultServerUrl:%v",
		operatorConfig.ClusterLocation,
		operatorConfig.PasswordStoreBackendNamespace,
		operatorConfig.ArtifactoryServerUrl,
		operatorConfig.ArtifactoryServerUser,
		operatorConfig.VaultServerUrl)

	controller.Run(wait.NeverStop)
	logger.Info().Msgf("Controller exited, terminating.")

	return store
}

func projectUpdate(new interface{}, projectService *services.ProjectService) {
	newProject := new.(*v1.Project)

	err := utils.CheckMandatoryParameters(newProject)
	if err != nil {
		utils.Log.Error().Msgf("Error, project resource does not have mandatory parameter to fill Artifactory: %v", err)
		return
	}

	err = projectService.HandleProject(newProject)
	if err != nil {
		utils.Log.Error().Msgf("Error when creating assets in artifactory: %v", err)
	} else {
		utils.Log.Info().Msgf("Operator: the project %v has been updated, updating associated resources: artifactory repositories, users, groups and permissions.", newProject.Name)
	}

}

func projectCreated(obj interface{}, projectService *services.ProjectService) {
	project := obj.(*v1.Project)

	err := utils.CheckMandatoryParameters(project)
	if err != nil {
		utils.Log.Error().Msgf("Error, project resource does not have mandatory parameter to fill Artifactory: %v", err)
		return
	}

	err = projectService.HandleProject(project)
	if err != nil {
		utils.Log.Error().Msgf("Error when creating assets in artifactory: %v", err)
	} else {
		utils.Log.Info().Msgf("Operator: the project %v has been created, generating associated resources: artifactory repositories, users, groups and permissions.", project.Name)
	}
}
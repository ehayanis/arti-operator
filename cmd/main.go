package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ca-gip/artifactory-operator/internal/types"
	v2types "github.com/ca-gip/artifactory-operator/internal/types/v2"
	"github.com/rs/zerolog"
	"goji.io"
	"goji.io/pat"

	"github.com/ca-gip/artifactory-operator/internal/config"
	"github.com/ca-gip/artifactory-operator/internal/services"
	"github.com/ca-gip/artifactory-operator/pkg/route"

	"github.com/ca-gip/artifactory-operator/internal/utils"
	v1 "github.com/ca-gip/kubi/pkg/apis/cagip/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

// Global variables for v2 support
var (
	useV2              bool
	externalAPIService *services.ExternalAPIService
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
	// Get Level debug configurable if DEBUG env var exist (don't care about the value)
	_, debug := os.LookupEnv("DEBUG")
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Check if v2 is enabled
	v2Enabled, _ := os.LookupEnv("ARTI_OP_ENABLE_V2")
	useV2 = v2Enabled == "true"

	var operatorConfig *types.ArtifactoryOperatorConfig
	var operatorConfigV2 *v2types.ArtifactoryOperatorConfigV2
	var err error

	if useV2 {
		// Load v2 configuration
		operatorConfigV2, err = config.LoadConfigV2()
		if err != nil {
			fmt.Println("Couldn't load v2 operator configuration:", err)
			os.Exit(1)
		}
		
		// For backward compatibility, also load v1 configuration
		operatorConfig, err = config.LoadConfig()
		if err != nil {
			fmt.Println("Couldn't load v1 operator configuration:", err)
			os.Exit(1)
		}
	} else {
		// Load only v1 configuration
		operatorConfig, err = config.LoadConfig()
		if err != nil {
			fmt.Println("Couldn't load operator configuration:", err)
			os.Exit(1)
		}
	}

	mux := goji.NewMux()
	mux.Use(debugHandler)
	mux.HandleFunc(pat.Get("/healthz"), route.Healthz)

	go http.ListenAndServe(":8080", mux)

	if useV2 {
		WatchProjectsV2(operatorConfigV2)
	} else {
		WatchProjects(operatorConfig)
	}
}

// WatchProjects is going to instanciate Kubernetes Client and Services (Project, Artifactory and Xray).
// It is going to listen the Projects CRD (on creation and updates) and create or update resources
// (artifactory client, vault secrets) through their services
func WatchProjects(operatorConfig *types.ArtifactoryOperatorConfig) cache.Store {
	logger := utils.Log.With().Str("service", "watcher").Logger()

	kconfig, v3 := config.InstanciateKubernetesClients()

	projectService := config.InstanciateServices(kconfig, operatorConfig, logger)

	watchlist := cache.NewListWatchFromClient(v3.CagipV1().RESTClient(), "projects", v12.NamespaceAll, fields.Everything())

	store, controller := cache.NewInformer(watchlist, &v1.Project{}, operatorConfig.ProjectResyncPeriod, cache.ResourceEventHandlerFuncs{
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

// WatchProjectsV2 is similar to WatchProjects but uses v2 configuration and services
func WatchProjectsV2(operatorConfig *v2types.ArtifactoryOperatorConfigV2) cache.Store {
	logger := utils.Log.With().Str("service", "watcher").Logger()

	kconfig, v3 := config.InstanciateKubernetesClients()

	projectService, apiService := config.InstanciateServicesV2(kconfig, operatorConfig, logger)
	externalAPIService = apiService // Store in global variable for access in handlers

	watchlist := cache.NewListWatchFromClient(v3.CagipV1().RESTClient(), "projects", v12.NamespaceAll, fields.Everything())

	store, controller := cache.NewInformer(watchlist, &v1.Project{}, operatorConfig.ProjectResyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			projectCreatedV2(obj, projectService)
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			projectUpdateV2(new, projectService)
		},
	})

	logger.Info().Msgf("[Operator V2 configuration] ClusterLocation:%v,"+
		"PasswordBackendNamespace:%v,"+
		"ArtifactoryServerUrl:%v,"+
		"ArtifactoryServerUser:%v,"+
		"VaultServerUrl:%v,"+
		"ExternalAPI.Enabled:%v,"+
		"ExternalAPI.Endpoint:%v",
		operatorConfig.ClusterLocation,
		operatorConfig.PasswordStoreBackendNamespace,
		operatorConfig.ArtifactoryServerUrl,
		operatorConfig.ArtifactoryServerUser,
		operatorConfig.VaultServerUrl,
		operatorConfig.ExternalAPI.Enabled,
		operatorConfig.ExternalAPI.Endpoint)

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
}

func projectCreated(obj interface{}, projectService *services.ProjectService) {
	project := obj.(*v1.Project)
	err := utils.CheckMandatoryParameters(project)
	if err != nil {
		utils.Log.Error().Msgf("Error, project resource does not have mandatory parameter to fill Artifactory: %v", err)
		return
	}

	err = projectService.HandleProject(project)
}

// projectUpdateV2 handles updates to Project resources for v2
func projectUpdateV2(new interface{}, projectService *services.ProjectService) {
	newProject := new.(*v1.Project)
	
	// Check if this is a v2 project
	if utils.IsV2Project(newProject) {
		// Use v2 validation
		err := utils.CheckMandatoryParametersV2(newProject)
		if err != nil {
			utils.Log.Error().Msgf("Error, v2 project resource does not have mandatory parameters: %v", err)
			return
		}
		
		// Call external API if enabled
		if externalAPIService != nil {
			err := externalAPIService.CallExternalAPI(newProject.Spec.Tenant, newProject.Spec.Project)
			if err != nil {
				utils.Log.Error().Msgf("Error calling external API for project %s: %v", newProject.Name, err)
				// Continue with Artifactory operations even if API call fails
			}
		}
	} else {
		// Use v1 validation for backward compatibility
		err := utils.CheckMandatoryParameters(newProject)
		if err != nil {
			utils.Log.Error().Msgf("Error, project resource does not have mandatory parameter to fill Artifactory: %v", err)
			return
		}
	}
	
	// Always handle the project with Artifactory for both v1 and v2
	err := projectService.HandleProject(newProject)
	if err != nil {
		utils.Log.Error().Msgf("Error handling project %s: %v", newProject.Name, err)
	}
}

// projectCreatedV2 handles creation of Project resources for v2
func projectCreatedV2(obj interface{}, projectService *services.ProjectService) {
	project := obj.(*v1.Project)
	
	// Check if this is a v2 project
	if utils.IsV2Project(project) {
		// Use v2 validation
		err := utils.CheckMandatoryParametersV2(project)
		if err != nil {
			utils.Log.Error().Msgf("Error, v2 project resource does not have mandatory parameters: %v", err)
			return
		}
		
		// Call external API if enabled
		if externalAPIService != nil {
			err := externalAPIService.CallExternalAPI(project.Spec.Tenant, project.Spec.Project)
			if err != nil {
				utils.Log.Error().Msgf("Error calling external API for project %s: %v", project.Name, err)
				// Continue with Artifactory operations even if API call fails
			}
		}
	} else {
		// Use v1 validation for backward compatibility
		err := utils.CheckMandatoryParameters(project)
		if err != nil {
			utils.Log.Error().Msgf("Error, project resource does not have mandatory parameter to fill Artifactory: %v", err)
			return
		}
	}
	
	// Always handle the project with Artifactory for both v1 and v2
	err := projectService.HandleProject(project)
	if err != nil {
		utils.Log.Error().Msgf("Error handling project %s: %v", project.Name, err)
	}
}
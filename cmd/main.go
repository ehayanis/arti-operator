package cmd

import (
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/ca-gip/kubi/pkg/apis/ca-gip/v1"
	"github.com/ca-gip/kubi/pkg/client/clientset/versioned"
	v12 "k8s.io/api/core/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"time"
)

func main() {

	WatchProjects()
}

// Watch NetworkPolicyConfig, which is a config object for namespace network bubble
// This CRD allow user to deploy global configuration for network configuration
// for update, the default network config is update
// for deletion, it is automatically recreated
// for create, just create it
func WatchProjects() cache.Store {
	kconfig, _ := rest.InClusterConfig()

	v3, _ := versioned.NewForConfig(kconfig)

	watchlist := cache.NewListWatchFromClient(v3.CagipV1().RESTClient(), "projects", v12.NamespaceAll, fields.Everything())
	resyncPeriod := 30 * time.Minute

	store, controller := cache.NewInformer(watchlist, &v1.Project{}, resyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc:    projectCreated,
		UpdateFunc: projectUpdate,
	})

	go controller.Run(wait.NeverStop)

	return store
}

func projectUpdate(old interface{}, new interface{}) {
	newProject := new.(*v1.Project)

	clientSet, err := kubernetes.NewForConfig(kconfig)
	obj, err := clientSet.CoreV1().Secrets("ij").Get("ij", v13.GetOptions{})



	utils.Log.Info().Msgf("Operator: the project %v has been updated, updating associated resources: namespace, networkpolicies.", newProject.Name)

}

func projectCreated(obj interface{}) {
	project := obj.(*v1.Project)
	utils.Log.Info().Msgf("Operator: the project %v has been created, generating associated resources: namespace, networkpolicies.", project.Name)

}

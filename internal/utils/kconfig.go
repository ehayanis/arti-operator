package utils

import (
	"errors"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"
)

func GetClientConfig() (*rest.Config, error) {
	kconfig, err := rest.InClusterConfig()

	if err == nil {
		return kconfig, nil
	}

	if err != rest.ErrNotInCluster {
		return nil, err
	}

	homeDir, err := homeDir()

	if err != nil {
		return nil, err
	}

	kubeConfigPath := filepath.Join(homeDir, ".kube", "config")

	// use the current context in kubeconfig
	kconfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return kconfig, nil
}

func homeDir() (string, error) {
	if h := os.Getenv("HOME"); h != "" {
		return h, nil
	}
	if h := os.Getenv("USERPROFILE"); h != "" { // windows
		return h, nil
	}

	return "", errors.New("Couldn't determine user's HOME directory.")

}

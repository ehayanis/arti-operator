package main

import (
	"fmt"
	"os"

	"github.com/ca-gip/artifactory-operator/internal/services"
	"github.com/ca-gip/artifactory-operator/internal/utils"
)

func main() {
	kconfig, err := utils.GetClientConfig()

	if err != nil {
		fmt.Println("Couldn't load K8S client config:", err)
		os.Exit(1)
	}

	namespace := "trololo"

	if len(os.Args) > 1 {
		namespace = os.Args[1]
	}

	ipsService := services.NewDockerConfigSecretsService(kconfig)

	secretInput := &services.DockerConfigSecret{
		Name: "myips",
		Registries: []services.DockerConfigRegistryInfo{
			services.DockerConfigRegistryInfo{
				Url:      "trololo.com",
				Username: "mat",
				Password: "RevePasJeVaisLeChanger",
			},
			services.DockerConfigRegistryInfo{
				Url:      "scm.sws.cagip.gca",
				Username: "garstecki_m",
				Password: "UnMotdePasse2",
			},
		},
	}

	secret, err := ipsService.CreateOrUpdateDockerConfigSecret(namespace, secretInput)

	if err != nil {
		fmt.Println("Error while generating secret:", err)
		os.Exit(1)
	}
	fmt.Println("Put secret :")
	fmt.Println("---------------------------------------------------")
	fmt.Println(secret)
	fmt.Println("---------------------------------------------------")
}

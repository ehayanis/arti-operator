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

	namespace := os.Args[1]

	ipsService := services.NewImagePullSecretsService(kconfig)

	secretInput := &services.ImagePullSecret{
		Name: "myips",
		Registries: []services.ImagePullSecretRegistry{
			services.ImagePullSecretRegistry{
				Url:      "trololo.com",
				Username: "mat",
				Password: "RevePasJeVaisLeChanger",
			},
			services.ImagePullSecretRegistry{
				Url:      "scm.sws.cagip.gca",
				Username: "garstecki_m",
				Password: "UnMotdePasse2",
			},
		},
	}

	secret, err := ipsService.CreateOrUpdateImagePullSecret(namespace, secretInput)

	if err != nil {
		fmt.Println("Error while generating secret:", err)
		os.Exit(1)
	}
	fmt.Println(secret)
}

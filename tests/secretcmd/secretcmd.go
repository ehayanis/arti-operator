package main

import (
	"fmt"
	"os"

	"github.com/ca-gip/artifactory-operator/internal/services"
	"github.com/ca-gip/artifactory-operator/internal/utils"
)

func main() {
	username := "trololo"

	if len(os.Args) >= 2 {
		username = os.Args[1]
	}

	kconfig, err := utils.GetClientConfig()

	if err != nil {
		fmt.Errorf("Couldn't load K8S client config:", err)
		os.Exit(1)
	}

	secretService := services.NewPasswordStoreService(kconfig)

	password, err := secretService.GetUserPassword(username)

	if err != nil {
		fmt.Errorf("Error retrieving user password:", err)
		os.Exit(1)
	}

	fmt.Println("Got password:", password)

	password, err = secretService.GetUserPassword(username)

	if err != nil {
		fmt.Errorf("Error retrieving user password:", err)
		os.Exit(1)
	}

	fmt.Println("Got password:", password)
}

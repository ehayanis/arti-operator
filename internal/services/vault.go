package services

import (
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/hashicorp/vault/api"
)

func VaultConnect(vaultURL string, vaultBasePathToken string) (*api.Client, error) {

	config := &api.Config{
		Address: vaultURL,
	}

	client, err := api.NewClient(config)
	if err != nil {
		utils.Log.Error().Err(err)
		return nil, err
	}

	client.SetToken(vaultBasePathToken)

	return client, nil
}

func VaultWriteSecret(c *api.Client, secretData map[string]interface{}, path string) (*api.Secret, error) {

	secret, err := c.Logical().Write(path, secretData)
	if err != nil {
		utils.Log.Error().Err(err)
		return nil, err
	}
	return secret, err
}

func VaultReadSecret(c *api.Client, path string) (*api.Secret, error) {

	secret, err := c.Logical().Read(path)
	if err != nil {
		utils.Log.Error().Err(err)
		return nil, err
	}
	return secret, err
}

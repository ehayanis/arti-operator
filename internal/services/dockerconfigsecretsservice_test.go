package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateSecretObject(t *testing.T) {
	input := &DockerConfigSecret{
		Name: "test-secret",
		Registries: []DockerConfigRegistryInfo{
			DockerConfigRegistryInfo{
				Url:      "registry1.registries.example.com",
				Username: "registry1_user",
				Password: "registry1_password",
			},
			DockerConfigRegistryInfo{
				Url:      "registry2.registries.example.com",
				Username: "registry2_USER",
				Password: "registry2_PASSWORD",
			},
		},
	}

	// Base64 auths generated in CLI : echo -n "USER:PASSWORD" | base64
	expectedJson := `
	{ "auths": {
			"registry1.registries.example.com": {
				"auth": "cmVnaXN0cnkxX3VzZXI6cmVnaXN0cnkxX3Bhc3N3b3Jk",
				"username": "registry1_user",
				"password": "registry1_password"
			},
			"registry2.registries.example.com": {
				"auth": "cmVnaXN0cnkyX1VTRVI6cmVnaXN0cnkyX1BBU1NXT1JE",
				"username": "registry2_USER",
				"password": "registry2_PASSWORD"
			}
		}
	}
	`

	result, err := generateSecretObject(input)

	if err != nil {
		t.Errorf("Got error from generateSecretObject: %v", err)
	}

	if strings.Compare(result.ObjectMeta.Name, "test-secret") != 0 {
		t.Errorf("Expected ObjectMeta.Name '%v', got '%v'.", "test-secret", result.ObjectMeta.Name)
	}

	assert.JSONEq(t, expectedJson, string(result.Data[".dockerconfigjson"]), "Generated docker config does not match reference.")
}

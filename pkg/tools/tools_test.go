package tools

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRepoToStructW(t *testing.T) {
	t.Run("withDashCharacter", func(r *testing.T) {
		reponame := "cacf-sofinco-api-docker-stable-intranet"
		repoFields, _ := RepoToStruct(reponame)

		assert.Equal(t, repoFields.Tenant, "cacf")
		assert.Equal(t, repoFields.Locator, "intranet")
		assert.Equal(t, repoFields.ProjectName, "sofinco-api")
		assert.Equal(t, repoFields.Stage, "stable")
	})

	t.Run("withMultipleDashCharacter", func(r *testing.T) {
		reponame := "cacf-sofinco-api-whatever-docker-stable-intranet"
		repoFields, _ := RepoToStruct(reponame)

		assert.Equal(t, repoFields.Tenant, "cacf")
		assert.Equal(t, repoFields.Locator, "intranet")
		assert.Equal(t, repoFields.ProjectName, "sofinco-api-whatever")
		assert.Equal(t, repoFields.Stage, "stable")
	})

	t.Run("withoutDashCharacter", func(r *testing.T) {
		reponame := "cacf-sofinco-docker-stable-intranet"
		repoFields, _ := RepoToStruct(reponame)

		assert.NotNil(t, repoFields)
		assert.Equal(t, repoFields.Tenant, "cacf")
		assert.Equal(t, repoFields.Locator, "intranet")
		assert.Equal(t, repoFields.ProjectName, "sofinco")
		assert.Equal(t, repoFields.Stage, "stable")
	})
}

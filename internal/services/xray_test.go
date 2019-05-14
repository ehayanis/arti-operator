package services

import (
	"github.com/ca-gip/artifactory-operator/internal/types"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func PrepareOperatorConfigStruct() *types.ArtifactoryOperatorConfig {
	return &types.ArtifactoryOperatorConfig{
		ClusterLocation:               utils.ArtifactoryLocationIntranet,
		PasswordStoreBackendNamespace: utils.DefaultPasswordStoreBackendNamespace,
		PasswordStoreSecretNamePrefix: utils.DefaultPasswordStoreSecretNamePrefix,
		ArtifactoryServerUrl:          "",
		ArtifactoryServerUser:         "",
		ArtifactoryServerPassword:     "",
		VaultServerToken:              "",
		VaultServerUrl:                "",
		XrayServerUrl:                 "",
		XrayServerUser:                "",
		XrayServerPassword:            "",
		XrayBinMgrID:                  utils.XrayTestBinMgrID,
	}
}

func TestPolicyFieldsforScratchRepo(t *testing.T) {
	xrayService := &XrayService{
		logger:       utils.Log.With().Str("service", "xray").Logger(),
		xrayClient:   nil,
		xrayBinMgrID: PrepareOperatorConfigStruct().XrayBinMgrID,
	}

	repoName := utils.XrayTestRepoScratch

	result := xrayService.policyFields(repoName)
	rules := result.Rules

	assert.Equal(t, repoName, result.Name)
	assert.Equal(t, "security", result.Type)
	assert.Equal(t, "High", rules[0].Criteria.MinSeverity)
	assert.Equal(t, false, rules[0].Actions.BlockDownload.Active)
	assert.Equal(t, false, rules[0].Actions.BlockDownload.Unscanned)
}

func TestPolicyFieldsforStagingRepo(t *testing.T) {
	xrayService := &XrayService{
		logger:       utils.Log.With().Str("service", "xray").Logger(),
		xrayClient:   nil,
		xrayBinMgrID: PrepareOperatorConfigStruct().XrayBinMgrID,
	}

	repoName := utils.XrayTestRepoStaging

	result := xrayService.policyFields(repoName)
	rules := result.Rules

	assert.Equal(t, repoName, result.Name)
	assert.Equal(t, "security", result.Type)
	assert.Equal(t, "High", rules[0].Criteria.MinSeverity)
	assert.Equal(t, false, rules[0].Actions.BlockDownload.Active)
	assert.Equal(t, false, rules[0].Actions.BlockDownload.Unscanned)
}

func TestPolicyFieldsforStableRepo(t *testing.T) {
	xrayService := &XrayService{
		logger:       utils.Log.With().Str("service", "xray").Logger(),
		xrayClient:   nil,
		xrayBinMgrID: PrepareOperatorConfigStruct().XrayBinMgrID,
	}

	repoName := utils.XrayTestRepoStable

	result := xrayService.policyFields(repoName)
	rules := result.Rules

	assert.Equal(t, repoName, result.Name)
	assert.Equal(t, "security", result.Type)
	assert.Equal(t, "High", rules[0].Criteria.MinSeverity)
	assert.Equal(t, true, rules[0].Actions.BlockDownload.Active)
	assert.Equal(t, true, rules[0].Actions.BlockDownload.Unscanned)
}

func TestWatchFieldsforScratchRepo(t *testing.T) {
	xrayService := &XrayService{
		logger:       utils.Log.With().Str("service", "xray").Logger(),
		xrayClient:   nil,
		xrayBinMgrID: PrepareOperatorConfigStruct().XrayBinMgrID,
	}

	repoName := utils.XrayTestRepoScratch
	result := xrayService.watchFields(repoName)

	assert.Equal(t, repoName, result.GeneralData.Name)
	assert.Equal(t, "repository", result.ProjectResources.Resources[0].Type)
	assert.Equal(t, xrayService.xrayBinMgrID, result.ProjectResources.Resources[0].BinMgrID)
	assert.Equal(t, "local", result.ProjectResources.Resources[0].RepoType)
	assert.Equal(t, repoName, result.ProjectResources.Resources[0].Name)
	assert.Equal(t, repoName, result.AssignedPolicies[0].Name)
	assert.Equal(t, "security", result.AssignedPolicies[0].Type)
}

func TestCheckKindOfRepoForScratchRepo(t *testing.T) {

	xrayService := &XrayService{
		logger:       utils.Log.With().Str("service", "xray").Logger(),
		xrayClient:   nil,
		xrayBinMgrID: PrepareOperatorConfigStruct().XrayBinMgrID,
	}

	repoName := utils.XrayTestRepoScratch

	result := xrayService.checkKindofRepo(repoName)

	assert.Equal(t, false, result, "Regexp should return false")
}

func TestCheckKindOfRepoForStagingRepo(t *testing.T) {

	xrayService := &XrayService{
		logger:       utils.Log.With().Str("service", "xray").Logger(),
		xrayClient:   nil,
		xrayBinMgrID: PrepareOperatorConfigStruct().XrayBinMgrID,
	}

	repoName := utils.XrayTestRepoStaging

	result := xrayService.checkKindofRepo(repoName)

	assert.Equal(t, false, result, "Regexp should return false")
}

func TestCheckKindOfRepoForStableRepo(t *testing.T) {

	xrayService := &XrayService{
		logger:       utils.Log.With().Str("service", "xray").Logger(),
		xrayClient:   nil,
		xrayBinMgrID: PrepareOperatorConfigStruct().XrayBinMgrID,
	}

	repoName := utils.XrayTestRepoStable

	result := xrayService.checkKindofRepo(repoName)

	assert.Equal(t, true, result, "Regexp should return true")
}

package v2

import "time"

// ArtifactoryInformation contains information about a project in Artifactory
type ArtifactoryInformation struct {
	Tenant              string
	ProjectName         string   // cr
	Stages              []string // cr
	Location            string   //param confmap
	Description         string
	Environment         string
	SourceDN            string
	ClusterDNSSubdomain string
	// Flag to indicate if this is a v2 project
	IsV2 bool
}

// ArtifactoryRepoUsers stores user credentials for accessing Artifactory repositories
type ArtifactoryRepoUsers struct {
	UserNameRO string
	PasswordRO string
	UserNameRW string
	PasswordRW string
}

// ExternalAPIConfig contains configuration for the external API
type ExternalAPIConfig struct {
	Enabled  bool
	Endpoint string
}

// ArtifactoryOperatorConfigV2 extends the v1 config with v2-specific fields
type ArtifactoryOperatorConfigV2 struct {
	ClusterLocation               string
	PasswordStoreBackendNamespace string
	PasswordStoreSecretNamePrefix string
	ArtifactoryServerUrl          string
	ArtifactoryServerUser         string
	VaultServerToken              string
	VaultServerUrl                string
	ClusterDNSSubdomain           string
	LDAPGroups                    LDAPGroups
	SharedRepository              string
	ArtifactoryServerToken        string
	ProjectResyncPeriod           time.Duration
	// V2-specific fields
	ExternalAPI       ExternalAPIConfig
	DockerRegistryURL string // Base URL for Docker registry
}

// LDAPGroups contains LDAP group information for different roles
type LDAPGroups struct {
	AppOPS      string
	CustomerOPS string
	Viewer      string
}

// ExternalAPIRequest represents the request to the external API
type ExternalAPIRequest struct {
	Tenant  string `json:"tenant"`
	Project string `json:"project"`
}

// ExternalAPIResponse represents the response from the external API
type ExternalAPIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

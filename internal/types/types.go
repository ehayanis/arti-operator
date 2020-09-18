package types

type ArtifactoryInformation struct {
	Tenant              string
	ProjectName         string   // cr
	Stages              []string // cr
	Location            string   //param confmap
	Description         string
	Environment         string
	SourceEntity        string
	ClusterDNSSubdomain string
}

type ArtifactoryRepoUsers struct {
	UserNameRO string
	PasswordRO string
	UserNameRW string
	PasswordRW string
}

type DockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Auth     string `json:"auth,omitempty"`
}

type DockerConfigRegistryInfo struct {
	Url      string
	Username string
	Password string
}

type DockerConfigSecret struct {
	Name       string
	Registries []DockerConfigRegistryInfo
}

type ArtifactoryOperatorConfig struct {
	ClusterLocation               string
	PasswordStoreBackendNamespace string
	PasswordStoreSecretNamePrefix string
	ArtifactoryServerUrl          string
	ArtifactoryServerUser         string
	ArtifactoryServerPassword     string
	VaultServerToken              string
	VaultServerUrl                string
	ClusterDNSSubdomain           string
	LDAPGroups                    LDAPGroups
}

type LDAPGroups struct {
	CustomerOPS string
	Viewer      string
}
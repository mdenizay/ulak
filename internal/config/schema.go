package config

// CurrentSchemaVersion is the latest config schema version.
// Bump this when adding new fields to GlobalConfig or ProjectConfig,
// and add a corresponding migration in internal/config/migration/.
const CurrentSchemaVersion = 1

// GlobalConfig is stored at ~/.ulak/config.json
type GlobalConfig struct {
	SchemaVersion int    `json:"schema_version"`
	DefaultPHP    string `json:"default_php"`    // e.g. "8.3"
	DefaultUser   string `json:"default_user"`   // system user owning deployments
	VaultKey      string `json:"vault_key"`      // base64-encoded, machine-derived
}

// ProjectConfig is stored at ~/.ulak/projects/<name>/config.json
type ProjectConfig struct {
	SchemaVersion int    `json:"schema_version"`
	Name          string `json:"name"`
	Domain        string `json:"domain"`
	RepoURL       string `json:"repo_url"`
	Branch        string `json:"branch"`
	DeployPath    string `json:"deploy_path"`   // absolute path on server, e.g. /var/www/myapp
	PHPVersion    string `json:"php_version"`   // e.g. "8.3"
	HasSSL        bool   `json:"has_ssl"`
	HasNode       bool   `json:"has_node"`      // run npm run build

	// MySQL — values stored encrypted via security.Vault
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"` // encrypted

	// SSH deploy key paths (relative to project dir)
	SSHKeyPath string `json:"ssh_key_path"`
}

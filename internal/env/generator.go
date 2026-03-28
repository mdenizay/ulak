// Package env generates and patches Laravel .env files.
package env

import (
	"fmt"
	"strings"

	"github.com/mdenizay/ulak/internal/config"
	"github.com/mdenizay/ulak/internal/server"
)

// Generate creates a .env file for a Laravel project.
// If an existing .env is present it is backed up to .env.ulak.bak.
func Generate(c *server.Client, cfg *config.ProjectConfig, appKey string) error {
	envPath := fmt.Sprintf("%s/.env", cfg.DeployPath)
	backupCmd := fmt.Sprintf("[ -f %s ] && cp %s %s.ulak.bak || true", envPath, envPath, envPath)
	if _, err := c.Run(backupCmd); err != nil {
		return err
	}

	content := buildEnv(cfg, appKey)
	writeCmd := fmt.Sprintf("cat > %s << 'ULAK_ENV_EOF'\n%s\nULAK_ENV_EOF", envPath, content)
	if out, err := c.Run(writeCmd); err != nil {
		return fmt.Errorf("env: write .env: %w\n%s", err, out)
	}

	// Secure the .env file
	if out, err := c.Run(fmt.Sprintf("chmod 600 %s", envPath)); err != nil {
		return fmt.Errorf("env: chmod: %w\n%s", err, out)
	}
	return nil
}

// Patch updates specific KEY=VALUE pairs in an existing .env file.
func Patch(c *server.Client, deployPath string, values map[string]string) error {
	envPath := fmt.Sprintf("%s/.env", deployPath)
	for k, v := range values {
		// Use sed to replace or append the key
		cmd := fmt.Sprintf(
			`grep -q '^%s=' %s && sed -i 's|^%s=.*|%s=%s|' %s || echo '%s=%s' >> %s`,
			k, envPath, k, k, v, envPath, k, v, envPath,
		)
		if out, err := c.Run(cmd); err != nil {
			return fmt.Errorf("env patch %s: %w\n%s", k, err, out)
		}
	}
	return nil
}

func buildEnv(cfg *config.ProjectConfig, appKey string) string {
	var sb strings.Builder
	sb.WriteString("APP_NAME=" + cfg.Name + "\n")
	sb.WriteString("APP_ENV=production\n")
	sb.WriteString("APP_KEY=" + appKey + "\n")
	sb.WriteString("APP_DEBUG=false\n")
	sb.WriteString("APP_URL=https://" + cfg.Domain + "\n")
	sb.WriteString("\n")
	sb.WriteString("LOG_CHANNEL=stack\n")
	sb.WriteString("LOG_LEVEL=error\n")
	sb.WriteString("\n")
	sb.WriteString("DB_CONNECTION=mysql\n")
	sb.WriteString("DB_HOST=127.0.0.1\n")
	sb.WriteString("DB_PORT=3306\n")
	sb.WriteString("DB_DATABASE=" + cfg.DBName + "\n")
	sb.WriteString("DB_USERNAME=" + cfg.DBUser + "\n")
	sb.WriteString("DB_PASSWORD=" + cfg.DBPassword + "\n")
	sb.WriteString("\n")
	sb.WriteString("CACHE_DRIVER=file\n")
	sb.WriteString("SESSION_DRIVER=file\n")
	sb.WriteString("QUEUE_CONNECTION=sync\n")
	return sb.String()
}

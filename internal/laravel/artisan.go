// Package laravel wraps Laravel artisan commands.
package laravel

import (
	"fmt"
	"strings"

	"github.com/mdenizay/ulak/internal/server"
)

// KeyGenerate runs php artisan key:generate and returns the generated key.
func KeyGenerate(c *server.Client, deployPath, phpVersion string) (string, error) {
	cmd := fmt.Sprintf("cd %s && php%s artisan key:generate --show --force", deployPath, phpVersion)
	out, err := c.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("key:generate: %w\n%s", err, out)
	}
	return strings.TrimSpace(out), nil
}

// Migrate runs php artisan migrate --force.
func Migrate(c *server.Client, deployPath, phpVersion string) error {
	cmd := fmt.Sprintf("cd %s && php%s artisan migrate --force", deployPath, phpVersion)
	if out, err := c.Run(cmd); err != nil {
		return fmt.Errorf("migrate: %w\n%s", err, out)
	}
	return nil
}

// CacheConfig caches the Laravel config.
func CacheConfig(c *server.Client, deployPath, phpVersion string) error {
	cmd := fmt.Sprintf("cd %s && php%s artisan config:cache && php%s artisan route:cache && php%s artisan view:cache",
		deployPath, phpVersion, phpVersion, phpVersion)
	if out, err := c.Run(cmd); err != nil {
		return fmt.Errorf("cache: %w\n%s", err, out)
	}
	return nil
}

// StorageLink runs php artisan storage:link.
func StorageLink(c *server.Client, deployPath, phpVersion string) error {
	cmd := fmt.Sprintf("cd %s && php%s artisan storage:link --force", deployPath, phpVersion)
	if out, err := c.Run(cmd); err != nil {
		return fmt.Errorf("storage:link: %w\n%s", err, out)
	}
	return nil
}

// SetPermissions sets correct ownership and permissions for a Laravel deployment.
func SetPermissions(c *server.Client, deployPath string) error {
	cmds := []string{
		fmt.Sprintf("chown -R www-data:www-data %s", deployPath),
		fmt.Sprintf("find %s -type f -exec chmod 644 {} \\;", deployPath),
		fmt.Sprintf("find %s -type d -exec chmod 755 {} \\;", deployPath),
		fmt.Sprintf("chmod -R 775 %s/storage %s/bootstrap/cache", deployPath, deployPath),
	}
	for _, cmd := range cmds {
		if out, err := c.Run(cmd); err != nil {
			return fmt.Errorf("permissions: %w\n%s", err, out)
		}
	}
	return nil
}

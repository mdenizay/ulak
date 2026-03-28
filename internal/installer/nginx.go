package installer

import (
	"fmt"

	"github.com/mdenizay/ulak/internal/server"
)

// Nginx installs and enables Nginx.
func Nginx(c *server.Client) error {
	cmds := []string{
		"apt-get install -y -qq nginx",
		"systemctl enable nginx",
		"systemctl start nginx",
	}
	for _, cmd := range cmds {
		if out, err := c.Run(cmd); err != nil {
			return fmt.Errorf("install nginx: %w\n%s", err, out)
		}
	}
	return nil
}

// MySQL installs MySQL Server.
func MySQL(c *server.Client) error {
	cmds := []string{
		"DEBIAN_FRONTEND=noninteractive apt-get install -y -qq mysql-server",
		"systemctl enable mysql",
		"systemctl start mysql",
	}
	for _, cmd := range cmds {
		if out, err := c.Run(cmd); err != nil {
			return fmt.Errorf("install mysql: %w\n%s", err, out)
		}
	}
	return nil
}

// Node installs Node.js LTS via NodeSource.
func Node(c *server.Client) error {
	cmds := []string{
		"curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -",
		"apt-get install -y -qq nodejs",
	}
	for _, cmd := range cmds {
		if out, err := c.Run(cmd); err != nil {
			return fmt.Errorf("install nodejs: %w\n%s", err, out)
		}
	}
	return nil
}

// Certbot installs Certbot and the Nginx plugin.
func Certbot(c *server.Client) error {
	cmds := []string{
		"apt-get install -y -qq certbot python3-certbot-nginx",
	}
	for _, cmd := range cmds {
		if out, err := c.Run(cmd); err != nil {
			return fmt.Errorf("install certbot: %w\n%s", err, out)
		}
	}
	return nil
}

// SystemUpdate runs apt-get update && upgrade.
func SystemUpdate(c *server.Client) error {
	cmds := []string{
		"apt-get update -qq",
		"DEBIAN_FRONTEND=noninteractive apt-get upgrade -y -qq",
	}
	for _, cmd := range cmds {
		if out, err := c.Run(cmd); err != nil {
			return fmt.Errorf("system update: %w\n%s", err, out)
		}
	}
	return nil
}

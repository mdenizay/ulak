// Package ssl manages Let's Encrypt certificates via certbot.
package ssl

import (
	"fmt"

	"github.com/mdenizay/ulak/internal/server"
)

// Issue obtains a new certificate for the domain using certbot --nginx.
// email is used for Let's Encrypt account registration.
func Issue(c *server.Client, domain, email string) error {
	cmd := fmt.Sprintf(
		"certbot --nginx --non-interactive --agree-tos --email %s -d %s",
		email, domain,
	)
	if out, err := c.Run(cmd); err != nil {
		return fmt.Errorf("ssl issue: %w\n%s", err, out)
	}
	return nil
}

// Renew runs certbot renew (typically called by a cron job, but exposed for manual use).
func Renew(c *server.Client) error {
	if out, err := c.Run("certbot renew --quiet"); err != nil {
		return fmt.Errorf("ssl renew: %w\n%s", err, out)
	}
	return nil
}

// SetupAutoRenew adds a systemd timer (or cron) to auto-renew certificates.
func SetupAutoRenew(c *server.Client) error {
	// certbot on Ubuntu 20.04+ ships with a systemd timer; just ensure it's enabled.
	cmds := []string{
		"systemctl enable certbot.timer",
		"systemctl start certbot.timer",
	}
	for _, cmd := range cmds {
		if out, err := c.Run(cmd); err != nil {
			return fmt.Errorf("ssl auto-renew: %w\n%s", err, out)
		}
	}
	return nil
}

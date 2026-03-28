// Package installer contains functions to install server dependencies.
// All functions receive an *server.Client and execute commands over SSH.
package installer

import (
	"fmt"

	"github.com/mdenizay/ulak/internal/server"
)

// PHP installs the given PHP version plus common Laravel extensions.
func PHP(c *server.Client, version string) error {
	cmds := []string{
		"apt-get update -qq",
		"apt-get install -y -qq software-properties-common",
		"add-apt-repository -y ppa:ondrej/php",
		"apt-get update -qq",
		fmt.Sprintf(
			"apt-get install -y -qq php%s php%s-fpm php%s-cli php%s-mbstring php%s-xml php%s-bcmath php%s-curl php%s-zip php%s-mysql php%s-tokenizer php%s-intl",
			version, version, version, version, version, version, version, version, version, version, version,
		),
	}
	for _, cmd := range cmds {
		if out, err := c.Run(cmd); err != nil {
			return fmt.Errorf("install php%s: %w\n%s", version, err, out)
		}
	}
	return nil
}

// Composer installs Composer 2 globally.
func Composer(c *server.Client) error {
	script := `
EXPECTED=$(curl -sS https://composer.github.io/installer.sig)
curl -sS https://getcomposer.org/installer -o /tmp/composer-setup.php
ACTUAL=$(php -r "echo hash_file('sha384', '/tmp/composer-setup.php');")
if [ "$EXPECTED" != "$ACTUAL" ]; then echo "Composer checksum mismatch" >&2; exit 1; fi
php /tmp/composer-setup.php --install-dir=/usr/local/bin --filename=composer
rm /tmp/composer-setup.php
`
	if out, err := c.Run(script); err != nil {
		return fmt.Errorf("install composer: %w\n%s", err, out)
	}
	return nil
}

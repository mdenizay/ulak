// Package nginx generates and manages Nginx site configurations.
package nginx

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/mdenizay/ulak/internal/server"
)

// SiteParams holds values for Nginx config template rendering.
type SiteParams struct {
	Domain     string
	DeployPath string
	PHPVersion string // e.g. "8.3"
	HasSSL     bool
}

// Deploy writes an Nginx config for the given params and reloads Nginx.
func Deploy(c *server.Client, p SiteParams) error {
	tmplName := "laravel.conf.tmpl"
	if p.HasSSL {
		tmplName = "laravel-ssl.conf.tmpl"
	}

	tmpl, err := template.ParseFiles("templates/nginx/" + tmplName)
	if err != nil {
		// Fall back to embedded template
		tmpl, err = template.New("nginx").Parse(defaultTemplate(p.HasSSL))
		if err != nil {
			return fmt.Errorf("nginx template: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, p); err != nil {
		return fmt.Errorf("nginx render: %w", err)
	}

	// Write config to a temp file, copy via SSH
	tmp, err := os.CreateTemp("", "ulak-nginx-*.conf")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(buf.String()); err != nil {
		return err
	}
	tmp.Close()

	remotePath := fmt.Sprintf("/etc/nginx/sites-available/%s", p.Domain)
	enablePath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", p.Domain)

	catCmd := fmt.Sprintf("cat > %s << 'ULAK_EOF'\n%s\nULAK_EOF", remotePath, buf.String())
	if out, err := c.Run(catCmd); err != nil {
		return fmt.Errorf("nginx: write config: %w\n%s", err, out)
	}

	cmds := []string{
		fmt.Sprintf("ln -sf %s %s", remotePath, enablePath),
		"nginx -t",
		"systemctl reload nginx",
	}
	for _, cmd := range cmds {
		if out, err := c.Run(cmd); err != nil {
			return fmt.Errorf("nginx: %s: %w\n%s", cmd, err, out)
		}
	}
	return nil
}

func defaultTemplate(ssl bool) string {
	if ssl {
		return `server {
    listen 443 ssl http2;
    server_name {{.Domain}};
    root {{.DeployPath}}/public;
    index index.php;
    server_tokens off;
    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options "nosniff";
    add_header X-XSS-Protection "1; mode=block";

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }
    location ~ \.php$ {
        fastcgi_pass unix:/var/run/php/php{{.PHPVersion}}-fpm.sock;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include fastcgi_params;
        fastcgi_hide_header X-Powered-By;
    }
    location ~ /\.(?!well-known).* {
        deny all;
    }
}
server {
    listen 80;
    server_name {{.Domain}};
    return 301 https://$host$request_uri;
}`
	}
	return `server {
    listen 80;
    server_name {{.Domain}};
    root {{.DeployPath}}/public;
    index index.php;
    server_tokens off;
    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options "nosniff";

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }
    location ~ \.php$ {
        fastcgi_pass unix:/var/run/php/php{{.PHPVersion}}-fpm.sock;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include fastcgi_params;
        fastcgi_hide_header X-Powered-By;
    }
    location ~ /\.(?!well-known).* {
        deny all;
    }
}`
}

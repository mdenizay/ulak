// Package ssl defines the `ulak ssl` subcommands.
package ssl

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mdenizay/ulak/internal/config"
	"github.com/mdenizay/ulak/internal/nginx"
	"github.com/mdenizay/ulak/internal/server"
	"github.com/mdenizay/ulak/internal/ssl"
)

// NewCmd returns the `ulak ssl` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssl",
		Short: "Manage SSL certificates",
	}
	cmd.AddCommand(newIssueCmd())
	cmd.AddCommand(newRenewCmd())
	return cmd
}

func newIssueCmd() *cobra.Command {
	var (
		host    string
		user    string
		keyPath string
		name    string
		email   string
	)

	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue a Let's Encrypt certificate for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadProject(name)
			if err != nil {
				return err
			}

			c, err := server.Connect(host, user, keyPath)
			if err != nil {
				return err
			}
			defer c.Close()

			fmt.Printf("  → %s için SSL sertifikası alınıyor... ", cfg.Domain)
			if err := ssl.Issue(c, cfg.Domain, email); err != nil {
				return err
			}
			fmt.Println("OK")

			// Redeploy Nginx with SSL config
			fmt.Print("  → Nginx SSL konfigürasyonu güncelleniyor... ")
			if err := nginx.Deploy(c, nginx.SiteParams{
				Domain:     cfg.Domain,
				DeployPath: cfg.DeployPath,
				PHPVersion: cfg.PHPVersion,
				HasSSL:     true,
			}); err != nil {
				return err
			}
			fmt.Println("OK")

			// Update project config
			cfg.HasSSL = true
			if err := config.SaveProject(cfg); err != nil {
				return err
			}

			// Ensure auto-renew is set up
			_ = ssl.SetupAutoRenew(c)

			fmt.Printf("\nSSL aktif: https://%s\n", cfg.Domain)
			return nil
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "", "Server IP or hostname (required)")
	cmd.Flags().StringVarP(&user, "user", "u", "root", "SSH user")
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "SSH private key path (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Project name (required)")
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email for Let's Encrypt (required)")
	_ = cmd.MarkFlagRequired("host")
	_ = cmd.MarkFlagRequired("key")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("email")

	return cmd
}

func newRenewCmd() *cobra.Command {
	var (
		host    string
		user    string
		keyPath string
	)

	cmd := &cobra.Command{
		Use:   "renew",
		Short: "Manually renew all certificates on the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := server.Connect(host, user, keyPath)
			if err != nil {
				return err
			}
			defer c.Close()

			fmt.Print("  → Sertifikalar yenileniyor... ")
			if err := ssl.Renew(c); err != nil {
				return err
			}
			fmt.Println("OK")
			return nil
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "", "Server IP or hostname (required)")
	cmd.Flags().StringVarP(&user, "user", "u", "root", "SSH user")
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "SSH private key path (required)")
	_ = cmd.MarkFlagRequired("host")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

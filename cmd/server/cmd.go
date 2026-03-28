// Package server defines the `ulak server` subcommands.
package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mdenizay/ulak/internal/installer"
	"github.com/mdenizay/ulak/internal/server"
)

// NewCmd returns the `ulak server` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Manage the Ubuntu server",
	}
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newUpdateCmd())
	return cmd
}

func newInitCmd() *cobra.Command {
	var (
		host       string
		user       string
		keyPath    string
		phpVersion string
		withNode   bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a fresh Ubuntu server (install Nginx, PHP, MySQL, Composer, Certbot)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := server.Connect(host, user, keyPath)
			if err != nil {
				return err
			}
			defer c.Close()

			steps := []struct {
				name string
				fn   func() error
			}{
				{"system update", func() error { return installer.SystemUpdate(c) }},
				{"nginx", func() error { return installer.Nginx(c) }},
				{"mysql", func() error { return installer.MySQL(c) }},
				{fmt.Sprintf("php%s", phpVersion), func() error { return installer.PHP(c, phpVersion) }},
				{"composer", func() error { return installer.Composer(c) }},
				{"certbot", func() error { return installer.Certbot(c) }},
			}
			if withNode {
				steps = append(steps, struct {
					name string
					fn   func() error
				}{"nodejs", func() error { return installer.Node(c) }})
			}

			for _, s := range steps {
				fmt.Printf("  → %s... ", s.name)
				if err := s.fn(); err != nil {
					fmt.Println("FAILED")
					return err
				}
				fmt.Println("OK")
			}
			fmt.Println("\nSunucu başarıyla hazırlandı.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "", "Server IP or hostname (required)")
	cmd.Flags().StringVarP(&user, "user", "u", "root", "SSH user")
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "SSH private key path (required)")
	cmd.Flags().StringVar(&phpVersion, "php", "8.3", "PHP version to install")
	cmd.Flags().BoolVar(&withNode, "node", false, "Also install Node.js")
	_ = cmd.MarkFlagRequired("host")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

func newUpdateCmd() *cobra.Command {
	var (
		host    string
		user    string
		keyPath string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Run apt-get update && upgrade on the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := server.Connect(host, user, keyPath)
			if err != nil {
				return err
			}
			defer c.Close()
			fmt.Print("Sistem güncelleniyor... ")
			if err := installer.SystemUpdate(c); err != nil {
				return err
			}
			fmt.Println("Tamamlandı.")
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

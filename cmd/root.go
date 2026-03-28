// Package cmd defines all CLI commands for Ulak.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mdenizay/ulak/cmd/project"
	"github.com/mdenizay/ulak/cmd/server"
	"github.com/mdenizay/ulak/cmd/ssl"
	"github.com/mdenizay/ulak/internal/wizard"
)

var rootCmd = &cobra.Command{
	Use:   "ulak",
	Short: "Laravel deployment tool for Ubuntu servers",
	Long: `Ulak — Ubuntu'da Laravel projelerini deploy etmek için güvenlik odaklı CLI aracı.

Komut girmeden çalıştırılırsa interaktif wizard başlar.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return wizard.Run()
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(server.NewCmd())
	rootCmd.AddCommand(project.NewCmd())
	rootCmd.AddCommand(ssl.NewCmd())
	rootCmd.AddCommand(newMigrateCmd())
	rootCmd.AddCommand(newVersionCmd())
}

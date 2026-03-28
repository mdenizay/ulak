package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mdenizay/ulak/internal/config"
	"github.com/mdenizay/ulak/internal/config/migration"
)

func newMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate all local config files to the latest schema version",
		Long: `Runs config schema migrations on ~/.ulak/config.json and all project configs.
Safe to run multiple times. Automatically called on startup when needed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrations()
		},
	}
}

// RunMigrationsIfNeeded is called on startup to silently upgrade configs.
func RunMigrationsIfNeeded() error {
	return runMigrations()
}

func runMigrations() error {
	// Global config
	globalPath := filepath.Join(config.UlakDir(), "config.json")
	if _, err := os.Stat(globalPath); err == nil {
		if err := migration.MigrateFile(globalPath, "global"); err != nil {
			return fmt.Errorf("migrate global config: %w", err)
		}
	}

	// All project configs
	projectsDir := filepath.Join(config.UlakDir(), "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		cfgPath := filepath.Join(projectsDir, entry.Name(), "config.json")
		if _, err := os.Stat(cfgPath); err != nil {
			continue
		}
		if err := migration.MigrateFile(cfgPath, "project"); err != nil {
			return fmt.Errorf("migrate project %q: %w", entry.Name(), err)
		}
		fmt.Printf("  migrated project: %s\n", entry.Name())
	}
	return nil
}

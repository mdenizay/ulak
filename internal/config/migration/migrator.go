// Package migration applies sequential schema migrations to GlobalConfig and ProjectConfig.
// Each migration is a function that transforms a raw map[string]any (parsed JSON).
// This approach avoids coupling migrations to current struct shapes.
//
// To add a migration:
//  1. Increment config.CurrentSchemaVersion in schema.go
//  2. Create v{N}_to_v{N+1}.go with a migrateVN function
//  3. Register it in the globalMigrations / projectMigrations slices below.
package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type migrationFn func(data map[string]any) error

// globalMigrations[i] migrates schema version i+1 → i+2
var globalMigrations = []migrationFn{
	// v1 → v2 will be: migrateGlobalV1ToV2
}

// projectMigrations[i] migrates schema version i+1 → i+2
var projectMigrations = []migrationFn{
	// v1 → v2 will be: migrateProjectV1ToV2
}

// MigrateFile reads a JSON config file, applies pending migrations, and writes it back.
// configType is "global" or "project" (selects migration list).
func MigrateFile(path string, configType string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	version, _ := raw["schema_version"].(float64)
	current := int(version)

	migrations := globalMigrations
	if configType == "project" {
		migrations = projectMigrations
	}

	for i := current; i < len(migrations); i++ {
		if err := migrations[i](raw); err != nil {
			return fmt.Errorf("migration v%d→v%d: %w", i+1, i+2, err)
		}
		raw["schema_version"] = float64(i + 2)
	}

	if int(version) == len(migrations) {
		return nil // nothing to do
	}

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0600)
}

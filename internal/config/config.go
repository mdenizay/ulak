package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// UlakDir returns the path to ~/.ulak
func UlakDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ulak")
}

// ProjectDir returns ~/.ulak/projects/<name>
func ProjectDir(name string) string {
	return filepath.Join(UlakDir(), "projects", name)
}

// LoadGlobal reads and returns the global config, running migrations as needed.
func LoadGlobal() (*GlobalConfig, error) {
	path := filepath.Join(UlakDir(), "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultGlobal(), nil
		}
		return nil, fmt.Errorf("read global config: %w", err)
	}
	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse global config: %w", err)
	}
	return &cfg, nil
}

// SaveGlobal writes the global config to disk.
func SaveGlobal(cfg *GlobalConfig) error {
	cfg.SchemaVersion = CurrentSchemaVersion
	path := filepath.Join(UlakDir(), "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadProject reads and returns a project config, running migrations as needed.
func LoadProject(name string) (*ProjectConfig, error) {
	path := filepath.Join(ProjectDir(name), "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read project config %q: %w", name, err)
	}
	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse project config %q: %w", name, err)
	}
	return &cfg, nil
}

// SaveProject writes a project config to disk.
func SaveProject(cfg *ProjectConfig) error {
	cfg.SchemaVersion = CurrentSchemaVersion
	dir := ProjectDir(cfg.Name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0600)
}

func defaultGlobal() *GlobalConfig {
	return &GlobalConfig{
		SchemaVersion: CurrentSchemaVersion,
		DefaultPHP:    "8.3",
		DefaultUser:   "www-data",
	}
}

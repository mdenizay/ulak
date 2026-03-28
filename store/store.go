// Package store manages the local SQLite database used as a project registry.
// This is separate from the JSON config files — the DB is used for quick lookups
// and audit logs, while JSON files hold the full config.
package store

import (
	"fmt"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	_ "modernc.org/sqlite" // CGO-free SQLite driver; enables cross-compilation without gcc
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/mdenizay/ulak/internal/config"
)

// DB is the global database handle.
var DB *gorm.DB

// Project is the registry entry for a project.
type Project struct {
	ID         uint      `gorm:"primarykey"`
	Name       string    `gorm:"uniqueIndex;not null"`
	Domain     string    `gorm:"not null"`
	RepoURL    string
	DeployPath string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// DeployLog records each deployment attempt.
type DeployLog struct {
	ID        uint      `gorm:"primarykey"`
	ProjectID uint      `gorm:"index;not null"`
	Status    string    // "success" | "failed"
	Duration  int64     // milliseconds
	Error     string
	CreatedAt time.Time
}

// Open initializes the database at ~/.ulak/ulak.db.
func Open() error {
	dbPath := filepath.Join(config.UlakDir(), "ulak.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("store: open db: %w", err)
	}
	if err := db.AutoMigrate(&Project{}, &DeployLog{}); err != nil {
		return fmt.Errorf("store: migrate: %w", err)
	}
	DB = db
	return nil
}

// UpsertProject inserts or updates a project registry entry.
func UpsertProject(name, domain, repoURL, deployPath string) error {
	result := DB.Where(Project{Name: name}).Assign(Project{
		Domain:     domain,
		RepoURL:    repoURL,
		DeployPath: deployPath,
	}).FirstOrCreate(&Project{})
	return result.Error
}

// LogDeploy records a deployment result.
func LogDeploy(projectName, status, errMsg string, durationMs int64) error {
	var p Project
	if err := DB.Where("name = ?", projectName).First(&p).Error; err != nil {
		return err
	}
	return DB.Create(&DeployLog{
		ProjectID: p.ID,
		Status:    status,
		Duration:  durationMs,
		Error:     errMsg,
	}).Error
}

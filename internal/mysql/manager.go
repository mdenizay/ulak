// Package mysql manages remote MySQL databases and users for projects.
package mysql

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Manager handles MySQL operations on a remote server via a root DSN.
type Manager struct {
	db *sql.DB
}

// New connects to MySQL using the given DSN (e.g. "root:pass@tcp(host:3306)/").
func New(dsn string) (*Manager, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql: open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql: ping: %w", err)
	}
	return &Manager{db: db}, nil
}

// CreateDatabase creates a database if it doesn't already exist.
func (m *Manager) CreateDatabase(name string) error {
	_, err := m.db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", name))
	return err
}

// CreateUser creates a MySQL user with a random password and grants privileges on dbName.
// Returns the generated password.
func (m *Manager) CreateUser(username, dbName string) (password string, err error) {
	password, err = randomPassword(24)
	if err != nil {
		return "", err
	}
	queries := []string{
		fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s'", username, password),
		fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, INDEX, ALTER, CREATE TEMPORARY TABLES, LOCK TABLES ON `%s`.* TO '%s'@'localhost'", dbName, username),
		"FLUSH PRIVILEGES",
	}
	for _, q := range queries {
		if _, err := m.db.Exec(q); err != nil {
			return "", fmt.Errorf("mysql: %w", err)
		}
	}
	return password, nil
}

// DropProjectResources removes the database and user for a project.
func (m *Manager) DropProjectResources(dbName, username string) error {
	queries := []string{
		fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName),
		fmt.Sprintf("DROP USER IF EXISTS '%s'@'localhost'", username),
		"FLUSH PRIVILEGES",
	}
	for _, q := range queries {
		if _, err := m.db.Exec(q); err != nil {
			return fmt.Errorf("mysql drop: %w", err)
		}
	}
	return nil
}

func (m *Manager) Close() error {
	return m.db.Close()
}

func randomPassword(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}

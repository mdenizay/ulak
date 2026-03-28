module github.com/mdenizay/ulak

go 1.22

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/charmbracelet/bubbles v0.18.0
	github.com/charmbracelet/bubbletea v0.26.4
	github.com/charmbracelet/lipgloss v0.11.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/google/go-github/v57 v57.0.0
	modernc.org/sqlite v1.29.9
	github.com/spf13/cobra v1.8.1
	golang.org/x/crypto v0.24.0
	golang.org/x/oauth2 v0.21.0
	gorm.io/driver/sqlite v1.5.5
	// modernc.org/sqlite is a CGO-free SQLite driver enabling cross-compilation without gcc toolchain
	gorm.io/gorm v1.25.10
)

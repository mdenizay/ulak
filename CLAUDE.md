# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Ulak** is a security-first Go CLI tool for deploying Laravel projects on Ubuntu servers. It provides an interactive terminal wizard (TUI) for users who don't want to memorize commands, and direct CLI subcommands for power users.

## Geliştirme Ortamı

Ulak **geliştirme makinende** (macOS/Linux) derlenir. Ubuntu sunucuya Go kurmana gerek yok — Ulak, sunucuyu SSH üzerinden uzaktan yönetir.

**Go kurulumu (geliştirme makinesi — macOS):**
```bash
brew install go
```

**Go kurulumu (geliştirme makinesi — Ubuntu/Debian):**
```bash
sudo apt install -y golang-go
# veya daha güncel versiyon için:
sudo snap install go --classic
```

## Build & Run Commands

```bash
# Bağımlılıkları indir
go mod tidy

# Build (geliştirme makinesi için)
go build -o ulak ./main.go

# Build (Ubuntu sunucuya deploy edilecek binary — CGO gerektirir, sqlite için)
GOOS=linux GOARCH=amd64 go build -o ulak-linux-amd64 ./main.go

# Doğrudan çalıştır
go run ./main.go

# Tüm testler
go test ./...

# Tek test
go test ./internal/config/... -run TestMigration_V1ToV2

# Coverage
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

# Lint (golangci-lint gerektirir)
golangci-lint run
```

> **Not:** `store` paketi SQLite kullandığı için `CGO_ENABLED=1` gerekir. Cross-compile için hedef platformda `gcc` toolchain gerekebilir (`apt install gcc-aarch64-linux-gnu` gibi). Alternatif olarak SQLite yerine pure-Go bir driver (`modernc.org/sqlite`) kullanılabilir.

## Architecture

### Directory Structure

```
ulak/
├── main.go                        # Entry point
├── cmd/                           # Cobra CLI commands
│   ├── root.go                    # Root command; launches wizard if no subcommand
│   ├── server/                    # ulak server init|update
│   ├── project/                   # ulak project add|deploy|list|remove|ssh
│   ├── ssl/                       # ulak ssl issue|renew
│   └── migrate.go                 # ulak migrate (config schema migrations)
├── internal/
│   ├── config/                    # Config structs + schema versioning
│   │   └── migration/             # Versioned migration functions (v1→v2, etc.)
│   ├── wizard/                    # Bubble Tea TUI wizards
│   │   └── components/            # Reusable TUI components (inputs, lists, spinners)
│   ├── installer/                 # PHP, Composer, Nginx, MySQL, Node.js, Certbot installers
│   ├── server/                    # SSH connection, key generation, system commands
│   ├── deploy/                    # Deployment orchestrator + rollback
│   ├── git/                       # GitHub API + git clone/pull
│   ├── nginx/                     # Nginx config generation & site management
│   ├── mysql/                     # Remote DB/user creation, secure credential generation
│   ├── ssl/                       # Let's Encrypt via certbot
│   ├── env/                       # .env generation and patching
│   ├── laravel/                   # artisan commands (key:generate, migrate, etc.)
│   └── security/                  # AES-256-GCM vault for secrets, file permission helpers
├── store/                         # Local SQLite state (project registry)
│   └── migrations/                # DB schema migrations (separate from config migrations)
└── templates/                     # Go text/template files
    ├── nginx/                     # laravel.conf.tmpl, laravel-ssl.conf.tmpl
    └── env/                       # laravel.env.tmpl
```

### Key Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI command structure |
| `github.com/charmbracelet/bubbletea` | Interactive TUI wizard |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/charmbracelet/bubbles` | TUI components (textinput, list, spinner, progress) |
| `golang.org/x/crypto/ssh` | SSH connections + RSA/Ed25519 key generation |
| `gorm.io/gorm` + `gorm.io/driver/sqlite` | Local state store (`~/.ulak/ulak.db`) |
| `github.com/google/go-github/v57` | GitHub API (repo info, webhooks) |
| `github.com/blang/semver/v4` | Semantic versioning for config schema migrations |
| `github.com/go-sql-driver/mysql` | Connecting to remote MySQL to create DBs/users |

### Local State

All Ulak state lives in `~/.ulak/`:

```
~/.ulak/
├── config.json                    # Global config (schema_version, defaults)
├── ulak.db                        # SQLite project registry
└── projects/
    └── <project-name>/
        ├── config.json            # Project config (schema_version field required)
        └── ssh/
            ├── id_ed25519         # Per-project deploy key (chmod 600)
            └── id_ed25519.pub
```

### Config Schema Versioning (Critical)

Every config file (global and per-project) has a `schema_version` integer field. On startup, Ulak runs `internal/config/migration/migrator.go` which applies sequential migrations until the config reaches the current version. **Never remove fields from migration functions** — old configs must always be upgradeable. When adding new fields to a config struct, add a migration that sets sensible defaults for existing configs.

Migration files are named `v{N}_to_v{N+1}.go`. The migrator reads `schema_version` and applies all pending migrations in order.

### Deployment Flow

`deploy.Deployer` orchestrates these steps in order, with rollback on any failure:

1. Pull latest code from GitHub (git pull / clone)
2. Run `composer install --no-dev`
3. Generate/update `.env` (MySQL credentials auto-injected)
4. Run `php artisan key:generate --force` (if new install)
5. Run `php artisan migrate --force`
6. Run `npm install && npm run build` (if package.json present)
7. Set file permissions (storage, bootstrap/cache → 775, owned by www-data)
8. Reload Nginx

Each step is logged. On failure, the deployer attempts rollback to the previous git commit and restores the previous `.env`.

### Security Rules

- **Secrets at rest**: MySQL passwords, any stored credentials encrypted via `internal/security/vault.go` (AES-256-GCM). The vault key is derived from a machine-specific secret (`/etc/machine-id` or equivalent) + optional user passphrase stored in the system keyring.
- **SSH keys**: Always created with `0600` permissions. Per-project deploy keys are Ed25519.
- **MySQL**: Each project gets its own MySQL user with `GRANT` only on its own database — never root credentials stored.
- **Nginx**: Configs must include `server_tokens off`, hide PHP version headers.
- **File permissions**: Laravel `storage/` and `bootstrap/cache/` are `775`, owned by `www-data`. No world-writable files.
- **Input sanitization**: Any user-provided value used in shell commands must go through `internal/security/sanitize.go` — never interpolate raw strings into exec.Command.

### TUI Wizard Pattern

When `ulak` is run with no subcommand, `cmd/root.go` launches the Bubble Tea wizard. Each wizard screen is a separate model in `internal/wizard/`. Wizards collect all required inputs then dispatch to the same underlying logic as the CLI commands — **wizard and CLI share the same internal packages**, wizards are only a UI layer.

### Adding a New Command

1. Create file in `cmd/<group>/<action>.go` with a `cobra.Command`
2. Register it in the parent command's `init()` or `NewXxxCmd()` factory
3. Add a corresponding wizard step in `internal/wizard/` if user-facing
4. If it stores new config fields, bump `CurrentSchemaVersion` in `internal/config/schema.go` and add a migration

### PHP Version Management

The installer supports multiple PHP versions via `ondrej/php` PPA. The active PHP version for a project is stored in its config. Nginx and CLI php paths are resolved dynamically from the stored version (e.g., `/usr/bin/php8.3`).

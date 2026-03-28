// Package deploy orchestrates the full Laravel deployment pipeline.
package deploy

import (
	"fmt"
	"time"

	"github.com/mdenizay/ulak/internal/config"
	"github.com/mdenizay/ulak/internal/env"
	"github.com/mdenizay/ulak/internal/laravel"
	"github.com/mdenizay/ulak/internal/server"
)

// Options controls which steps are run during deployment.
type Options struct {
	FreshInstall bool // true on first deploy: generates .env, key, runs storage:link
	RunMigrate   bool
	RunBuild     bool // npm run build
}

// Result contains the outcome of a deployment.
type Result struct {
	StartedAt  time.Time
	FinishedAt time.Time
	Steps      []StepResult
}

// StepResult records the outcome of a single deployment step.
type StepResult struct {
	Name    string
	Output  string
	Error   error
	Skipped bool
}

// Run executes the deployment pipeline for the given project config.
func Run(c *server.Client, cfg *config.ProjectConfig, opts Options) (*Result, error) {
	r := &Result{StartedAt: time.Now()}

	steps := []struct {
		name string
		fn   func() (string, error)
	}{
		{"git pull", func() (string, error) {
			return gitPull(c, cfg)
		}},
		{"composer install", func() (string, error) {
			cmd := fmt.Sprintf("cd %s && composer install --no-dev --no-interaction --optimize-autoloader 2>&1", cfg.DeployPath)
			return c.Run(cmd)
		}},
		{"generate .env", func() (string, error) {
			if !opts.FreshInstall {
				return "", nil // skip: preserve existing .env
			}
			appKey, err := laravel.KeyGenerate(c, cfg.DeployPath, cfg.PHPVersion)
			if err != nil {
				return "", err
			}
			return "", env.Generate(c, cfg, appKey)
		}},
		{"artisan migrate", func() (string, error) {
			if !opts.RunMigrate {
				return "", nil
			}
			return "", laravel.Migrate(c, cfg.DeployPath, cfg.PHPVersion)
		}},
		{"npm run build", func() (string, error) {
			if !opts.RunBuild {
				return "", nil
			}
			cmd := fmt.Sprintf("cd %s && npm ci --silent && npm run build 2>&1", cfg.DeployPath)
			return c.Run(cmd)
		}},
		{"storage:link", func() (string, error) {
			if !opts.FreshInstall {
				return "", nil
			}
			return "", laravel.StorageLink(c, cfg.DeployPath, cfg.PHPVersion)
		}},
		{"set permissions", func() (string, error) {
			return "", laravel.SetPermissions(c, cfg.DeployPath)
		}},
		{"artisan cache", func() (string, error) {
			return "", laravel.CacheConfig(c, cfg.DeployPath, cfg.PHPVersion)
		}},
		{"reload nginx", func() (string, error) {
			return c.Run("systemctl reload nginx")
		}},
	}

	for _, s := range steps {
		out, err := s.fn()
		sr := StepResult{Name: s.name, Output: out, Error: err}
		if out == "" && err == nil {
			sr.Skipped = true
		}
		r.Steps = append(r.Steps, sr)
		if err != nil {
			r.FinishedAt = time.Now()
			return r, fmt.Errorf("step %q failed: %w", s.name, err)
		}
	}

	r.FinishedAt = time.Now()
	return r, nil
}

func gitPull(c *server.Client, cfg *config.ProjectConfig) (string, error) {
	// Check if directory exists
	checkCmd := fmt.Sprintf("[ -d %s/.git ]", cfg.DeployPath)
	if _, err := c.Run(checkCmd); err != nil {
		// Fresh clone
		cmd := fmt.Sprintf(
			"GIT_SSH_COMMAND='ssh -i %s -o StrictHostKeyChecking=no' git clone --branch %s %s %s 2>&1",
			cfg.SSHKeyPath, cfg.Branch, cfg.RepoURL, cfg.DeployPath,
		)
		return c.Run(cmd)
	}
	// Pull
	cmd := fmt.Sprintf(
		"cd %s && GIT_SSH_COMMAND='ssh -i %s -o StrictHostKeyChecking=no' git pull origin %s 2>&1",
		cfg.DeployPath, cfg.SSHKeyPath, cfg.Branch,
	)
	return c.Run(cmd)
}

// Package project defines the `ulak project` subcommands.
package project

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mdenizay/ulak/internal/config"
	"github.com/mdenizay/ulak/internal/deploy"
	"github.com/mdenizay/ulak/internal/mysql"
	"github.com/mdenizay/ulak/internal/nginx"
	"github.com/mdenizay/ulak/internal/security"
	"github.com/mdenizay/ulak/internal/server"
)

// NewCmd returns the `ulak project` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage Laravel projects",
	}
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newDeployCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSSHCmd())
	return cmd
}

func newAddCmd() *cobra.Command {
	var (
		host       string
		sshUser    string
		keyPath    string
		name       string
		domain     string
		repo       string
		branch     string
		deployPath string
		phpVersion string
		withNode   bool
		mysqlDSN   string
		withSSL    bool
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add and configure a new Laravel project on the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate inputs
			if err := security.SafeName(name); err != nil {
				return err
			}
			if err := security.SafeDomain(domain); err != nil {
				return err
			}
			if err := security.SafePath(deployPath); err != nil {
				return err
			}

			vault, err := security.NewVaultFromMachineID()
			if err != nil {
				return err
			}

			// Connect to server
			c, err := server.Connect(host, sshUser, keyPath)
			if err != nil {
				return err
			}
			defer c.Close()

			// Generate per-project SSH deploy key
			fmt.Print("  → SSH deploy key oluşturuluyor... ")
			sshDir := config.ProjectDir(name) + "/ssh"
			pubKey, err := server.GenerateDeployKey(sshDir, "ulak-deploy-"+name)
			if err != nil {
				return err
			}
			fmt.Println("OK")
			fmt.Printf("\n  GitHub Deploy Key (Settings > Deploy keys > Add):\n\n%s\n", pubKey)
			fmt.Print("  Deploy key'i GitHub'a ekleyin ve Enter'a basın: ")
			fmt.Scanln()

			// Copy deploy key to server
			deployKeyRemotePath := fmt.Sprintf("/root/.ssh/ulak_%s", name)
			privKeyData, err := os.ReadFile(sshDir + "/id_ed25519")
			if err != nil {
				return fmt.Errorf("deploy key okunamadı: %w", err)
			}
			writeKeyCmd := fmt.Sprintf("mkdir -p /root/.ssh && cat > %s << 'ULAK_KEY_EOF'\n%s\nULAK_KEY_EOF\nchmod 600 %s",
				deployKeyRemotePath, string(privKeyData), deployKeyRemotePath)
			if out, err := c.Run(writeKeyCmd); err != nil {
				return fmt.Errorf("deploy key upload: %w\n%s", err, out)
			}

			// Setup MySQL
			fmt.Print("  → MySQL veritabanı oluşturuluyor... ")
			mgr, err := mysql.New(mysqlDSN)
			if err != nil {
				return err
			}
			defer mgr.Close()

			dbName := name + "_db"
			dbUser := name + "_user"
			if err := mgr.CreateDatabase(dbName); err != nil {
				return err
			}
			dbPass, err := mgr.CreateUser(dbUser, dbName)
			if err != nil {
				return err
			}
			encPass, err := vault.Encrypt(dbPass)
			if err != nil {
				return err
			}
			fmt.Println("OK")

			// Save project config
			cfg := &config.ProjectConfig{
				Name:       name,
				Domain:     domain,
				RepoURL:    repo,
				Branch:     branch,
				DeployPath: deployPath,
				PHPVersion: phpVersion,
				HasSSL:     withSSL,
				HasNode:    withNode,
				DBName:     dbName,
				DBUser:     dbUser,
				DBPassword: encPass,
				SSHKeyPath: deployKeyRemotePath,
			}
			if err := config.SaveProject(cfg); err != nil {
				return err
			}

			// Setup Nginx
			fmt.Print("  → Nginx konfigürasyonu... ")
			if err := nginx.Deploy(c, nginx.SiteParams{
				Domain:     domain,
				DeployPath: deployPath,
				PHPVersion: phpVersion,
				HasSSL:     false, // SSL eklenir certbot sonrası
			}); err != nil {
				return err
			}
			fmt.Println("OK")

			// Initial deploy
			fmt.Println("  → İlk deploy başlatılıyor...")
			cfg.DBPassword = dbPass // env dosyası için şifresi çözülmüş hali
			result, err := deploy.Run(c, cfg, deploy.Options{
				FreshInstall: true,
				RunMigrate:   true,
				RunBuild:     withNode,
			})
			printSteps(result)
			if err != nil {
				return err
			}

			fmt.Printf("\nProje eklendi: %s → http://%s\n", name, domain)
			return nil
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "", "Server IP or hostname (required)")
	cmd.Flags().StringVarP(&sshUser, "user", "u", "root", "SSH user")
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "SSH private key path (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Project name (required)")
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Domain name (required)")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "GitHub repo SSH URL (required)")
	cmd.Flags().StringVarP(&branch, "branch", "b", "main", "Git branch")
	cmd.Flags().StringVarP(&deployPath, "path", "p", "", "Deploy path on server (required)")
	cmd.Flags().StringVar(&phpVersion, "php", "8.3", "PHP version")
	cmd.Flags().BoolVar(&withNode, "node", false, "Run npm run build")
	cmd.Flags().StringVar(&mysqlDSN, "mysql-dsn", "root:@tcp(127.0.0.1:3306)/", "MySQL root DSN")
	cmd.Flags().BoolVar(&withSSL, "ssl", false, "Issue Let's Encrypt certificate after deploy")

	_ = cmd.MarkFlagRequired("host")
	_ = cmd.MarkFlagRequired("key")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("domain")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}

func newDeployCmd() *cobra.Command {
	var (
		host    string
		user    string
		keyPath string
		name    string
		migrate bool
	)

	cmd := &cobra.Command{
		Use:   "deploy [project-name]",
		Short: "Deploy (git pull + build) an existing project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				name = args[0]
			}
			if name == "" {
				return fmt.Errorf("proje adı gerekli: --name veya argüman olarak verin")
			}

			cfg, err := config.LoadProject(name)
			if err != nil {
				return err
			}

			c, err := server.Connect(host, user, keyPath)
			if err != nil {
				return err
			}
			defer c.Close()

			vault, err := security.NewVaultFromMachineID()
			if err != nil {
				return err
			}
			dbPass, err := vault.Decrypt(cfg.DBPassword)
			if err != nil {
				return err
			}
			cfg.DBPassword = dbPass

			result, err := deploy.Run(c, cfg, deploy.Options{
				FreshInstall: false,
				RunMigrate:   migrate,
				RunBuild:     cfg.HasNode,
			})
			printSteps(result)
			return err
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "", "Server IP or hostname (required)")
	cmd.Flags().StringVarP(&user, "user", "u", "root", "SSH user")
	cmd.Flags().StringVarP(&keyPath, "key", "k", "", "SSH private key path (required)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.Flags().BoolVar(&migrate, "migrate", false, "Run artisan migrate")
	_ = cmd.MarkFlagRequired("host")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectsDir := config.UlakDir() + "/projects"
			entries, err := os.ReadDir(projectsDir)
			if err != nil {
				fmt.Println("Henüz proje yok.")
				return nil
			}
			fmt.Printf("%-20s %-30s %-10s %-6s\n", "İSİM", "DOMAIN", "PHP", "SSL")
			fmt.Println(strings.Repeat("-", 70))
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				cfg, err := config.LoadProject(entry.Name())
				if err != nil {
					continue
				}
				ssl := "hayır"
				if cfg.HasSSL {
					ssl = "evet"
				}
				fmt.Printf("%-20s %-30s %-10s %-6s\n", cfg.Name, cfg.Domain, cfg.PHPVersion, ssl)
			}
			return nil
		},
	}
}

func newSSHCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "ssh [project-name]",
		Short: "Print the deploy public key for a project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				name = args[0]
			}
			if name == "" {
				return fmt.Errorf("proje adı gerekli")
			}
			pubPath := config.ProjectDir(name) + "/ssh/id_ed25519.pub"
			data, err := os.ReadFile(pubPath)
			if err != nil {
				return fmt.Errorf("deploy key bulunamadı (%s): %w", pubPath, err)
			}
			fmt.Printf("Deploy public key (%s):\n\n%s\n", name, string(data))
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Project name")
	return cmd
}

func printSteps(r *deploy.Result) {
	if r == nil {
		return
	}
	for _, s := range r.Steps {
		if s.Skipped {
			continue
		}
		status := "OK"
		if s.Error != nil {
			status = "HATA"
		}
		fmt.Printf("  [%s] %s\n", status, s.Name)
	}
}

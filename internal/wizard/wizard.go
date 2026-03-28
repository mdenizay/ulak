// Package wizard provides an interactive Bubble Tea TUI for Ulak.
// The wizard is a top-level menu that guides the user through common operations
// without needing to remember CLI flags.
package wizard

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Run launches the interactive wizard.
func Run() error {
	m := newMainMenu()
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	if mm, ok := finalModel.(mainMenu); ok && mm.chosen != nil {
		return mm.chosen.run()
	}
	return nil
}

// --- Main menu model ---

type action struct {
	title string
	desc  string
	run   func() error
}

func (a action) Title() string       { return a.title }
func (a action) Description() string { return a.desc }
func (a action) FilterValue() string { return a.title }

type mainMenu struct {
	list   list.Model
	chosen *action
	width  int
	height int
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginLeft(2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginLeft(2)
)

func newMainMenu() mainMenu {
	actions := []list.Item{
		action{
			title: "Sunucu Hazırla",
			desc:  "Yeni Ubuntu sunucusuna Nginx, PHP, MySQL, Composer kur",
			run:   runServerInit,
		},
		action{
			title: "Proje Ekle",
			desc:  "Yeni Laravel projesi ekle ve deploy et",
			run:   runProjectAdd,
		},
		action{
			title: "Proje Deploy Et",
			desc:  "Mevcut projeyi GitHub'dan güncelle",
			run:   runProjectDeploy,
		},
		action{
			title: "SSL Ekle",
			desc:  "Let's Encrypt ile ücretsiz SSL sertifikası al",
			run:   runSSLIssue,
		},
		action{
			title: "Projeleri Listele",
			desc:  "Kayıtlı projeleri görüntüle",
			run:   runProjectList,
		},
		action{
			title: "Çıkış",
			desc:  "Wizard'dan çık",
			run:   func() error { os.Exit(0); return nil },
		},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#7C3AED")).
		BorderLeftForeground(lipgloss.Color("#7C3AED"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#9CA3AF")).
		BorderLeftForeground(lipgloss.Color("#7C3AED"))

	l := list.New(actions, delegate, 60, 20)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle

	return mainMenu{list: l}
}

func (m mainMenu) Init() tea.Cmd {
	return nil
}

func (m mainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-8)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(action); ok {
				m.chosen = &item
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m mainMenu) View() string {
	header := titleStyle.Render("🚀 Ulak — Laravel Deploy Aracı") + "\n" +
		subtitleStyle.Render("↑↓ hareket • enter seç • q çıkış") + "\n\n"
	return header + m.list.View()
}

// --- Action implementations (launch sub-wizards) ---

func runServerInit() error {
	fmt.Println("\nSunucu hazırlama wizard'ı henüz yapım aşamasında.")
	fmt.Println("Kullanım: ulak server init --host <IP> --key <key-path>")
	return nil
}

func runProjectAdd() error {
	fmt.Println("\nProje ekleme wizard'ı henüz yapım aşamasında.")
	fmt.Println("Kullanım: ulak project add --host <IP> --key <key-path> --name <ad> --domain <domain> --repo <ssh-url> --path <path>")
	return nil
}

func runProjectDeploy() error {
	fmt.Println("\nDeploy wizard'ı henüz yapım aşamasında.")
	fmt.Println("Kullanım: ulak project deploy <proje-adı> --host <IP> --key <key-path>")
	return nil
}

func runSSLIssue() error {
	fmt.Println("\nSSL wizard'ı henüz yapım aşamasında.")
	fmt.Println("Kullanım: ulak ssl issue --host <IP> --key <key-path> --name <proje-adı> --email <email>")
	return nil
}

func runProjectList() error {
	fmt.Println("\nProje listesi:")
	fmt.Println("Kullanım: ulak project list")
	return nil
}

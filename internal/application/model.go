package application

import (
	"context"
	"fmt"
	"os"
	"time"

	db "github.com/JonMunkholm/TUI/internal/database"
	"github.com/JonMunkholm/TUI/internal/handler"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgxpool"
)




var (
	keywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8FA3A1")).Background(lipgloss.Color("#8affff"))
	helpStyle	 = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
)

/* ----------------------------------------
	MODEL
---------------------------------------- */

type Model struct {
	currentMenu *Menu
	cursor		int
	loading		bool
	spinner		spinner.Model
	db			*db.Queries
	pool        *pgxpool.Pool
	output		string
}

// InitialModel creates and initializes the application model with database connection.
// Returns an error if initialization fails. Caller must call Close() on the returned
// model when done, even if an error is returned (to clean up partial initialization).
func InitialModel() (*Model, error) {
	model := &Model{
		spinner: spinnerModel(),
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return model, fmt.Errorf("DB_URL environment variable must be set")
	}

	// Parse and configure connection pool
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return model, fmt.Errorf("unable to parse DB_URL: %w", err)
	}

	// Connection pool settings
	config.MaxConns = 10                               // Maximum connections
	config.MinConns = 2                                // Keep minimum connections alive
	config.MaxConnLifetime = 1 * time.Hour             // Recycle connections after 1 hour
	config.MaxConnIdleTime = 5 * time.Minute           // Close idle connections after 5 min
	config.HealthCheckPeriod = 1 * time.Minute         // Check connection health every minute
	config.ConnConfig.ConnectTimeout = 5 * time.Second // Connection timeout

	// Create pool with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return model, fmt.Errorf("unable to create connection pool: %w", err)
	}
	model.pool = pool

	// Verify database is reachable
	if err := pool.Ping(ctx); err != nil {
		return model, fmt.Errorf("unable to connect to database: %w", err)
	}

	model.db = db.New(pool)
	model.currentMenu = buildMenuTree(model)

	return model, nil
}

// Close releases database resources. Call this when the application exits.
func (m *Model) Close() {
	if m.pool != nil {
		m.pool.Close()
	}
}

func spinnerModel() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return s
}

/* ----------------------------------------
	UPDATE
---------------------------------------- */

func (m Model) Init() tea.Cmd {return nil}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	if handleResultMessage(m, msg) {
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case MenuMsg:
		msg.Menu.Parent = m.currentMenu
		m.currentMenu = msg.Menu
		m.cursor = 0
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		if m.output != "" {
			m.output = ""
			m.cursor = 0
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.currentMenu.Items) - 1 {
				m.cursor ++
			}
		case "backspace":
			if m.currentMenu.Parent != nil {
				m.currentMenu = m.currentMenu.Parent
				m.cursor = 0
			}
		case "enter", " ":
			item := m.currentMenu.Items[m.cursor]
			if item.Submenu != nil {
				m.currentMenu = item.Submenu
				m.cursor = 0
			} else if item.Action != nil {
				m.loading = true
				cmds = append(cmds,
				m.spinner.Tick,
			HandleTeaCmdErrorWithTitle(m.currentMenu.Title, m.currentMenu, item.Action()))
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func handleResultMessage(m *Model, msg tea.Msg) bool {
	switch msg := msg.(type) {
	case handler.DoneMsg, handler.WdMsg:
		m.output = fmt.Sprintf("%v", msg)
		m.loading = false
		return true
	case handler.ErrMsg:
		m.output = "Error: " + msg.Err.Error()
		m.loading = false
		return true
	}
	return false
}

/* ----------------------------------------
	VIEW
---------------------------------------- */

func (m Model) View() string {

	// Spinner view
	if m.loading {
		return fmt.Sprintf("\n	Running... %s\n", m.spinner.View())
	}

	// Output view
	if m.output != "" {
		return fmt.Sprintf("\n%s\n\nPress any key to return.\n", m.output)
	}

	// Menu view
	s := fmt.Sprintf("%s\n\n", m.currentMenu.Title)

	for i, item := range m.currentMenu.Items {
		text := item.Label
		if m.cursor == i {
			text = keywordStyle.Render(text)
		}
		s += text + "\n"
	}

	s += "\nBackspace to go back, q to quit.\n"

	return "\n" + helpStyle.Render(s) + "\n"
}


/* ----------------------------------------
	HANDLER MIDDLEWARE
---------------------------------------- */
type MenuMsg struct { Menu *Menu }

func HandleTeaCmdErrorWithTitle(title string, parent *Menu, cmd tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		msg := cmd()

		if errMsg, ok := msg.(handler.ErrMsg); ok {
			return MenuMsg{
				Menu: &Menu{
					Title: title,
					Parent: parent,
					Items: []MenuItem{
						{Label: "Error: " + errMsg.Err.Error()},
						{Label: "Back", Submenu: parent},
					},
				},
			}
		}

		return msg
	}
}

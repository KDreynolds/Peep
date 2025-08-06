package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kylereynolds/peep/internal/storage"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	levelStyles = map[string]lipgloss.Style{
		"error": lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
		"warn":  lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")),
		"info":  lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")),
		"debug": lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
	}
)

// LogItem represents a log entry in the list
type LogItem struct {
	Entry storage.LogEntry
}

func (i LogItem) FilterValue() string {
	return i.Entry.Message + " " + i.Entry.Service + " " + i.Entry.Level
}

func (i LogItem) Title() string {
	levelStyle, exists := levelStyles[strings.ToLower(i.Entry.Level)]
	if !exists {
		levelStyle = lipgloss.NewStyle()
	}

	timestamp := i.Entry.Timestamp.Format("15:04:05")
	level := levelStyle.Render(strings.ToUpper(i.Entry.Level))
	service := fmt.Sprintf("[%s]", i.Entry.Service)

	return fmt.Sprintf("%s %s %s %s", timestamp, level, service, i.Entry.Message)
}

func (i LogItem) Description() string {
	return i.Entry.RawLog
}

// Model represents the TUI application state
type Model struct {
	list         list.Model
	search       textinput.Model
	storage      *storage.Storage
	searchMode   bool
	lastRefresh  time.Time
	refreshTimer *time.Timer
	width        int
	height       int
	err          error
}

// NewModel creates a new TUI model
func NewModel(store *storage.Storage) *Model {
	// Create search input
	search := textinput.New()
	search.Placeholder = "Search logs..."
	search.Focus()

	// Create list
	items := []list.Item{}
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("#F25D94"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("#AD58B4"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "🔍 Peep - Live Logs"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	m := &Model{
		list:        l,
		search:      search,
		storage:     store,
		lastRefresh: time.Now(),
	}

	// Load initial logs
	m.refreshLogs()

	return m
}

// refreshLogs loads the latest logs from storage
func (m *Model) refreshLogs() {
	logs, err := m.storage.GetLogs(100) // Get last 100 logs
	if err != nil {
		m.err = err
		return
	}

	items := make([]list.Item, len(logs))
	for i, log := range logs {
		items[i] = LogItem{Entry: log}
	}

	m.list.SetItems(items)
	m.lastRefresh = time.Now()
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.tickRefresh(),
	)
}

// tickRefresh returns a command that refreshes logs every 2 seconds
func (m *Model) tickRefresh() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return refreshMsg{}
	})
}

type refreshMsg struct{}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // Leave space for search and help

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "/":
			// Toggle search mode
			m.searchMode = !m.searchMode
			if m.searchMode {
				m.search.Focus()
				return m, textinput.Blink
			} else {
				m.search.Blur()
				m.list.SetFilteringEnabled(true)
			}

		case "r":
			// Manual refresh
			if !m.searchMode {
				m.refreshLogs()
				return m, m.tickRefresh()
			}

		case "esc":
			if m.searchMode {
				m.searchMode = false
				m.search.Blur()
				m.search.SetValue("")
				m.list.SetFilteringEnabled(true)
			}

		case "enter":
			if m.searchMode {
				// Apply search filter
				searchTerm := m.search.Value()
				if searchTerm != "" {
					m.list.SetFilteringEnabled(false)
					// Filter items based on search term
					allItems := m.list.Items()
					var filteredItems []list.Item
					for _, item := range allItems {
						if logItem, ok := item.(LogItem); ok {
							if strings.Contains(strings.ToLower(logItem.FilterValue()), strings.ToLower(searchTerm)) {
								filteredItems = append(filteredItems, item)
							}
						}
					}
					m.list.SetItems(filteredItems)
				}
				m.searchMode = false
				m.search.Blur()
			}
		}

	case refreshMsg:
		// Auto-refresh logs
		m.refreshLogs()
		return m, m.tickRefresh()
	}

	// Update components based on mode
	if m.searchMode {
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m *Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var content strings.Builder

	// Main list view
	content.WriteString(m.list.View())
	content.WriteString("\n")

	// Search bar (if in search mode)
	if m.searchMode {
		content.WriteString("Search: " + m.search.View())
	} else {
		// Status bar
		status := fmt.Sprintf("Last refresh: %s | %d logs",
			m.lastRefresh.Format("15:04:05"),
			len(m.list.Items()))
		content.WriteString(statusStyle.Render(status))
	}
	content.WriteString("\n")

	// Help text
	help := "Press 'q' to quit, '/' to search, 'r' to refresh, 'esc' to cancel search"
	content.WriteString(helpStyle.Render(help))

	return content.String()
}

// Start runs the TUI application
func Start(store *storage.Storage) error {
	model := NewModel(store)

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

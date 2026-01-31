package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/config"
	"troveler/db"
	"troveler/internal/search"
	"troveler/internal/update"
	"troveler/tui/panels"
)

// PanelID identifies which panel is active
type PanelID int

const (
	PanelSearch PanelID = iota
	PanelTools
	PanelInfo
	PanelInstall
)

// Model is the main TUI model
type Model struct {
	// Core dependencies
	db            *db.SQLiteDB
	config        *config.Config
	searchService *search.Service

	// Terminal size
	width  int
	height int

	// Panel management
	activePanel  PanelID
	searchPanel  *panels.SearchPanel
	toolsPanel   *panels.ToolsPanel
	infoPanel    *panels.InfoPanel
	installPanel *panels.InstallPanel

	// Keybindings
	keys KeyMap

	// Data state
	tools        []db.SearchResult
	selectedTool *db.Tool
	installs     []db.InstallInstruction
	searching    bool

	// Modal states
	showHelp             bool
	showInfoModal        bool
	showUpdateModal      bool
	showInstallModal     bool
	showBatchConfigModal bool

	// Install execution state
	executing     bool
	executeOutput string

	// Batch install state
	batchConfig     *BatchInstallConfig
	batchProgress   *BatchInstallProgress

	// Update state
	updating       bool
	updateService  *update.Service
	updateSlugWave *update.SlugWave
	updateProgress chan update.ProgressUpdate
	updateCancel   context.CancelFunc

	// Error state
	err error
}

// Panel interface for all panel types
type Panel interface {
	Update(msg tea.Msg) (Panel, tea.Cmd)
	View(width, height int) string
	Focus()
	Blur()
	IsFocused() bool
}

// NewModel creates a new TUI model
func NewModel(database *db.SQLiteDB, cfg *config.Config) *Model {
	searchPanel := panels.NewSearchPanel()
	searchPanel.Focus() // Start with search focused

	toolsPanel := panels.NewToolsPanel()
	infoPanel := panels.NewInfoPanel()
	installPanel := panels.NewInstallPanel(
		"", // CLI override (will be set from command line later)
		cfg.Install.PlatformOverride,
		cfg.Install.FallbackPlatform,
		cfg.Install.MiseMode,
	)

	m := &Model{
		db:            database,
		config:        cfg,
		searchService: search.NewService(database),
		keys:          DefaultKeyMap(),
		activePanel:   PanelSearch,
		searchPanel:   searchPanel,
		toolsPanel:    toolsPanel,
		infoPanel:     infoPanel,
		installPanel:  installPanel,
		tools:         []db.SearchResult{},
	}

	return m
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	// Load initial tools (empty query = all tools)
	return m.performSearch("")
}

// SetSize sets the terminal size
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// performSearch executes a search query
func (m *Model) performSearch(query string) tea.Cmd {
	return func() tea.Msg {
		opts := search.Options{
			Query:     query,
			Limit:     1000,
			SortField: "name",
			SortOrder: "ASC",
		}

		result, err := m.searchService.Search(context.Background(), opts)
		if err != nil {
			return searchErrorMsg{err: err}
		}

		return searchResultMsg{
			tools: result.Tools,
			query: query,
		}
	}
}

// searchResultMsg contains search results
type searchResultMsg struct {
	tools []db.SearchResult
	query string
}

// searchErrorMsg contains search errors
type searchErrorMsg struct {
	err error
}

// NextPanel cycles to the next panel
func (m *Model) NextPanel() {
	// Blur current panel
	switch m.activePanel {
	case PanelSearch:
		m.searchPanel.Blur()
		m.activePanel = PanelTools
		m.toolsPanel.Focus()
	case PanelTools:
		m.toolsPanel.Blur()
		m.activePanel = PanelInstall
		m.installPanel.Focus()
	case PanelInstall:
		m.installPanel.Blur()
		m.activePanel = PanelSearch
		m.searchPanel.Focus()
	}
}

// SetError sets an error to display
func (m *Model) SetError(err error) {
	m.err = err
}

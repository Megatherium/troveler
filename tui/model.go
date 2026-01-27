package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"troveler/config"
	"troveler/db"
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
	db     *db.SQLiteDB
	config *config.Config

	// Terminal size
	width  int
	height int

	// Panel management
	activePanel PanelID
	panels      map[PanelID]Panel

	// Keybindings
	keys KeyMap

	// Data state
	tools        []db.SearchResult
	selectedTool *db.Tool
	installs     []db.InstallInstruction

	// Modal states
	showHelp       bool
	showInfoModal  bool
	showUpdateModal bool

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
	m := &Model{
		db:          database,
		config:      cfg,
		keys:        DefaultKeyMap(),
		activePanel: PanelSearch,
		panels:      make(map[PanelID]Panel),
		tools:       []db.SearchResult{},
	}

	// Initialize panels (will be implemented in next phase)
	// m.panels[PanelSearch] = NewSearchPanel()
	// m.panels[PanelTools] = NewToolsPanel()
	// m.panels[PanelInfo] = NewInfoPanel()
	// m.panels[PanelInstall] = NewInstallPanel()

	return m
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// SetSize sets the terminal size
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetActivePanel returns the currently active panel
func (m *Model) GetActivePanel() Panel {
	return m.panels[m.activePanel]
}

// NextPanel cycles to the next panel
func (m *Model) NextPanel() {
	// Blur current panel
	if p := m.panels[m.activePanel]; p != nil {
		p.Blur()
	}

	// Cycle through panels: Search -> Tools -> Install (Info is passive)
	switch m.activePanel {
	case PanelSearch:
		m.activePanel = PanelTools
	case PanelTools:
		m.activePanel = PanelInstall
	case PanelInstall:
		m.activePanel = PanelSearch
	}

	// Focus new panel
	if p := m.panels[m.activePanel]; p != nil {
		p.Focus()
	}
}

// SetError sets an error to display
func (m *Model) SetError(err error) {
	m.err = err
}

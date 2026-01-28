package tui

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/install"
	"troveler/internal/update"
	"troveler/tui/panels"
)

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case panels.SearchTriggeredMsg:
		// Search was triggered (debounced or Enter)
		m.searching = true
		return m, m.performSearch(msg.Query)

	case searchResultMsg:
		// Search results received
		m.tools = msg.tools
		m.toolsPanel.SetTools(msg.tools)
		m.searching = false

		// Auto-select first tool to populate info/install panels
		if len(msg.tools) > 0 {
			firstTool := &msg.tools[0].Tool
			m.selectedTool = firstTool
			m.infoPanel.SetTool(firstTool, []db.InstallInstruction{})

			// Load install instructions for first tool
			installs, err := m.db.GetInstallInstructions(firstTool.ID)
			if err == nil {
				m.installs = installs
				m.infoPanel.SetTool(firstTool, installs)
				m.installPanel.SetTool(firstTool, installs)
			}
		}
		return m, nil

	case searchErrorMsg:
		// Search error
		m.err = msg.err
		m.searching = false
		return m, nil

	case panels.ToolCursorChangedMsg:
		// Cursor moved to a different tool - update info/install panels
		m.selectedTool = &msg.Tool.Tool

		// Update info panel
		m.infoPanel.SetTool(m.selectedTool, []db.InstallInstruction{})

		// Load install instructions
		installs, err := m.db.GetInstallInstructions(m.selectedTool.ID)
		if err == nil {
			m.installs = installs
			m.infoPanel.SetTool(m.selectedTool, installs)
			m.installPanel.SetTool(m.selectedTool, installs)
		}
		return m, nil

	case panels.ToolSelectedMsg:
		// Tool was selected (Enter pressed in tools panel)
		// Info/install panels already populated by cursor change, just jump to install panel
		m.toolsPanel.Blur()
		m.activePanel = PanelInstall
		m.installPanel.Focus()
		return m, nil

	case panels.InstallExecuteMsg:
		// User wants to execute install command
		m.showInstallModal = true
		m.executing = true
		m.executeOutput = ""
		return m, m.executeInstallCommand(msg.Command)

	case panels.InstallExecuteMiseMsg:
		// User wants to execute install command via mise
		m.showInstallModal = true
		m.executing = true
		m.executeOutput = ""
		return m, m.executeInstallCommand(msg.Command)

	case installCompleteMsg:
		// Install finished
		m.executing = false
		m.executeOutput = msg.output
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil

	case updateProgressMsg:
		// Update progress received
		upd := update.ProgressUpdate(msg)

		// Initialize slug wave when we know total count
		if upd.Type == "progress" && upd.Total > 0 && m.updateSlugWave == nil {
			m.updateSlugWave = update.NewSlugWave(upd.Total)
		}

		if upd.Type == "slug" && m.updateSlugWave != nil {
			m.updateSlugWave.AddSlug(upd.Slug)
			m.updateSlugWave.IncProcessed()
		}

		if upd.Type == "complete" {
			m.updating = false
			return m, nil
		}

		if upd.Type == "error" {
			m.updating = false
			m.err = upd.Error
			return m, nil
		}

		// Continue listening for more updates
		return m, m.listenForUpdates()

	case slugTickMsg:
		// Slug wave animation frame - keep ticking while updating
		if m.updating {
			if m.updateSlugWave != nil {
				m.updateSlugWave.AdvanceFrame()
			}
			return m, m.tickSlugWave()
		}
		return m, nil

	case tea.MouseMsg:
		// Mouse support disabled per spec
		return m, nil
	}

	// Forward to active panel
	switch m.activePanel {
	case PanelSearch:
		cmd, updatedPanel := m.searchPanel.Update(msg)
		m.searchPanel = updatedPanel
		cmds = append(cmds, cmd)
	case PanelTools:
		cmd := m.toolsPanel.Update(msg)
		cmds = append(cmds, cmd)
	case PanelInstall:
		cmd := m.installPanel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyPress processes keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check critical global keys first (help, quit, etc.)
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		return m, nil
	case key.Matches(msg, m.keys.Tab):
		m.NextPanel()
		return m, nil
	}

	// CRITICAL: Forward regular keys to search panel FIRST when it's focused
	// This allows typing characters like 'i', 'u', etc. in the search box
	if m.activePanel == PanelSearch && !msg.Alt && msg.Type == tea.KeyRunes {
		cmd, updatedPanel := m.searchPanel.Update(msg)
		m.searchPanel = updatedPanel
		return m, cmd
	}

	// Other global keybindings
	switch {
	case key.Matches(msg, m.keys.Escape):
		// Close modals if open
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		if m.showInfoModal {
			m.showInfoModal = false
			return m, nil
		}
		if m.showUpdateModal {
			m.showUpdateModal = false
			m.updating = false
			// Cancel update context to stop goroutines
			if m.updateCancel != nil {
				m.updateCancel()
				m.updateCancel = nil
			}
			// Don't close channel - let service close it
			m.updateProgress = nil
			return m, nil
		}
		if m.showInstallModal {
			// Only close if not currently executing
			if !m.executing {
				m.showInstallModal = false
				m.executeOutput = ""
				m.err = nil
			}
			return m, nil
		}
		if m.executeOutput != "" {
			m.executeOutput = ""
			m.err = nil
			return m, nil
		}
		// Also forward ESC to active panel
		if m.activePanel == PanelSearch {
			cmd, updatedPanel := m.searchPanel.Update(msg)
			m.searchPanel = updatedPanel
			return m, cmd
		}
		return m, nil

	case key.Matches(msg, m.keys.Update):
		m.showUpdateModal = true
		m.updating = true
		m.updateService = update.NewService(m.db)
		m.updateProgress = make(chan update.ProgressUpdate, 100)
		return m, tea.Batch(
			m.startUpdate(),
			m.tickSlugWave(),
		)

	case key.Matches(msg, m.keys.InfoModal):
		// Only show modal if NOT typing in search
		if m.selectedTool != nil && m.activePanel != PanelSearch {
			m.showInfoModal = true
		}
		return m, nil

	case msg.Alt && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'i':
		// Execute install from any panel (Alt+i)
		if m.installPanel.HasCommands() {
			cmd := m.installPanel.GetSelectedCommand()
			if cmd != "" {
				return m, func() tea.Msg {
					return panels.InstallExecuteMsg{Command: cmd}
				}
			}
		}
		return m, nil

	case msg.Alt && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'm':
		// Execute install via mise from any panel (Alt+m) - forces mise transformation
		if m.installPanel.HasCommands() {
			cmd := m.installPanel.GetSelectedCommand()
			if cmd != "" {
				// Transform to mise
				transformedCmd := install.TransformToMise(cmd)
				return m, func() tea.Msg {
					return panels.InstallExecuteMiseMsg{Command: transformedCmd}
				}
			}
		}
		return m, nil

	case msg.Alt && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'r':
		// Open repository URL in browser (Alt+r)
		return m, m.openRepositoryURL()
	}

	// Forward to search panel for non-rune keys (Enter, Backspace, etc.)
	if m.activePanel == PanelSearch {
		cmd, updatedPanel := m.searchPanel.Update(msg)
		m.searchPanel = updatedPanel
		return m, cmd
	}

	// Forward to other panels
	switch m.activePanel {
	case PanelTools:
		cmd := m.toolsPanel.Update(msg)
		return m, cmd
	case PanelInstall:
		cmd := m.installPanel.Update(msg)
		return m, cmd
	}

	return m, nil
}

// installCompleteMsg is sent when install completes
type installCompleteMsg struct {
	output string
	err    error
}

// executeInstallCommand runs the install command
func (m *Model) executeInstallCommand(command string) tea.Cmd {
	return func() tea.Msg {
		// Execute command using shell
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()

		return installCompleteMsg{
			output: string(output),
			err:    err,
		}
	}
}

// openRepositoryURL opens the repository URL in the default browser
func (m *Model) openRepositoryURL() tea.Cmd {
	return func() tea.Msg {
		if m.selectedTool == nil || m.selectedTool.CodeRepository == "" {
			return nil
		}

		var cmd *exec.Cmd
		url := m.selectedTool.CodeRepository

		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default: // Linux and others
			cmd = exec.Command("xdg-open", url)
		}

		// Start command but don't wait - runs in background
		_ = cmd.Start()
		return nil
	}
}

// updateProgressMsg wraps update progress events
type updateProgressMsg update.ProgressUpdate

// slugTickMsg triggers slug wave animation frame
type slugTickMsg struct{}

// startUpdate begins the database update
func (m *Model) startUpdate() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.updateCancel = cancel

	opts := update.UpdateOptions{
		Limit:    0, // Fetch all tools
		Progress: m.updateProgress,
	}

	// Run update in background
	go func() {
		m.updateService.FetchAndUpdate(ctx, opts)
	}()

	// Return a command that listens for progress updates
	return m.listenForUpdates()
}

// listenForUpdates returns a command that blocks waiting for channel updates
func (m *Model) listenForUpdates() tea.Cmd {
	return func() tea.Msg {
		if m.updateProgress == nil {
			return nil
		}
		upd, ok := <-m.updateProgress
		if !ok {
			return updateProgressMsg{Type: "complete"}
		}
		return updateProgressMsg(upd)
	}
}

// tickSlugWave returns a command that triggers the next animation frame at ~30fps
func (m *Model) tickSlugWave() tea.Cmd {
	return tea.Tick(time.Millisecond*33, func(t time.Time) tea.Msg {
		return slugTickMsg{}
	})
}

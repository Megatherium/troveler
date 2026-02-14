package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/install"
	"troveler/internal/update"
	"troveler/lib"
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
		return m.handleSearchTriggered(msg)

	case searchResultMsg:
		return m.handleSearchResult(msg)

	case searchErrorMsg:
		return m.handleSearchError(msg)

	case panels.ToolMarkedMsg:
		// Tool mark status changed - just update display
		return m, nil

	case panels.ToolCursorChangedMsg:
		return m.handleToolCursorChanged(msg)

	case panels.ToolSelectedMsg:
		return m.handleToolSelected()

	case panels.InstallExecuteMsg:
		return m.handleInstallExecute(msg)

	case panels.InstallExecuteMiseMsg:
		return m.handleInstallExecuteMise(msg)

	case installCompleteMsg:
		return m.handleInstallComplete(msg)

	case batchInstallStartMsg:
		return m, m.processBatchTool(0)

	case batchInstallProgressMsg:
		return m.handleBatchInstallProgress(msg)

	case batchInstallCompleteMsg:
		return m.handleBatchInstallComplete()

	case updateProgressMsg:
		return m.handleUpdateProgress(msg)

	case slugTickMsg:
		return m.handleSlugTick()

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

// Message handlers extracted to reduce cyclomatic complexity

func (m *Model) handleSearchTriggered(msg panels.SearchTriggeredMsg) (tea.Model, tea.Cmd) {
	m.searching = true

	return m, m.performSearch(msg.Query)
}

func (m *Model) handleSearchResult(msg searchResultMsg) (tea.Model, tea.Cmd) {
	m.tools = msg.tools
	m.toolsPanel.SetTools(msg.tools)
	m.searching = false

	// Update installed status for all tools
	m.toolsPanel.UpdateAllInstalledStatus(m.db.GetInstallInstructions)

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
}

func (m *Model) handleSearchError(msg searchErrorMsg) (tea.Model, tea.Cmd) {
	m.err = msg.err
	m.searching = false

	return m, nil
}

func (m *Model) handleToolCursorChanged(msg panels.ToolCursorChangedMsg) (tea.Model, tea.Cmd) {
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
}

func (m *Model) handleToolSelected() (tea.Model, tea.Cmd) {
	// Info/install panels already populated by cursor change, just jump to install panel
	m.toolsPanel.Blur()
	m.activePanel = PanelInstall
	m.installPanel.Focus()

	return m, nil
}

func (m *Model) handleInstallExecute(msg panels.InstallExecuteMsg) (tea.Model, tea.Cmd) {
	m.showInstallModal = true
	m.executing = true
	m.executeOutput = ""

	return m, m.executeInstallCommand(msg.Command)
}

func (m *Model) handleInstallExecuteMise(msg panels.InstallExecuteMiseMsg) (tea.Model, tea.Cmd) {
	m.showInstallModal = true
	m.executing = true
	m.executeOutput = ""

	return m, m.executeInstallCommand(msg.Command)
}

func (m *Model) handleInstallComplete(msg installCompleteMsg) (tea.Model, tea.Cmd) {
	m.executing = false
	m.executeOutput = msg.output
	if msg.err != nil {
		m.err = msg.err
	}

	return m, nil
}

func (m *Model) handleBatchInstallProgress(msg batchInstallProgressMsg) (tea.Model, tea.Cmd) {
	if m.batchProgress == nil {
		return m, nil
	}

	if msg.skipped {
		m.batchProgress.Skipped = append(m.batchProgress.Skipped, msg.toolID)
	} else if msg.err != nil {
		m.batchProgress.Failed = append(m.batchProgress.Failed, msg.toolID)
	} else {
		m.batchProgress.Completed = append(m.batchProgress.Completed, msg.toolID)
	}
	m.batchProgress.CurrentOutput = msg.output
	m.batchProgress.CurrentError = msg.err
	m.batchProgress.CurrentIndex++

	// Process next tool or complete
	if m.batchProgress.CurrentIndex < len(m.batchProgress.Tools) {
		return m, m.processBatchTool(m.batchProgress.CurrentIndex)
	}
	// All done
	m.batchProgress.IsComplete = true
	m.executing = false
	m.toolsPanel.ClearMarks()

	return m, nil
}

func (m *Model) handleBatchInstallComplete() (tea.Model, tea.Cmd) {
	m.executing = false
	if m.batchProgress != nil {
		m.batchProgress.IsComplete = true
	}

	return m, nil
}

func (m *Model) handleUpdateProgress(msg updateProgressMsg) (tea.Model, tea.Cmd) {
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
}

func (m *Model) handleSlugTick() (tea.Model, tea.Cmd) {
	// Slug wave animation frame - keep ticking while updating
	if m.updating && m.updateSlugWave != nil {
		m.updateSlugWave.AdvanceFrame()

		return m, m.tickSlugWave()
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check critical global keys first (help, quit, etc.)
	if result, cmd, handled := m.handleGlobalKeys(msg); handled {
		return result, cmd
	}

	// Handle batch config modal input
	if m.showBatchConfigModal && m.batchConfig != nil && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		r := msg.Runes[0]
		if r == '1' || r == '2' {
			optionIndex := int(r - '1')
			m.batchConfig.SetCurrentStepValue(optionIndex)
			if !m.batchConfig.NextStep() {
				// Config complete, start batch install
				m.showBatchConfigModal = false

				return m, m.startBatchInstall()
			}
		}

		return m, nil
	}

	// CRITICAL: Forward regular keys to search panel FIRST when it's focused
	// This allows typing characters like 'i', 'u', etc. in the search box
	if m.activePanel == PanelSearch && !msg.Alt && msg.Type == tea.KeyRunes {
		cmd, updatedPanel := m.searchPanel.Update(msg)
		m.searchPanel = updatedPanel

		return m, cmd
	}

	// Handle Escape key - close modals
	if key.Matches(msg, m.keys.Escape) {
		return m.handleEscapeKey()
	}

	// Handle action keys (update, info modal, install shortcuts)
	if result, cmd, handled := m.handleActionKeys(msg); handled {
		return result, cmd
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

// handleGlobalKeys processes global keys (quit, help, tab). Returns (model, cmd, handled).
func (m *Model) handleGlobalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit, true
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp

		return m, nil, true
	case key.Matches(msg, m.keys.Tab):
		m.NextPanel()

		return m, nil, true
	}

	return m, nil, false
}

// handleEscapeKey handles the escape key to close modals.
func (m *Model) handleEscapeKey() (tea.Model, tea.Cmd) {
	// Close modals in order of priority
	if m.showHelp {
		m.showHelp = false

		return m, nil
	}
	if m.showInfoModal {
		m.showInfoModal = false

		return m, nil
	}
	if m.showUpdateModal {
		return m.closeUpdateModal()
	}
	if m.showInstallModal {
		return m.closeInstallModal()
	}
	if m.showBatchConfigModal {
		m.showBatchConfigModal = false
		m.batchConfig = nil

		return m, nil
	}
	if m.executeOutput != "" {
		m.executeOutput = ""
		m.err = nil

		return m, nil
	}

	// Forward ESC to search panel if it's active
	if m.activePanel == PanelSearch {
		cmd, updatedPanel := m.searchPanel.Update(m.keys.Escape)
		m.searchPanel = updatedPanel

		return m, cmd
	}

	return m, nil
}

// closeUpdateModal closes the update modal and cancels any ongoing update.
func (m *Model) closeUpdateModal() (tea.Model, tea.Cmd) {
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

// closeInstallModal closes the install modal if not currently executing.
func (m *Model) closeInstallModal() (tea.Model, tea.Cmd) {
	// Only close if not currently executing
	if !m.executing {
		m.showInstallModal = false
		m.executeOutput = ""
		m.err = nil
		m.batchProgress = nil
		m.batchConfig = nil
	}

	return m, nil
}

// handleActionKeys processes action keys (update, info, install shortcuts). Returns (model, cmd, handled).
func (m *Model) handleActionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Update):
		return m.handleUpdateKey()
	case key.Matches(msg, m.keys.InfoModal):
		return m.handleInfoModalKey()
	case msg.Alt && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'i':
		return m.handleAltIKey()
	case msg.Alt && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'm':
		return m.handleAltMKey()
	case msg.Alt && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'r':
		return m, m.openRepositoryURL(), true
	}

	return m, nil, false
}

// handleUpdateKey handles the update key press.
func (m *Model) handleUpdateKey() (tea.Model, tea.Cmd, bool) {
	m.showUpdateModal = true
	m.updating = true
	m.updateService = update.NewService(m.db)
	m.updateProgress = make(chan update.ProgressUpdate, 100)

	return m, tea.Batch(
		m.startUpdate(),
		m.tickSlugWave(),
	), true
}

// handleInfoModalKey handles the info modal key press.
func (m *Model) handleInfoModalKey() (tea.Model, tea.Cmd, bool) {
	// Only show modal if NOT typing in search
	if m.selectedTool != nil && m.activePanel != PanelSearch {
		m.showInfoModal = true
	}

	return m, nil, true
}

// handleAltIKey handles Alt+i key press for install.
func (m *Model) handleAltIKey() (tea.Model, tea.Cmd, bool) {
	// If tools are marked, show batch config modal
	if m.toolsPanel.GetMarkedCount() > 0 {
		m.batchConfig = NewBatchInstallConfig()
		m.showBatchConfigModal = true

		return m, nil, true
	}
	// Otherwise, single tool install
	if m.installPanel.HasCommands() {
		cmd := m.installPanel.GetSelectedCommand()
		if cmd != "" {
			return m, func() tea.Msg {
				return panels.InstallExecuteMsg{Command: cmd}
			}, true
		}
	}

	return m, nil, true
}

// handleAltMKey handles Alt+m key press for mise install.
func (m *Model) handleAltMKey() (tea.Model, tea.Cmd, bool) {
	// If tools are marked, show batch config modal with mise enabled
	if m.toolsPanel.GetMarkedCount() > 0 {
		m.batchConfig = NewBatchInstallConfig()
		m.batchConfig.UseMise = true
		m.showBatchConfigModal = true

		return m, nil, true
	}
	// Otherwise, single tool install
	if m.installPanel.HasCommands() {
		cmd := m.installPanel.GetSelectedCommand()
		if cmd != "" {
			transformedCmd := install.TransformToMise(cmd)

			return m, func() tea.Msg {
				return panels.InstallExecuteMiseMsg{Command: transformedCmd}
			}, true
		}
	}

	return m, nil, true
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
		cmd := exec.Command("sh", "-c", command) //nolint:noctx //nolint:gosec // G204
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
			cmd = exec.Command("open", url) //nolint:gosec // G204: opening URL in browser
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url) //nolint:gosec // G204
		default: // Linux and others
			cmd = exec.Command("xdg-open", url) //nolint:gosec // G204: opening URL in browser
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
		_ = m.updateService.FetchAndUpdate(ctx, opts)
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

// batchInstallStartMsg signals batch install is starting
type batchInstallStartMsg struct {
	tools []db.SearchResult
}

// batchInstallProgressMsg contains progress for a single tool
type batchInstallProgressMsg struct {
	toolID  string
	output  string
	err     error
	skipped bool
}

// batchInstallCompleteMsg signals all tools are processed
type batchInstallCompleteMsg struct{}

// startBatchInstall begins the batch installation process
func (m *Model) startBatchInstall() tea.Cmd {
	markedTools := m.toolsPanel.GetMarkedTools()
	if len(markedTools) == 0 {
		return nil
	}

	m.batchProgress = NewBatchInstallProgress(markedTools)
	m.showInstallModal = true
	m.executing = true

	return func() tea.Msg {
		return batchInstallStartMsg{tools: markedTools}
	}
}

// processBatchTool processes a single tool in the batch
func (m *Model) processBatchTool(index int) tea.Cmd {
	if m.batchProgress == nil || index >= len(m.batchProgress.Tools) {
		return func() tea.Msg { return batchInstallCompleteMsg{} }
	}

	tool := m.batchProgress.Tools[index]
	config := m.batchConfig

	return func() tea.Msg {
		// Get install instructions for this tool
		installs, err := m.db.GetInstallInstructions(tool.ID)
		if err != nil || len(installs) == 0 {
			if config != nil && config.SkipIfBlind {
				return batchInstallProgressMsg{
					toolID:  tool.ID,
					skipped: true,
				}
			}

			return batchInstallProgressMsg{
				toolID: tool.ID,
				err:    fmt.Errorf("no install instructions found"),
			}
		}

		// Select install command using same logic as InstallPanel
		selector := install.NewPlatformSelector("", "", "", tool.Language)
		osInfo, _ := lib.DetectOS()
		detectedOS := ""
		if osInfo != nil {
			detectedOS = osInfo.ID
		}
		platform := selector.SelectPlatform(detectedOS)

		filtered, _ := install.FilterCommands(installs, platform, tool.Language)
		if len(filtered) == 0 {
			if config != nil && config.SkipIfBlind {
				return batchInstallProgressMsg{
					toolID:  tool.ID,
					skipped: true,
				}
			}

			return batchInstallProgressMsg{
				toolID: tool.ID,
				err:    fmt.Errorf("no compatible install method"),
			}
		}

		// Get command
		defaultCmd := install.SelectDefaultCommand(filtered, false, detectedOS)
		cmd := filtered[0].Command
		if defaultCmd != nil {
			cmd = defaultCmd.Command
		}

		// Transform if mise mode
		if config != nil && config.UseMise {
			cmd = install.TransformToMise(cmd)
		}

		// Add sudo if needed
		if config != nil && config.UseSudo {
			isSystemPM := isSystemPackageManager(filtered[0].Platform)
			if !config.SudoOnlySystem || isSystemPM {
				cmd = "sudo " + cmd
			}
		}

		// Execute the command
		execCmd := exec.Command("sh", "-c", cmd) //nolint:gosec // G204: user install
		output, err := execCmd.CombinedOutput()

		return batchInstallProgressMsg{
			toolID: tool.ID,
			output: string(output),
			err:    err,
		}
	}
}

// isSystemPackageManager returns true for system package managers that typically need sudo
func isSystemPackageManager(platform string) bool {
	switch platform {
	case "apt", "dnf", "yum", "pacman", "apk", "zypper", "nix":
		return true
	default:
		return false
	}
}

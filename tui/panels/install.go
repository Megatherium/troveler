package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/install"
	"troveler/lib"
	"troveler/tui/styles"
)

// InstallPanel displays install options
type InstallPanel struct {
	commands       []install.CommandInfo
	cursor         int
	focused        bool
	cliOverride    string
	configOverride string
	fallback       string
	toolLanguage   string
	usedFallback   bool
	miseMode       bool
}

// InstallExecuteMsg is sent when user wants to execute install
type InstallExecuteMsg struct {
	Command string
}

// InstallExecuteMiseMsg is sent when user wants to execute install via mise
type InstallExecuteMiseMsg struct {
	Command string
}

// NewInstallPanel creates a new install panel
func NewInstallPanel(cliOverride, configOverride, fallback string, miseMode bool) *InstallPanel {
	return &InstallPanel{
		commands:       []install.CommandInfo{},
		cursor:         0,
		focused:        false,
		cliOverride:    cliOverride,
		configOverride: configOverride,
		fallback:       fallback,
		miseMode:       miseMode,
	}
}

// SetTool updates the install commands for a tool
func (p *InstallPanel) SetTool(tool *db.Tool, installs []db.InstallInstruction) {
	p.toolLanguage = tool.Language

	// Determine platform using priority logic
	osInfo, _ := lib.DetectOS()
	detectedOS := ""
	if osInfo != nil {
		detectedOS = osInfo.ID
	}

	// If mise mode is enabled AND no CLI override was provided, force LANG override
	// CLI parameters have higher priority than config settings
	cliOverride := p.cliOverride
	configOverride := p.configOverride

	// Resolve virtual platforms (mise:* → source platform)
	if cliOverride != "" {
		cliOverride = p.resolveVirtualPlatform(cliOverride)
	}

	if p.miseMode && p.cliOverride == "" {
		cliOverride = "LANG"
	}

	selector := install.NewPlatformSelector(
		cliOverride,
		configOverride,
		p.fallback,
		tool.Language,
	)
	platform := selector.SelectPlatform(detectedOS)

	// Filter commands based on platform
	filtered, usedFallback := install.FilterCommands(installs, platform, tool.Language)
	defaultCmd := install.SelectDefaultCommand(filtered, usedFallback, detectedOS)

	// Generate virtual install instructions
	virtuals := install.GenerateVirtualInstallInstructions(installs)

	// Format for display and transform if mise mode is enabled
	p.commands = install.FormatCommands(filtered, defaultCmd)

	// Add virtual commands to the display
	for _, v := range virtuals {
		cmd := install.CommandInfo{
			Platform:  v.Platform,
			Command:   v.Command,
			IsDefault: false, // Virtuals are never the default
		}
		p.commands = append(p.commands, cmd)
	}

	if p.miseMode {
		for i := range p.commands {
			p.commands[i].Command = install.TransformToMise(p.commands[i].Command)
		}
	}
	p.usedFallback = usedFallback
	p.cursor = 0
}

// Clear clears the install commands
func (p *InstallPanel) Clear() {
	p.commands = []install.CommandInfo{}
	p.cursor = 0
}

// GetSelectedCommand returns the currently selected install command
func (p *InstallPanel) GetSelectedCommand() string {
	if p.cursor >= 0 && p.cursor < len(p.commands) {
		return p.commands[p.cursor].Command
	}

	return ""
}

// HasCommands returns true if there are install commands available
func (p *InstallPanel) HasCommands() bool {
	return len(p.commands) > 0
}

// IsFallbackMode returns true if showing all install entries due to platform detection failure
func (p *InstallPanel) IsFallbackMode() bool {
	return p.usedFallback
}

// Update handles messages
func (p *InstallPanel) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if p.cursor < len(p.commands)-1 {
				p.cursor++
			}

			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if p.cursor > 0 {
				p.cursor--
			}

			return nil

		case msg.Alt && (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'i'):
			// Execute selected install command
			if p.cursor >= 0 && p.cursor < len(p.commands) {
				cmd := p.commands[p.cursor].Command
				// Transform command if mise mode is enabled (already done in SetTool, but for safety)
				if p.miseMode {
					cmd = install.TransformToMise(cmd)
				}

				return func() tea.Msg {
					return InstallExecuteMsg{Command: cmd}
				}
			}

			return nil

		case msg.Alt && (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'm'):
			// Execute selected install command via mise (force mise transformation)
			if p.cursor >= 0 && p.cursor < len(p.commands) {
				cmd := p.commands[p.cursor].Command
				// Always transform to mise for Alt+m
				cmd = install.TransformToMise(cmd)

				return func() tea.Msg {
					return InstallExecuteMiseMsg{Command: cmd}
				}
			}

			return nil
		}
	}

	return nil
}

// View renders the install panel
func (p *InstallPanel) View(width, height int) string {
	if len(p.commands) == 0 {
		return styles.MutedStyle.Render("Select a tool to see install options")
	}

	var b strings.Builder

	for i, cmd := range p.commands {
		// Render command
		var line string

		if i == p.cursor && p.focused {
			// Selected item
			if cmd.IsDefault {
				line = styles.SelectedStyle.Bold(true).Render(fmt.Sprintf("> %s: %s  [DEFAULT]", cmd.Platform, cmd.Command))
			} else {
				line = styles.SelectedStyle.Render(fmt.Sprintf("> %s: %s", cmd.Platform, cmd.Command))
			}
		} else {
			// Non-selected item
			if cmd.IsDefault {
				line = styles.HighlightStyle.Bold(true).Render(fmt.Sprintf("  %s: %s  [DEFAULT]", cmd.Platform, cmd.Command))
			} else {
				line = styles.UnselectedStyle.Render(fmt.Sprintf("  %s: %s", cmd.Platform, cmd.Command))
			}
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

// Focus focuses the panel
func (p *InstallPanel) Focus() {
	p.focused = true
}

// Blur unfocuses the panel
func (p *InstallPanel) Blur() {
	p.focused = false
}

// resolveVirtualPlatform resolves virtual mise:* platforms back to their source platforms
// For example: mise:github → github, mise:go → go
func (p *InstallPanel) resolveVirtualPlatform(platform string) string {
	if strings.HasPrefix(platform, "mise:") {
		return strings.TrimPrefix(platform, "mise:")
	}

	return platform
}

// IsFocused returns focus state
func (p *InstallPanel) IsFocused() bool {
	return p.focused
}

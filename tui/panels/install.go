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
	commands      []install.CommandInfo
	cursor        int
	focused       bool
	cliOverride   string
	configOverride string
	fallback      string
	toolLanguage  string
}

// InstallExecuteMsg is sent when user wants to execute install
type InstallExecuteMsg struct {
	Command string
}

// NewInstallPanel creates a new install panel
func NewInstallPanel(cliOverride, configOverride, fallback string) *InstallPanel {
	return &InstallPanel{
		commands:       []install.CommandInfo{},
		cursor:         0,
		focused:        false,
		cliOverride:    cliOverride,
		configOverride: configOverride,
		fallback:       fallback,
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

	selector := install.NewPlatformSelector(
		p.cliOverride,
		p.configOverride,
		p.fallback,
		tool.Language,
	)
	platform := selector.SelectPlatform(detectedOS)

	// Filter commands based on platform
	filtered := install.FilterCommands(installs, platform, tool.Language)
	defaultCmd := install.SelectDefaultCommand(filtered)

	// Format for display
	p.commands = install.FormatCommands(filtered, defaultCmd)
	p.cursor = 0
}

// Clear clears the install commands
func (p *InstallPanel) Clear() {
	p.commands = []install.CommandInfo{}
	p.cursor = 0
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
				return func() tea.Msg {
					return InstallExecuteMsg{Command: p.commands[p.cursor].Command}
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

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("[Alt+i to execute selected]"))

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

// IsFocused returns focus state
func (p *InstallPanel) IsFocused() bool {
	return p.focused
}

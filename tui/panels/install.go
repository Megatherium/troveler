package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/install"
	"troveler/internal/platform"
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

// isMiseLangResolved returns true if the resolved platform is mise_lang,
// meaning commands should be transformed to mise format.
func (p *InstallPanel) isMiseLangResolved() bool {
	return p.configOverride == "mise_lang" || p.fallback == "mise_lang"
}

func (p *InstallPanel) resolveCLIOverride() string {
	return platform.ResolveVirtual(p.cliOverride)
}

func (p *InstallPanel) appendVirtualCommands(installs []db.InstallInstruction) {
	virtuals := install.GenerateVirtualInstallInstructions(installs)
	for _, v := range virtuals {
		p.commands = append(p.commands, install.CommandInfo{
			Platform:  v.Platform,
			Command:   v.Command,
			IsDefault: false,
		})
	}
}

func (p *InstallPanel) transformCommandsToMise() {
	for i := range p.commands {
		p.commands[i].Command = install.TransformToMise(p.commands[i].Command)
	}
}

// SetTool updates the install commands for a tool
func (p *InstallPanel) SetTool(tool *db.Tool, installs []db.InstallInstruction) {
	p.toolLanguage = tool.Language

	// Determine platform using priority logic
	osInfo, _ := platform.DetectOS()
	detectedOS := ""
	if osInfo != nil {
		detectedOS = osInfo.ID
	}

	// Resolve virtual platforms (mise:* → source platform)
	cliOverride := p.resolveCLIOverride()
	selector := install.NewPlatformSelector(cliOverride, p.configOverride, p.fallback, tool.Language)

	// Use ResolvePlatform to try fallback_platform when detected OS yields no matches
	result := install.ResolvePlatform(selector, installs, detectedOS, tool.Language)
	defaultCmd := install.SelectDefaultCommand(result.Installs, result.UsedFallback, detectedOS)

	p.commands = install.FormatCommands(result.Installs, defaultCmd)
	p.appendVirtualCommands(installs)
	if p.isMiseLangResolved() {
		p.transformCommandsToMise()
	}
	p.usedFallback = result.UsedFallback
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
func (p *InstallPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !p.focused {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if p.cursor < len(p.commands)-1 {
				p.cursor++
			}

			return p, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if p.cursor > 0 {
				p.cursor--
			}

			return p, nil

		case msg.Alt && (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'i'):
			if p.cursor >= 0 && p.cursor < len(p.commands) {
				cmd := p.commands[p.cursor].Command

				return p, func() tea.Msg {
					return InstallExecuteMsg{Command: cmd}
				}
			}

			return p, nil

		case msg.Alt && (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'm'):
			if p.cursor >= 0 && p.cursor < len(p.commands) {
				cmd := install.TransformToMise(p.commands[p.cursor].Command)

				return p, func() tea.Msg {
					return InstallExecuteMiseMsg{Command: cmd}
				}
			}

			return p, nil
		}
	}

	return p, nil
}

// View renders the install panel
func (p *InstallPanel) View() string {
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

// SetSize stores the panel dimensions for use in View()
func (p *InstallPanel) SetSize(width, height int) {
	// InstallPanel does not use dimensions currently
}

// Init satisfies tea.Model
func (p *InstallPanel) Init() tea.Cmd {
	return nil
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

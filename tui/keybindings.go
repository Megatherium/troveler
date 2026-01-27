package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the TUI
type KeyMap struct {
	// Navigation
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Tab   key.Binding

	// Actions
	Enter  key.Binding
	Escape key.Binding
	Quit   key.Binding

	// Special actions
	Install     key.Binding // Alt+i
	Update      key.Binding // Alt+u
	Sort        key.Binding // Alt+s
	InfoModal   key.Binding // i for full-screen info modal
	Help        key.Binding // ?
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next panel"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/cancel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("alt+q", "ctrl+c"),
			key.WithHelp("alt+q", "quit"),
		),
		Install: key.NewBinding(
			key.WithKeys("alt+i"),
			key.WithHelp("alt+i", "install"),
		),
		Update: key.NewBinding(
			key.WithKeys("alt+u"),
			key.WithHelp("alt+u", "update db"),
		),
		Sort: key.NewBinding(
			key.WithKeys("alt+s"),
			key.WithHelp("alt+s", "sort"),
		),
		InfoModal: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "full info"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// ShortHelp returns a short help string
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Tab, k.Enter, k.Install, k.Update, k.Quit}
}

// FullHelp returns the full help keybindings
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.Enter, k.Escape},
		{k.Install, k.Update, k.Sort, k.InfoModal},
		{k.Help, k.Quit},
	}
}

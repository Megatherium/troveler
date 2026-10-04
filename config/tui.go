package config

import (
	"encoding/hex"
	"fmt"
)

// Resolve applies TUI defaults and validates appearance settings.
func (c TUIConfig) Resolve() (TUIConfig, error) {
	if c.Theme == "" {
		c.Theme = "gradient"
	}
	switch c.Theme {
	case "gradient", "default", "custom":
	default:
		return TUIConfig{}, fmt.Errorf("tui.theme must be gradient, default or custom; got %q", c.Theme)
	}
	if c.TaglineMaxWidth == 0 {
		c.TaglineMaxWidth = 40
	}
	if c.TaglineMaxWidth < 0 {
		return TUIConfig{}, fmt.Errorf("tui.tagline_max_width must be positive (or 0 for the default)")
	}
	if c.Theme == "custom" && len(c.GradientColors) == 0 {
		return TUIConfig{}, fmt.Errorf("tui.gradient_colors must contain at least one color for the custom theme")
	}
	for i, color := range c.GradientColors {
		if len(color) != 7 || color[0] != '#' {
			return TUIConfig{}, fmt.Errorf("tui.gradient_colors[%d] must be a #RRGGBB color; got %q", i, color)
		}
		if _, err := hex.DecodeString(color[1:]); err != nil {
			return TUIConfig{}, fmt.Errorf("tui.gradient_colors[%d] must be a #RRGGBB color; got %q", i, color)
		}
	}
	return c, nil
}

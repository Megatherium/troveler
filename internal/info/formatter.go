package info

import (
	"fmt"
	"strings"

	"troveler/db"
)

// ToolInfo represents formatted tool information
type ToolInfo struct {
	Name           string
	Tagline        string
	Description    string
	Language       string
	License        string
	Repository     string
	DatePublished  string
	Slug           string
	InstallOptions []InstallOption
}

// InstallOption represents a single install command
type InstallOption struct {
	Platform string
	Command  string
}

// FormatTool converts a db.Tool to formatted ToolInfo
func FormatTool(tool *db.Tool, installs []db.InstallInstruction) *ToolInfo {
	info := &ToolInfo{
		Name:          tool.Name,
		Tagline:       tool.Tagline,
		Description:   tool.Description,
		Language:      tool.Language,
		License:       tool.License,
		Repository:    tool.CodeRepository,
		DatePublished: tool.DatePublished,
		Slug:          tool.Slug,
	}

	for _, inst := range installs {
		info.InstallOptions = append(info.InstallOptions, InstallOption{
			Platform: inst.Platform,
			Command:  inst.Command,
		})
	}

	return info
}

// RenderPlainText renders tool info as plain text
func (ti *ToolInfo) RenderPlainText() string {
	var b strings.Builder

	b.WriteString(ti.Name)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", len(ti.Name)))
	b.WriteString("\n\n")

	if ti.Tagline != "" {
		b.WriteString(ti.Tagline)
		b.WriteString("\n\n")
	}

	if ti.Description != "" {
		b.WriteString("Description:\n")
		b.WriteString(WrapText(ti.Description, 70))
		b.WriteString("\n\n")
	}

	b.WriteString("Info:\n")
	if ti.Language != "" {
		b.WriteString(fmt.Sprintf("  Language:   %s\n", ti.Language))
	}
	if ti.License != "" {
		b.WriteString(fmt.Sprintf("  License:    %s\n", ti.License))
	}
	if ti.Repository != "" {
		b.WriteString(fmt.Sprintf("  Repository: %s\n", ti.Repository))
	}
	if ti.DatePublished != "" {
		b.WriteString(fmt.Sprintf("  Published:  %s\n", ti.DatePublished))
	}
	b.WriteString(fmt.Sprintf("  Slug:       %s\n", ti.Slug))

	if len(ti.InstallOptions) > 0 {
		b.WriteString("\nInstall Instructions:\n")
		for _, opt := range ti.InstallOptions {
			b.WriteString(fmt.Sprintf("  %s: %s\n", opt.Platform, opt.Command))
		}
	}

	return b.String()
}

// WrapText wraps text to the specified width
func WrapText(text string, width int) string {
	var result strings.Builder
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 > width {
			result.WriteString(currentLine)
			result.WriteString("\n")
			currentLine = word
		} else {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		}
	}
	if currentLine != "" {
		result.WriteString(currentLine)
	}

	return result.String()
}

// GetKeyValuePairs returns tool info as key-value pairs for table rendering
func (ti *ToolInfo) GetKeyValuePairs() [][]string {
	pairs := [][]string{}

	if ti.Language != "" {
		pairs = append(pairs, []string{"Language", ti.Language})
	}
	if ti.License != "" {
		pairs = append(pairs, []string{"License", ti.License})
	}
	if ti.Repository != "" {
		pairs = append(pairs, []string{"Repository", ti.Repository})
	}
	if ti.DatePublished != "" {
		pairs = append(pairs, []string{"Published", ti.DatePublished})
	}
	pairs = append(pairs, []string{"Slug", ti.Slug})

	return pairs
}

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"troveler/db"
)

var InfoCmd = &cobra.Command{
	Use:   "info <slug>",
	Short: "Show detailed information about a tool",
	Long:  "Show detailed information about a tool including description, repository, and install instructions.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		cfg := GetConfig(cmd.Context())
		if cfg == nil {
			return fmt.Errorf("config not loaded")
		}

		database, err := db.New(cfg.DSN)
		if err != nil {
			return fmt.Errorf("db init: %w", err)
		}
		defer database.Close()

		return runInfo(database, slug)
	},
}

func runInfo(database *db.SQLiteDB, slug string) error {
	tools, err := database.GetToolBySlug(slug)
	if err != nil {
		return fmt.Errorf("tool not found: %s", slug)
	}
	if len(tools) == 0 {
		return fmt.Errorf("tool not found: %s", slug)
	}

	tool := tools[0]
	installs, err := database.GetInstallInstructions(tool.ID)
	if err != nil {
		installs = []db.InstallInstruction{}
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render(tool.Name))
	fmt.Println(strings.Repeat("─", len(tool.Name)))
	fmt.Println()

	if tool.Tagline != "" {
		fmt.Println(lipgloss.NewStyle().Italic(true).Render(tool.Tagline))
		fmt.Println()
	}

	if tool.Description != "" {
		fmt.Println("Description:")
		fmt.Println(wrapText(tool.Description, 70))
		fmt.Println()
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Field", "Value"})

	if tool.Language != "" {
		t.AppendRow(table.Row{"Language", tool.Language})
	}
	if tool.License != "" {
		t.AppendRow(table.Row{"License", tool.License})
	}
	if tool.CodeRepository != "" {
		t.AppendRow(table.Row{"Repository", tool.CodeRepository})
	}
	if tool.DatePublished != "" {
		t.AppendRow(table.Row{"Published", tool.DatePublished})
	}
	t.AppendRow(table.Row{"Slug", tool.Slug})

	fmt.Println("Info:")
	t.Render()
	fmt.Println()

	if len(installs) > 0 {
		it := table.NewWriter()
		it.SetOutputMirror(os.Stdout)
		it.AppendHeader(table.Row{"Platform", "Command"})

		for _, inst := range installs {
			it.AppendRow(table.Row{inst.Platform, inst.Command})
		}

		fmt.Println("Install Instructions:")
		it.Render()
	}

	return nil
}

func wrapText(text string, width int) string {
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

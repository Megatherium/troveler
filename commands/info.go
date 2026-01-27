package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"troveler/db"
	"troveler/pkg/ui"
)

var InfoCmd = &cobra.Command{
	Use:   "info <slug>",
	Short: "Show detailed information about a tool",
	Long:  "Show detailed information about a tool including description, repository, and install instructions.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			return runInfo(database, slug)
		})
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
		fmt.Println(lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#AAAAAA")).Render(tool.Tagline))
		fmt.Println()
	}

	if tool.Description != "" {
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).Render("Description:"))
		fmt.Println(wrapText(tool.Description, 70))
		fmt.Println()
	}

	rows := [][]string{}
	if tool.Language != "" {
		rows = append(rows, []string{"Language", tool.Language})
	}
	if tool.License != "" {
		rows = append(rows, []string{"License", tool.License})
	}
	if tool.CodeRepository != "" {
		rows = append(rows, []string{"Repository", tool.CodeRepository})
	}
	if tool.DatePublished != "" {
		rows = append(rows, []string{"Published", tool.DatePublished})
	}
	rows = append(rows, []string{"Slug", tool.Slug})

	if len(rows) > 0 {
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).Render("Info:"))
		fmt.Println(ui.RenderKeyValueTable(rows))
		fmt.Println()
	}

	if len(installs) > 0 {
		instRows := make([][]string, len(installs))
		for i, inst := range installs {
			instRows[i] = []string{inst.Platform, inst.Command}
		}
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).Render("Install Instructions:"))
		fmt.Println(ui.RenderKeyValueTable(instRows))
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

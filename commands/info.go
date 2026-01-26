package commands

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

func renderInfoTable(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	colWidths := []int{12, 0}
	for _, row := range rows {
		if len(row[0]) > colWidths[0] {
			colWidths[0] = len(row[0])
		}
		if len(row[1]) > colWidths[1] {
			colWidths[1] = len(row[1])
		}
	}

	borderChar := "│"
	topBorder := "┌"
	botBorder := "└"
	joinChar := "┬"
	joinBot := "┴"
	rightEnd := "┐"
	rightBot := "┘"

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))

	var b strings.Builder

	b.WriteString(topBorder)
	for i, w := range colWidths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			b.WriteString(joinChar)
		}
	}
	b.WriteString(rightEnd + "\n")

	for _, row := range rows {
		b.WriteString(borderChar)
		b.WriteString(" ")
		b.WriteString(labelStyle.Render(row[0]))
		pad := colWidths[0] - len(row[0])
		b.WriteString(strings.Repeat(" ", pad+1))
		b.WriteString(borderChar)
		b.WriteString(" ")
		b.WriteString(valueStyle.Render(row[1]))
		pad = colWidths[1] - len(row[1])
		b.WriteString(strings.Repeat(" ", pad+1))
		b.WriteString(borderChar + "\n")
	}

	b.WriteString(botBorder)
	for i, w := range colWidths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			b.WriteString(joinBot)
		}
	}
	b.WriteString(rightBot)

	return b.String()
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
		fmt.Println(renderInfoTable(rows))
		fmt.Println()
	}

	if len(installs) > 0 {
		headers := []string{"Platform", "Command"}
		instRows := make([][]string, len(installs))
		for i, inst := range installs {
			instRows[i] = []string{inst.Platform, inst.Command}
		}
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).Render("Install Instructions:"))
		fmt.Println(renderInstallTable(headers, instRows))
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

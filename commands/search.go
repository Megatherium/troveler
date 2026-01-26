package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"troveler/db"
)

func renderTable(headers []string, rows [][]string) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	borderChar := "│"
	topBorder := "┌"
	midBorder := "├"
	botBorder := "└"
	joinChar := "┬"
	joinMid := "┼"
	joinBot := "┴"
	rightEnd := "┐"
	rightMid := "┤"
	rightBot := "┘"

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))

	var b strings.Builder

	b.WriteString(topBorder)
	for i, w := range colWidths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			b.WriteString(joinChar)
		}
	}
	b.WriteString(rightEnd + "\n")

	b.WriteString(borderChar)
	for i, h := range headers {
		pad := colWidths[i] - len(h)
		b.WriteString(" ")
		b.WriteString(headerStyle.Render(h))
		b.WriteString(strings.Repeat(" ", pad+1))
		b.WriteString(borderChar)
	}
	b.WriteString("\n")

	b.WriteString(midBorder)
	for i, w := range colWidths {
		b.WriteString(strings.Repeat("─", w+2))
		if i < len(colWidths)-1 {
			b.WriteString(joinMid)
		}
	}
	b.WriteString(rightMid + "\n")

	for rowIdx, row := range rows {
		b.WriteString(borderChar)
		for i, cell := range row {
			pad := colWidths[i] - len(cell)
			color := getGradientColorSimple(rowIdx)
			cellStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(color))
			b.WriteString(" ")
			b.WriteString(cellStyle.Render(cell))
			b.WriteString(strings.Repeat(" ", pad+1))
			b.WriteString(borderChar)
		}
		b.WriteString("\n")
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

var SearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the local database for tools",
	Long:  "Search for tools by name, tagline, or description in the local database.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		cfg := GetConfig(cmd.Context())
		if cfg == nil {
			return fmt.Errorf("config not loaded")
		}

		database, err := db.New(cfg.DSN)
		if err != nil {
			return fmt.Errorf("db init: %w", err)
		}
		defer database.Close()

		return runSearch(cmd.Context(), database, query)
	},
}

func runSearch(ctx context.Context, database *db.SQLiteDB, query string) error {
	results, err := database.Search(ctx, query, 50)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("No tools found matching '%s'\n", query)
		return nil
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Render(fmt.Sprintf("Found %d results for '%s'", len(results), query)))
	fmt.Println(strings.Repeat("─", len(fmt.Sprintf("Found %d results for '%s'", len(results), query))))
	fmt.Println()

	headers := []string{"#", "Name", "Tagline", "Language"}
	rows := make([][]string, len(results))
	for i, r := range results {
		tagline := r.Tagline
		if len(tagline) > 50 {
			tagline = tagline[:47] + "..."
		}
		rows[i] = []string{fmt.Sprintf("%d", i+1), r.Name, tagline, r.Language}
	}

	fmt.Println(renderTable(headers, rows))

	return nil
}

package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"troveler/db"
)

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

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleDefault)

	t.AppendHeader(table.Row{"#", "Name", "Tagline", "Language"})

	for i, r := range results {
		tagline := r.Tagline
		if len(tagline) > 50 {
			tagline = tagline[:47] + "..."
		}
		nameStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(getGradientColorSimple(i)))
		t.AppendRow(table.Row{
			fmt.Sprintf("%d", i+1),
			nameStyle.Render(r.Name),
			tagline,
			r.Language,
		})
	}

	t.Render()

	return nil
}

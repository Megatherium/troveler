package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

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

	fmt.Printf("Found %d results for '%s'\n", len(results), query)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name", "Tagline", "Language"})
	t.SetStyle(table.StyleDefault)

	for i, r := range results {
		tagline := r.Tagline
		if len(tagline) > 60 {
			tagline = tagline[:57] + "..."
		}
		t.AppendRow(table.Row{i + 1, r.Name, tagline, r.Language})
	}

	t.Render()

	return nil
}

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"troveler/db"
	"troveler/internal/search"
	"troveler/pkg/ui"
)

var SearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the local database for tools",
	Long:  "Search for tools by name, tagline, or description in the local database.",
	Args:  cobra.MinimumNArgs(1),
	Example: "troveler search go-cli --limit 10 --sort language --desc --width 40\n\n" +
		"troveler search tagline=cli\n" +
		"troveler search installed=true\n" +
		"troveler search \"name=bat | name=batcat\"\n" +
		"troveler search \"(name=git|tagline=git)&language=go\"",
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		limit, _ := cmd.Flags().GetInt("limit")
		sortField, _ := cmd.Flags().GetString("sort")
		desc, _ := cmd.Flags().GetBool("desc")
		width, _ := cmd.Flags().GetInt("width")
		format, _ := cmd.Flags().GetString("format")

		sortOrder := "ASC"
		if desc {
			sortOrder = "DESC"
		}

		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			cfg := GetConfig(ctx)
			if cfg == nil {
				return fmt.Errorf("config not loaded")
			}

			taglineWidth := cfg.Search.TaglineWidth
			if width > 0 {
				taglineWidth = width
			}

			opts := db.SearchOptions{
				Query:     query,
				Limit:     limit,
				SortField: sortField,
				SortOrder: sortOrder,
			}

			return runSearch(ctx, database, opts, taglineWidth, format)
		})
	},
}

func init() {
	SearchCmd.Flags().IntP("limit", "l", 0, "Limit number of results to display (0 for default: 50)")
	SearchCmd.Flags().StringP("sort", "s", "name", "Sort field (name, tagline, language)")
	SearchCmd.Flags().BoolP("desc", "d", false, "Sort in descending order")
	SearchCmd.Flags().IntP("width", "w", 0, "Tagline column width in characters (0 for config default)")
	SearchCmd.Flags().StringP("format", "f", "pretty", "Output format (pretty, json)")
}

type searchColumn struct {
	Header string
	Field  string
}

var searchColumns = []searchColumn{
	{"Name", "name"},
	{"Tagline", "tagline"},
	{"Language", "language"},
	{"Installed", "installed"},
}

func runSearch(
	ctx context.Context, database *db.SQLiteDB, opts db.SearchOptions,
	taglineWidth int, format string,
) error {
	searchService := search.NewService(database)

	result, err := searchService.Search(ctx, search.Options{
		Query:     opts.Query,
		Limit:     opts.Limit,
		SortField: opts.SortField,
		SortOrder: opts.SortOrder,
	})
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	results := result.Tools

	if len(results) == 0 {
		if format == "json" {
			fmt.Println("[]")
		} else {
			fmt.Printf("No tools found matching '%s'\n", opts.Query)
		}
		return nil
	}

	switch format {
	case "json":
		return outputJSON(results)
	default:
		return outputPretty(results, opts, taglineWidth)
	}
}

func outputJSON(results []db.SearchResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func outputPretty(results []db.SearchResult, opts db.SearchOptions, taglineWidth int) error {
	fmt.Println()
	title := fmt.Sprintf("Found %d results for '%s' (sorted by %s %s)",
		len(results), opts.Query, opts.SortField, opts.SortOrder)
	fmt.Println(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Render(title))
	fmt.Println(strings.Repeat("─", len(title)))
	fmt.Println()

	headers := []string{"#"}
	for _, col := range searchColumns {
		headers = append(headers, col.Header)
	}

	rows := make([][]string, len(results))
	for i, r := range results {
		row := []string{fmt.Sprintf("%d", i+1)}

		for _, col := range searchColumns {
			val := ""
			switch col.Field {
			case "name":
				val = r.Name
			case "tagline":
				val = r.Tagline
				if len(val) > taglineWidth {
					val = val[:taglineWidth-3] + "..."
				}
			case "language":
				val = r.Language
			case "installed":
				if r.Installed {
					val = "✓"
				}
			}
			row = append(row, val)
		}
		rows[i] = row
	}

	config := ui.TableConfig{
		Headers:    headers,
		Rows:       rows,
		ShowHeader: true,
	}

	fmt.Println(ui.RenderTable(config))

	return nil
}

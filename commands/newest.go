package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"troveler/db"
	"troveler/pkg/ui"
)

// NewestCmd lists the most recently published tools.
var NewestCmd = &cobra.Command{
	Use:   "newest",
	Short: "Show newest tools sorted by published date",
	Long: "Show newest tools sorted by published date descending, " +
		"with options to filter by language or tool of the week.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		onlyTotw, _ := cmd.Flags().GetBool("only-totw")
		langs, _ := cmd.Flags().GetStringSlice("lang")
		limit, _ := cmd.Flags().GetInt("limit")

		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			opts := db.SearchOptions{
				Query:     "",
				Limit:     1000,
				SortField: "date_published",
				SortOrder: "DESC",
			}

			results, err := database.Search(ctx, opts)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			filtered := filterNewestResults(results, onlyTotw, langs)

			if limit > 0 && limit < len(filtered) {
				filtered = filtered[:limit]
			}

			return outputNewest(filtered)
		})
	},
}

func init() {
	NewestCmd.Flags().Bool("only-totw", false, "Only show tool of the week")
	NewestCmd.Flags().StringSliceP("lang", "l", []string{}, "Filter by language(s) (comma-separated)")
	NewestCmd.Flags().IntP("limit", "n", 10, "Limit number of results (post-filter)")
}

func filterNewestResults(results []db.SearchResult, onlyTotw bool, langs []string) []db.SearchResult {
	var filtered []db.SearchResult

	langSet := make(map[string]bool)
	for _, lang := range langs {
		langSet[strings.ToLower(lang)] = true
	}

	for _, r := range results {
		if onlyTotw && !r.ToolOfTheWeek {
			continue
		}

		if len(langSet) > 0 && !langSet[strings.ToLower(r.Language)] {
			continue
		}

		filtered = append(filtered, r)
	}

	return filtered
}

func formatDate(dateStr string) string {
	if dateStr == "" {
		return "unknown"
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}

	return t.Format("2006-01-02")
}

func outputNewest(results []db.SearchResult) error {
	if len(results) == 0 {
		fmt.Println("No tools found.")

		return nil
	}

	fmt.Println()
	title := "Newest Tools"
	fmt.Println(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Render(title))
	fmt.Println(strings.Repeat("─", len(title)))
	fmt.Println()

	for i, r := range results {
		nameStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ui.GradientColors[i%len(ui.GradientColors)]))

		taglineStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA"))

		dateStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

		totwStar := ""
		if r.ToolOfTheWeek {
			totwStar = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Render(" ⭐")
		}

		line := fmt.Sprintf("%s - %s - %s%s",
			nameStyle.Render(r.Name),
			taglineStyle.Render(r.Tagline),
			dateStyle.Render(formatDate(r.DatePublished)),
			totwStar,
		)

		fmt.Println(line)
	}

	fmt.Println()

	return nil
}

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
)

var TagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage user-curated tags on tools",
	Long: `Manage user-curated tags on tools.

Tags allow you to organize tools with your own labels. Use subcommands
to add, remove, list, or clear tags.`,
}

var tagAddCmd = &cobra.Command{
	Use:     "add <slug> <tag>",
	Short:   "Add a tag to a tool",
	Args:    cobra.ExactArgs(2),
	Example: "  troveler tag add fzf fuzzy",
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]
		tag := args[1]

		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			if err := database.AddTag(slug, tag); err != nil {
				return fmt.Errorf("failed to add tag: %w", err)
			}
			fmt.Printf("Added tag '%s' to '%s'\n", tag, slug)
			return nil
		})
	},
}

var tagRemoveCmd = &cobra.Command{
	Use:     "remove <slug> <tag>",
	Short:   "Remove a tag from a tool",
	Args:    cobra.ExactArgs(2),
	Example: "  troveler tag remove fzf cli",
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]
		tag := args[1]

		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			if err := database.RemoveTag(slug, tag); err != nil {
				return fmt.Errorf("failed to remove tag: %w", err)
			}
			fmt.Printf("Removed tag '%s' from '%s'\n", tag, slug)
			return nil
		})
	},
}

var tagListCmd = &cobra.Command{
	Use:     "list [slug]",
	Short:   "List tags (all tags or tags for a specific tool)",
	Args:    cobra.MaximumNArgs(1),
	Example: "  troveler tag list\n  troveler tag list fzf",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			if len(args) == 0 {
				return listAllTags(database, jsonOutput)
			}
			return listToolTags(database, args[0], jsonOutput)
		})
	},
}

var tagClearCmd = &cobra.Command{
	Use:     "clear <slug>",
	Short:   "Remove all tags from a tool",
	Args:    cobra.ExactArgs(1),
	Example: "  troveler tag clear fzf",
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			if err := database.ClearTags(slug); err != nil {
				return fmt.Errorf("failed to clear tags: %w", err)
			}
			fmt.Printf("Cleared all tags from '%s'\n", slug)
			return nil
		})
	},
}

func init() {
	tagListCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	TagCmd.AddCommand(tagAddCmd)
	TagCmd.AddCommand(tagRemoveCmd)
	TagCmd.AddCommand(tagListCmd)
	TagCmd.AddCommand(tagClearCmd)
}

func listAllTags(database *db.SQLiteDB, jsonOutput bool) error {
	tags, err := database.GetAllTags()
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			fmt.Println("No tags found")
		}
		return nil
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(tags)
	}

	fmt.Println()
	title := fmt.Sprintf("All tags (%d)", len(tags))
	fmt.Println(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Render(title))
	fmt.Println(strings.Repeat("─", len(title)))
	fmt.Println()

	for _, tc := range tags {
		fmt.Printf("  %-20s %d\n", tc.Name, tc.Count)
	}
	fmt.Println()
	return nil
}

func listToolTags(database *db.SQLiteDB, slug string, jsonOutput bool) error {
	tags, err := database.GetTags(slug)
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		if jsonOutput {
			fmt.Println("[]")
		} else {
			fmt.Printf("No tags on '%s'\n", slug)
		}
		return nil
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(tags)
	}

	fmt.Printf("Tags on '%s': %s\n", slug, strings.Join(tags, ", "))
	return nil
}

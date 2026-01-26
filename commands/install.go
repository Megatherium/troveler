package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"troveler/config"
	"troveler/db"
	"troveler/lib"
)

var all bool

var InstallCmd = &cobra.Command{
	Use:   "install <slug>",
	Short: "Show install command for a tool",
	Long: `Show the appropriate install command for the current OS.
Without flags, shows only the command for your detected OS.
Use --all to show all available install commands.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("config load: %w", err)
		}

		database, err := db.New(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("db init: %w", err)
		}
		defer database.Close()

		return runInstall(database, slug, all)
	},
}

func init() {
	InstallCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all install commands")
}

func runInstall(database *db.SQLiteDB, slug string, showAll bool) error {
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

	if len(installs) == 0 {
		return fmt.Errorf("no install instructions available for %s", slug)
	}

	if showAll {
		return showAllInstalls(tool.Name, installs)
	}

	osInfo, err := lib.DetectOS()
	if err != nil {
		fmt.Printf("Warning: Could not detect OS: %v\n\n", err)
		return showAllInstalls(tool.Name, installs)
	}

	matched := findMatchingInstalls(osInfo.ID, installs)
	if len(matched) == 0 {
		fmt.Printf("No install command found for %s (%s).\n\n", osInfo.Name, osInfo.ID)
		fmt.Println("Available commands:")
		return showAllInstalls(tool.Name, installs)
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).Render("Install command for " + osInfo.Name + ":"))
	fmt.Println()

	for _, inst := range matched {
		fmt.Println(lipgloss.NewStyle().Bold(true).Render(inst.Command))
	}
	fmt.Println()

	return nil
}

func showAllInstalls(name string, installs []db.InstallInstruction) error {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render(name + " - All Install Commands:"))
	fmt.Println(strings.Repeat("─", len(name)+len(" - All Install Commands:")))
	fmt.Println()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Platform", "Command"})

	for _, inst := range installs {
		t.AppendRow(table.Row{inst.Platform, inst.Command})
	}

	t.Render()
	return nil
}

func findMatchingInstalls(osID string, installs []db.InstallInstruction) []db.InstallInstruction {
	var matched []db.InstallInstruction

	for _, inst := range installs {
		if lib.MatchPlatform(osID, inst.Platform) {
			matched = append(matched, inst)
		}
	}

	return matched
}

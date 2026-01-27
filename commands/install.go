package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"troveler/db"
	"troveler/lib"
	"troveler/pkg/ui"
)

var all bool
var run bool
var sudo bool

var InstallCmd = &cobra.Command{
	Use:   "install <slug> [platform]",
	Short: "Show install command for a tool",
	Long: `Show that appropriate install command for your current OS.
Without flags, shows only the command for your detected OS.
Use --all to show all available install commands.
Specify a platform as a second argument to show commands for that platform.
Use --run to execute the install command after confirmation.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]
		platform := ""
		if len(args) > 1 {
			platform = args[1]
		}

		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			cfg := GetConfig(ctx)
			return runInstall(database, slug, all, run, sudo, platform, cfg.Install.FallbackPlatform, cfg.Install.AlwaysRun, cfg.Install.UseSudo)
		})
	},
}

func init() {
	InstallCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all install commands")
	InstallCmd.Flags().BoolVarP(&run, "run", "r", false, "Run the install command after confirmation")
	InstallCmd.Flags().BoolVarP(&sudo, "sudo", "s", false, "Prepend sudo to the install command")
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

func runInstall(database *db.SQLiteDB, slug string, showAll bool, run bool, sudo bool, platform string, fallbackPlatform string, alwaysRun bool, useSudo string) error {
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

	var matched []db.InstallInstruction

	if platform != "" {
		normalizedPlatform := lib.NormalizePlatform(platform)
		if normalizedPlatform != platform {
			platform = normalizedPlatform
		}
		matched = findMatchingInstalls(platform, installs)
	} else {
		osInfo, err := lib.DetectOS()
		if err != nil {
			fmt.Printf("Warning: Could not detect OS: %v\n\n", err)
			if fallbackPlatform != "" {
				fallback := lib.NormalizePlatform(fallbackPlatform)
				if fallback == "LANG" {
					fallback = tool.Language
					for _, inst := range installs {
						if lib.MatchLanguage(fallback, inst.Platform) {
							matched = append(matched, inst)
						}
					}
					platform = fallback
				} else {
					matched = findMatchingInstalls(fallback, installs)
					platform = fallback
				}
			}
		} else {
			matched = findMatchingInstalls(osInfo.ID, installs)
			platform = osInfo.ID
		}
	}

	if len(matched) == 0 {
		fmt.Printf("No install command found for %s.\n\n", platform)
		if platform != "LANG" && platform != tool.Language {
			var langMatched []db.InstallInstruction
			for _, inst := range installs {
				if lib.MatchLanguage(tool.Language, inst.Platform) {
					langMatched = append(langMatched, inst)
				}
			}
			if len(langMatched) > 0 {
				fmt.Printf("Trying language (%s):\n", tool.Language)
				for _, inst := range langMatched {
					fmt.Println(lipgloss.NewStyle().Bold(true).Render(inst.Command))
				}
				fmt.Println()
				return nil
			}
		}
		fmt.Println("Available commands:")
		return showAllInstalls(tool.Name, installs)
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).Render("Install command for " + platform + ":"))
	fmt.Println()

	for _, inst := range matched {
		fmt.Println(lipgloss.NewStyle().Bold(true).Render(inst.Command))
	}
	fmt.Println()

	if run {
		return executeInstall(matched[0].Command, sudo, useSudo, alwaysRun)
	}

	return nil
}

func executeInstall(command string, sudo bool, useSudo string, alwaysRun bool) error {
	shouldSudo := sudo

	if !sudo && useSudo == "ask" {
		fmt.Print("Use sudo? [y/N] ")
		var confirm string
		if _, err := fmt.Scanln(&confirm); err != nil {
			return fmt.Errorf("aborted")
		}
		shouldSudo = confirm == "y" || confirm == "Y"
	} else if !sudo && useSudo == "true" {
		shouldSudo = true
	}

	if shouldSudo {
		command = "sudo " + command
	}

	if !alwaysRun {
		fmt.Print("Execute this command? [y/N] ")
		var confirm string
		if _, err := fmt.Scanln(&confirm); err != nil {
			return fmt.Errorf("aborted")
		}

		if confirm != "y" && confirm != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	fmt.Printf("\nExecuting: %s\n\n", command)
	return nil
}

func showAllInstalls(name string, installs []db.InstallInstruction) error {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render(name + " - All Install Commands:"))
	fmt.Println(strings.Repeat("─", len(name)+len(" - All Install Commands:")))
	fmt.Println()

	headers := []string{"Platform", "Command"}
	rows := make([][]string, len(installs))
	for i, inst := range installs {
		rows[i] = []string{inst.Platform, inst.Command}
	}

	config := ui.TableConfig{
		Headers: headers,
		Rows:    rows,
		RowFunc: func(cell string, rowIdx, colIdx int) string {
			var color string
			if colIdx == 0 {
				color = ui.GetGradientColorSimple(rowIdx)
			} else {
				color = ui.GetGradientColorSimple((rowIdx + len(rows)/2) % len(ui.GradientColors))
			}
			return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(cell)
		},
		ShowHeader: true,
	}

	fmt.Println(ui.RenderTable(config))

	return nil
}

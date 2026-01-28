package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"troveler/db"
	"troveler/internal/install"
	"troveler/lib"
	"troveler/pkg/ui"
)

var all bool
var run bool
var sudo bool
var override string
var mise bool

var InstallCmd = &cobra.Command{
	Use:   "install <slug>",
	Short: "Show install command for a tool",
	Long: `Show that appropriate install command for your current OS.
Without flags, shows only the command for your detected OS.
Use --all to show all install commands.
Use --override to specify a platform manually.
Use --run to execute the install command after confirmation.
Use --mise to output mise use commands for language-based installations.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			cfg := GetConfig(ctx)
			miseEnabled := mise || cfg.Install.MiseMode
			return runInstall(database, slug, all, run, sudo, override, cfg.Install.PlatformOverride, cfg.Install.FallbackPlatform, cfg.Install.AlwaysRun, cfg.Install.UseSudo, miseEnabled)
		})
	},
}

func init() {
	InstallCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all install commands")
	InstallCmd.Flags().StringVarP(&override, "override", "o", "", "Override platform detection (e.g., macos, linux:arch, LANG)")
	InstallCmd.Flags().BoolVarP(&run, "run", "r", false, "Run the install command after confirmation")
	InstallCmd.Flags().BoolVarP(&sudo, "sudo", "s", false, "Prepend sudo to the install command")
	InstallCmd.Flags().BoolVar(&mise, "mise", false, "Output mise use commands for language-based installations (forces LANG override)")
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

func runInstall(database *db.SQLiteDB, slug string, showAll bool, run bool, sudo bool, cliOverride string, configOverride string, fallbackPlatform string, alwaysRun bool, useSudo string, miseEnabled bool) error {
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

	// Priority: mise mode (force LANG) > CLI override > config override > OS detection > fallback
	var platform string
	var matched []db.InstallInstruction

	// If mise mode is enabled, force LANG override
	if miseEnabled {
		cliOverride = "LANG"
	}

	// Check for CLI or config override first
	override := selectOverride(cliOverride, configOverride)
	if override != "" {
		// Override is set, use it
		if override == "LANG" {
			// Use language matching
			for _, inst := range installs {
				if lib.MatchLanguage(tool.Language, inst.Platform) {
					matched = append(matched, inst)
				}
			}
			platform = tool.Language
		} else {
			// Use platform matching
			platform = lib.NormalizePlatform(override)
			matched = findMatchingInstalls(platform, installs)
		}
	} else {
		// No override, try OS detection first
		osInfo, err := lib.DetectOS()
		if err == nil {
			platform = osInfo.ID
			matched = findMatchingInstalls(platform, installs)
		}

		// If OS detection failed or no match, try fallback
		if len(matched) == 0 && fallbackPlatform != "" {
			if fallbackPlatform == "LANG" {
				// Use language matching
				for _, inst := range installs {
					if lib.MatchLanguage(tool.Language, inst.Platform) {
						matched = append(matched, inst)
					}
				}
				platform = tool.Language
			} else {
				// Use platform matching
				platform = lib.NormalizePlatform(fallbackPlatform)
				matched = findMatchingInstalls(platform, installs)
			}
		}

		// If still no match and OS detection failed, show error
		if len(matched) == 0 && err != nil {
			fmt.Printf("Warning: Could not detect OS: %v\n\n", err)
			fmt.Println("Available commands:")
			return showAllInstalls(tool.Name, installs)
		}
	}

	// No match found, try fallback then show all
	if len(matched) == 0 {
		fmt.Printf("No install command found for %s.\n\n", platform)

		// If we haven't tried language matching yet, try it as fallback
		if platform != tool.Language {
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
		cmd := inst.Command
		if miseEnabled {
			cmd = install.TransformToMise(cmd)
		}
		fmt.Println(lipgloss.NewStyle().Bold(true).Render(cmd))
	}
	fmt.Println()

	if run {
		cmd := matched[0].Command
		if miseEnabled {
			cmd = install.TransformToMise(cmd)
		}
		return executeInstall(cmd, sudo, useSudo, alwaysRun)
	}

	return nil
}

// selectOverride returns CLI override if set, otherwise config override
func selectOverride(cliOverride, configOverride string) string {
	if cliOverride != "" {
		return cliOverride
	}
	return configOverride
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

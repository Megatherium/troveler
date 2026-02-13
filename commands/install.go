package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"troveler/config"
	"troveler/db"
	"troveler/internal/install"
	"troveler/lib"
	"troveler/pkg/ui"
)

var all bool
var run bool
var sudo bool
var sudoOnlySystem bool
var override string
var mise bool
var reuseConfig string
var skipIfBlind bool

var InstallCmd = &cobra.Command{
	Use:   "install <slug> [slug2] [slug3]...",
	Short: "Show install command for one or more tools",
	Long: `Show the appropriate install command for your current OS.
Without flags, shows only the command for your detected OS.
Use --all to show all install commands.
Use --override to specify a platform manually.
Use --run to execute the install command after confirmation.
Use --mise to output mise use commands for language-based installations.

For multiple tools:
  --reuse-config: true (use same config for all), ask (prompt), false (configure each)
  --sudo-only-system: use sudo only for system package managers (apt, dnf, etc.)
  --skip-if-blind: skip tools without a compatible install method`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			cfg := GetConfig(ctx)
			miseEnabled := mise || cfg.Install.MiseMode

			if len(args) == 1 {
				// Single tool - use existing logic
				return runInstall(
					database, args[0], all, run, sudo, override,
					cfg.Install.PlatformOverride, cfg.Install.FallbackPlatform,
					cfg.Install.AlwaysRun, cfg.Install.UseSudo, miseEnabled,
				)
			}

			// Multiple tools - batch install
			return runBatchInstall(database, args, run, sudo, sudoOnlySystem, skipIfBlind, miseEnabled, reuseConfig, cfg)
		})
	},
}

func init() {
	InstallCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all install commands")
	InstallCmd.Flags().StringVarP(&override, "override", "o", "",
		"Override platform detection (e.g., macos, linux:arch, LANG)")
	InstallCmd.Flags().BoolVarP(&run, "run", "r", false, "Run the install command after confirmation")
	InstallCmd.Flags().BoolVarP(&sudo, "sudo", "s", false, "Prepend sudo to the install command")
	InstallCmd.Flags().BoolVar(&mise, "mise", false,
		"Output mise use commands for language-based installations (forces LANG override)")
	InstallCmd.Flags().StringVar(&reuseConfig, "reuse-config", "ask", "Reuse config for all tools: true, ask, false")
	InstallCmd.Flags().BoolVar(&sudoOnlySystem, "sudo-only-system", false, "Use sudo only for system package managers")
	InstallCmd.Flags().BoolVar(&skipIfBlind, "skip-if-blind", false, "Skip tools without compatible install method")
}

func FindMatchingInstalls(osID string, installs []db.InstallInstruction) []db.InstallInstruction {
	var matched []db.InstallInstruction

	for _, inst := range installs {
		if lib.MatchPlatform(osID, inst.Platform) {
			matched = append(matched, inst)
		}
	}

	return matched
}

func ResolveVirtualPlatform(platform string) string {
	if strings.HasPrefix(platform, "mise:") {
		return strings.TrimPrefix(platform, "mise:")
	}
	return platform
}

func runInstall(
	database *db.SQLiteDB, slug string, showAll bool, run bool, sudo bool,
	cliOverride string, configOverride string, fallbackPlatform string,
	alwaysRun bool, useSudo string, miseEnabled bool,
) error {
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

	// If mise mode is enabled AND no CLI override was provided, force LANG override
	// CLI parameters have higher priority than config settings
	if miseEnabled && override == "" {
		cliOverride = "LANG"
	}

	// Check for CLI or config override first
	override := selectOverride(cliOverride, configOverride)

	// Resolve virtual platforms (mise:* → source platform)
	override = ResolveVirtualPlatform(override)

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
			matched = FindMatchingInstalls(platform, installs)
		}
	} else {
		// No override, try OS detection first
		osInfo, err := lib.DetectOS()
		if err == nil {
			platform = osInfo.ID
			matched = FindMatchingInstalls(platform, installs)
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
				matched = FindMatchingInstalls(platform, installs)
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
	fmt.Println(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Render("Install command for " + platform + ":"))
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
			shouldSudo = false
		} else {
			shouldSudo = confirm == "y" || confirm == "Y"
		}
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
			confirm = ""
		}

		if confirm != "y" && confirm != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	fmt.Printf("\nExecuting: %s\n\n", command)

	cmd := exec.Command("sh", "-c", command) //nolint:noctx
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func showAllInstalls(name string, installs []db.InstallInstruction) error {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Render(name + " - All Install Commands:"))
	fmt.Println(strings.Repeat("─", len(name)+len(" - All Install Commands:")))
	fmt.Println()

	// Generate virtual install instructions from raw commands
	virtuals := install.GenerateVirtualInstallInstructions(installs)

	// Combine raw installs with virtual installs
	headers := []string{"Platform", "Command"}

	// Calculate total rows needed
	totalRows := len(installs) + len(virtuals)
	rows := make([][]string, 0, totalRows)

	// Add raw install instructions first
	for _, inst := range installs {
		rows = append(rows, []string{inst.Platform, inst.Command})
	}

	// Add virtual install instructions
	for _, v := range virtuals {
		rows = append(rows, []string{v.Platform, v.Command})
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

// BatchConfig holds configuration for batch installs
type BatchConfig struct {
	UseSudo        bool
	SudoOnlySystem bool
	SkipIfBlind    bool
	UseMise        bool
	AlwaysRun      bool
}

// runBatchInstall handles installing multiple tools
func runBatchInstall(
	database *db.SQLiteDB, slugs []string, run, sudo, sudoOnlySystem, skipIfBlind, miseEnabled bool,
	reuseConfig string, cfg *config.Config,
) error {
	fmt.Printf("\n🔧 Batch Install: %d tools\n\n", len(slugs))

	// Determine if we should ask for config
	var batchCfg *BatchConfig
	shouldReuse := reuseConfig == "true"
	shouldAsk := reuseConfig == "ask"

	if shouldAsk {
		fmt.Print("Use same configuration for all tools? [Y/n] ")
		var confirm string
		fmt.Scanln(&confirm)
		shouldReuse = confirm != "n" && confirm != "N"
	}

	if shouldReuse {
		batchCfg = &BatchConfig{
			UseSudo:        sudo,
			SudoOnlySystem: sudoOnlySystem,
			SkipIfBlind:    skipIfBlind,
			UseMise:        miseEnabled,
			AlwaysRun:      cfg.Install.AlwaysRun,
		}

		// Ask for config if flags not set
		if !sudo && !sudoOnlySystem {
			fmt.Print("Use sudo? [y/N/s=system-only] ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm == "y" || confirm == "Y" {
				batchCfg.UseSudo = true
			} else if confirm == "s" || confirm == "S" {
				batchCfg.SudoOnlySystem = true
			}
		}

		if !skipIfBlind {
			fmt.Print("Skip tools without install method? [y/N] ")
			var confirm string
			fmt.Scanln(&confirm)
			batchCfg.SkipIfBlind = confirm == "y" || confirm == "Y"
		}

		if !miseEnabled {
			fmt.Print("Use mise for installations? [y/N] ")
			var confirm string
			fmt.Scanln(&confirm)
			batchCfg.UseMise = confirm == "y" || confirm == "Y"
		}
	}

	var completed, failed, skipped []string

	for i, slug := range slugs {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(slugs), slug)
		fmt.Println(strings.Repeat("─", 40))

		err := installSingleTool(database, slug, batchCfg, run, cfg)
		if err != nil {
			if strings.Contains(err.Error(), "skipped") {
				skipped = append(skipped, slug)
				fmt.Printf("○ Skipped: %s\n", slug)
			} else {
				failed = append(failed, slug)
				fmt.Printf("✗ Failed: %s - %v\n", slug, err)
			}
		} else {
			completed = append(completed, slug)
			fmt.Printf("✓ Completed: %s\n", slug)
		}
	}

	// Summary
	fmt.Printf("\n%s\n", strings.Repeat("═", 40))
	fmt.Printf("Batch Install Summary:\n")
	if len(completed) > 0 {
		fmt.Printf("  ✓ Completed: %d\n", len(completed))
	}
	if len(failed) > 0 {
		fmt.Printf("  ✗ Failed: %d\n", len(failed))
	}
	if len(skipped) > 0 {
		fmt.Printf("  ○ Skipped: %d\n", len(skipped))
	}

	return nil
}

// installSingleTool installs a single tool with the given config
func installSingleTool(database *db.SQLiteDB, slug string, batchCfg *BatchConfig, run bool, cfg *config.Config) error {
	tools, err := database.GetToolBySlug(slug)
	if err != nil || len(tools) == 0 {
		return fmt.Errorf("tool not found: %s", slug)
	}

	tool := tools[0]
	installs, err := database.GetInstallInstructions(tool.ID)
	if err != nil || len(installs) == 0 {
		if batchCfg != nil && batchCfg.SkipIfBlind {
			return fmt.Errorf("skipped: no install instructions")
		}
		return fmt.Errorf("no install instructions available")
	}

	// Detect OS
	osInfo, _ := lib.DetectOS()
	detectedOS := ""
	if osInfo != nil {
		detectedOS = osInfo.ID
	}

	// Select platform
	selector := install.NewPlatformSelector("", cfg.Install.PlatformOverride, cfg.Install.FallbackPlatform, tool.Language)
	platform := selector.SelectPlatform(detectedOS)

	// Find matching install
	var matched []db.InstallInstruction
	for _, inst := range installs {
		if lib.MatchPlatform(platform, inst.Platform) || lib.MatchLanguage(tool.Language, inst.Platform) {
			matched = append(matched, inst)
		}
	}

	if len(matched) == 0 {
		if batchCfg != nil && batchCfg.SkipIfBlind {
			return fmt.Errorf("skipped: no compatible install method")
		}
		return fmt.Errorf("no compatible install method for %s", platform)
	}

	// Get command
	cmd := matched[0].Command

	// Transform if mise
	if batchCfg != nil && batchCfg.UseMise {
		cmd = install.TransformToMise(cmd)
	}

	// Add sudo if needed
	if batchCfg != nil {
		if batchCfg.UseSudo {
			cmd = "sudo " + cmd
		} else if batchCfg.SudoOnlySystem && isSystemPM(matched[0].Platform) {
			cmd = "sudo " + cmd
		}
	}

	fmt.Printf("Command: %s\n", cmd)

	if !run {
		return nil
	}

	// Execute
	alwaysRun := cfg.Install.AlwaysRun
	if batchCfg != nil {
		alwaysRun = batchCfg.AlwaysRun
	}

	if !alwaysRun {
		fmt.Print("Execute? [y/N] ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			return fmt.Errorf("skipped: user declined")
		}
	}

	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}

// isSystemPM returns true for system package managers
func isSystemPM(platform string) bool {
	switch platform {
	case "apt", "dnf", "yum", "pacman", "apk", "zypper", "nix":
		return true
	default:
		return false
	}
}

package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"troveler/config"
	"troveler/db"
	"troveler/internal/install"
	"troveler/internal/platform"
)

var all bool
var run bool
var sudo bool
var sudoOnlySystem bool
var override string
var mise bool
var reuseConfig string
var skipIfBlind bool

// InstallCmd shows or executes install commands for one or more tools.
var InstallCmd = &cobra.Command{
	Use:   "install <slug> [slug2] [slug3]...",
	Short: "Show install command for one or more tools",
	Long: `Show the appropriate install command for your current OS.
Without flags, shows only the command for your detected OS.
Use --all to show all install commands.
Use --override to specify a platform manually (e.g., macos, lang, mise_lang).
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

			if len(args) == 1 {
				return runInstall(
					database, args[0], all, run, sudo, override,
					cfg.Install.PlatformOverride, cfg.Install.FallbackPlatform,
					cfg.Install.AlwaysRun, cfg.Install.UseSudo,
				)
			}

			return runBatchInstall(database, args, run, sudo, sudoOnlySystem, skipIfBlind, reuseConfig, cfg)
		})
	},
}

func init() {
	InstallCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all install commands")
	InstallCmd.Flags().StringVarP(&override, "override", "o", "",
		"Override platform detection (e.g., macos, linux:arch, lang, mise_lang)")
	InstallCmd.Flags().BoolVarP(&run, "run", "r", false, "Run the install command after confirmation")
	InstallCmd.Flags().BoolVarP(&sudo, "sudo", "s", false, "Prepend sudo to the install command")
	InstallCmd.Flags().BoolVar(&mise, "mise", false,
		"Output mise use commands for language-based installations (shorthand for --override mise_lang)")
	InstallCmd.Flags().StringVar(&reuseConfig, "reuse-config", "ask", "Reuse config for all tools: true, ask, false")
	InstallCmd.Flags().BoolVar(&sudoOnlySystem, "sudo-only-system", false, "Use sudo only for system package managers")
	InstallCmd.Flags().BoolVar(&skipIfBlind, "skip-if-blind", false, "Skip tools without compatible install method")
}

func runInstall(
	database *db.SQLiteDB, slug string, showAll bool, runFlag bool, sudoFlag bool,
	cliOverride string, configOverride string, fallbackPlatform string,
	alwaysRun bool, useSudo string,
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

	// --mise flag acts as shorthand for --override mise_lang when no explicit override given
	if mise && cliOverride == "" {
		cliOverride = PlatformMiseLang
	}

	selector := platform.NewSelector(cliOverride, configOverride, fallbackPlatform, tool.Language)

	osInfo, _ := platform.DetectOS()
	detectedOS := ""
	if osInfo != nil {
		detectedOS = osInfo.ID
	}

	platformID := selector.Select(detectedOS)

	// Check for explicit virtual platform requests (e.g., "mise:github")
	// before resolving the virtual prefix
	if strings.HasPrefix(platformID, "mise:") {
		virtuals := install.GenerateVirtualInstallInstructions(installs)
		requestedBackend := strings.TrimPrefix(platformID, "mise:")

		for _, v := range virtuals {
			if v.Platform == platformID {
				displayInstallCommands(platformID, []db.InstallInstruction{
					{Platform: v.Platform, Command: v.Command},
				})

				if runFlag {
					return executeInstall(v.Command, sudoFlag, useSudo, alwaysRun)
				}

				return nil
			}
		}

		// Virtual platform not found - show helpful error
		fmt.Printf("No install command found for %s (backend: %s).\n\n", platformID, requestedBackend)

		if len(virtuals) > 0 {
			fmt.Println("Available virtual platforms:")
			for _, v := range virtuals {
				fmt.Printf("  - %s\n", v.Platform)
			}
			fmt.Println()
		}

		fmt.Println("Available commands:")

		return showAllInstalls(tool.Name, installs)
	}

	platformID = platform.ResolveVirtual(platformID)

	// Use ResolvePlatform to try fallback_platform when detected OS yields no matches
	result := install.ResolvePlatform(selector, installs, detectedOS, tool.Language)
	resolvedID := platform.ResolveVirtual(result.PlatformID)

	if result.UsedFallback && resolvedID == platformID {
		displayNoInstallMethod(tool.Name, platformID, installs)

		return nil
	}

	matched := result.Installs

	if len(matched) == 0 {
		fmt.Printf("No install command found for %s.\n\n", platformID)

		if !platform.IsLangPlatform(platformID) && platformID != tool.Language {
			var langMatched []db.InstallInstruction
			langMatched, _ = platform.FilterDBInstalls(installs, PlatformLang, tool.Language)
			if len(langMatched) > 0 {
				displayLanguageFallback(tool.Language, langMatched)

				return nil
			}
		}

		fmt.Println("Available commands:")

		return showAllInstalls(tool.Name, installs)
	}

	displayInstallCommands(resolvedID, matched)

	if runFlag {
		cmd := matched[0].Command
		if platform.IsMiseLangPlatform(resolvedID) {
			cmd = install.TransformToMise(cmd)
		}

		return executeInstall(cmd, sudoFlag, useSudo, alwaysRun)
	}

	return nil
}

// BatchConfig holds user preferences for batch install mode.
type BatchConfig struct {
	UseSudo        bool
	SudoOnlySystem bool
	SkipIfBlind    bool
	UseMise        bool
	AlwaysRun      bool
}

func runBatchInstall(
	database *db.SQLiteDB, slugs []string, runFlag, sudoFlag, sudoOnlySystemFlag, skipIfBlindFlag bool,
	reuseConfig string, cfg *config.Config,
) error {
	fmt.Printf("\n🔧 Batch Install: %d tools\n\n", len(slugs))

	batchCfg := promptBatchConfig(
		sudoFlag, sudoOnlySystemFlag, skipIfBlindFlag,
		reuseConfig, cfg.Install.AlwaysRun,
	)

	var completed, failed, skipped []string

	for i, slug := range slugs {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(slugs), slug)
		fmt.Println(strings.Repeat("─", 40))

		err := installSingleTool(database, slug, batchCfg, runFlag, cfg)
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

func installSingleTool(
	database *db.SQLiteDB, slug string, batchCfg *BatchConfig, runFlag bool, cfg *config.Config,
) error {
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

	osInfo, _ := platform.DetectOS()
	detectedOS := ""
	if osInfo != nil {
		detectedOS = osInfo.ID
	}

	selector := platform.NewSelector("", cfg.Install.PlatformOverride, cfg.Install.FallbackPlatform, tool.Language)

	result := install.ResolvePlatform(selector, installs, detectedOS, tool.Language)
	matched := result.Installs

	if result.UsedFallback || len(matched) == 0 {
		if batchCfg != nil && batchCfg.SkipIfBlind {
			return fmt.Errorf("skipped: no compatible install method")
		}

		return fmt.Errorf("no compatible install method for %s", result.PlatformID)
	}

	cmd := matched[0].Command

	if batchCfg != nil && batchCfg.UseMise {
		cmd = install.TransformToMise(cmd)
	}

	if batchCfg != nil {
		if batchCfg.UseSudo {
			cmd = "sudo " + cmd
		} else if batchCfg.SudoOnlySystem && isSystemPM(matched[0].Platform) {
			cmd = "sudo " + cmd
		}
	}

	fmt.Printf("Command: %s\n", cmd)

	if !runFlag {
		return nil
	}

	alwaysRun := cfg.Install.AlwaysRun
	if batchCfg != nil {
		alwaysRun = batchCfg.AlwaysRun
	}

	if !alwaysRun {
		if err := promptExecute(); err != nil {
			return err
		}
	}

	execCmd := exec.Command("sh", "-c", cmd) //nolint:noctx,gosec // G204: user-requested install command
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

func isSystemPM(platformID string) bool {
	switch platformID {
	case "apt", "dnf", "yum", "pacman", "apk", "zypper", "nix":
		return true
	default:
		return false
	}
}

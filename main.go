package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"troveler/commands"
)

var RootCmd = &cobra.Command{
	Use:   "troveler",
	Short: "Local terminaltrove.com mirror",
	Long: `Troveler is a CLI tool that creates a local searchable copy of terminaltrove.com.

Use 'troveler update' to fetch all tools from terminaltrove.com.
Use 'troveler search <query>' to search your local database.
Use 'troveler info <slug>' to see details of a specific tool.
Use 'troveler install <slug>' to get install commands for your OS.
Use 'troveler tui' to launch the interactive Terminal User Interface.`,
	Version: "0.1.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		cfg, err := commands.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("config load: %w", err)
		}
		cmd.SetContext(commands.WithConfig(cmd.Context(), cfg))
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand specified and DefaultToTUI is enabled, launch TUI
		cfg := commands.GetConfig(cmd.Context())
		if cfg.DefaultToTUI {
			return commands.TUICmd.RunE(cmd, args)
		}
		// Otherwise show help
		return cmd.Help()
	},
}

var configPath string

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion script",
	Long: `To load completions:

Bash:

  $ source <(troveler completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ troveler completion bash > /etc/bash_completion.d/troveler
  # macOS:
  $ troveler completion bash > /usr/local/etc/bash_completion.d/troveler

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ troveler completion zsh > "${fpath[1]}/_troveler"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ troveler completion fish | source

  # To load completions for each session, execute once:
  $ troveler completion fish > ~/.config/fish/completions/troveler.fish`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			RootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			RootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			RootCmd.GenFishCompletion(os.Stdout, true)
		}
	},
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	RootCmd.AddCommand(commands.TUICmd)
	RootCmd.AddCommand(commands.UpdateCmd)
	RootCmd.AddCommand(commands.SearchCmd)
	RootCmd.AddCommand(commands.InfoCmd)
	RootCmd.AddCommand(commands.InstallCmd)
	RootCmd.AddCommand(completionCmd)
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

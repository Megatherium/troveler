package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"troveler/commands"
)

var rootCmd = &cobra.Command{
	Use:   "troveler",
	Short: "Local terminaltrove.com mirror",
	Long: `Troveler is a CLI tool that creates a local searchable copy of terminaltrove.com.

Use 'troveler update' to fetch all tools from terminaltrove.com.
Use 'troveler search <query>' to search your local database.
Use 'troveler info <slug>' to see details of a specific tool.
Use 'troveler install <slug>' to get install commands for your OS.`,
	Version: "0.1.0",
}

func init() {
	rootCmd.AddCommand(commands.UpdateCmd)
	rootCmd.AddCommand(commands.SearchCmd)
	rootCmd.AddCommand(commands.InfoCmd)
	rootCmd.AddCommand(commands.InstallCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

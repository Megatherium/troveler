package commands

import (
	"context"

	"github.com/spf13/cobra"

	"troveler/db"
	"troveler/tui"
)

// TUICmd is the cobra command for launching the Terminal User Interface.
var TUICmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the Terminal User Interface",
	Long: `Launch the interactive Terminal User Interface for browsing and installing tools.

The TUI provides a rich interface with:
- Live search filtering
- Tool browsing with keyboard navigation
- Detailed tool information
- Easy install command execution
- Database update with progress animation

Use Alt+Q to quit, ? for help.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
			cfg := GetConfig(ctx)

			return tui.Run(database, cfg)
		})
	},
}

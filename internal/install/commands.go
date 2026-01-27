package install

import (
	"fmt"

	"troveler/db"
)

// CommandInfo represents an install command with metadata
type CommandInfo struct {
	Platform  string
	Command   string
	IsDefault bool // True if this is the auto-selected command
}

// FormatCommands prepares install commands for display
func FormatCommands(installs []db.InstallInstruction, defaultCmd *db.InstallInstruction) []CommandInfo {
	var commands []CommandInfo

	for _, inst := range installs {
		isDefault := defaultCmd != nil && inst.ID == defaultCmd.ID
		commands = append(commands, CommandInfo{
			Platform:  inst.Platform,
			Command:   inst.Command,
			IsDefault: isDefault,
		})
	}

	return commands
}

// RenderCommand formats a single command for CLI display
func RenderCommand(platform, command string) string {
	return fmt.Sprintf("%s: %s", platform, command)
}

// Package install provides helpers for selecting and formatting install commands.
package install

import (
	"fmt"

	"troveler/db"
	"troveler/internal/platform"
)

// CommandInfo pairs a platform command with a default flag.
type CommandInfo struct {
	Platform  string
	Command   string
	IsDefault bool
}

// FormatCommands converts raw install instructions into CommandInfo slices.
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

// RenderCommand formats a platform:command pair for display.
func RenderCommand(platformID, command string) string {
	return fmt.Sprintf("%s: %s", platformID, command)
}

// Selector is a re-export of platform.Selector.
type Selector = platform.Selector

// NewPlatformSelector creates a platform selector with the given overrides.
func NewPlatformSelector(cliOverride, configOverride, fallback, toolLanguage string) *Selector {
	return platform.NewSelector(cliOverride, configOverride, fallback, toolLanguage)
}

// FilterCommands filters install instructions by platform and language.
func FilterCommands(
	installs []db.InstallInstruction, platformID string, toolLanguage string,
) ([]db.InstallInstruction, bool) {
	return platform.FilterDBInstalls(installs, platformID, toolLanguage)
}

// SelectDefaultCommand picks the best default command from the filtered set.
func SelectDefaultCommand(
	commands []db.InstallInstruction, usedFallback bool, detectedOS string,
) *db.InstallInstruction {
	return platform.SelectDefaultDBInstalls(commands, usedFallback, detectedOS)
}

// TransformToMise rewrites a command into mise use format.
func TransformToMise(command string) string {
	return platform.TransformToMise(command)
}

// VirtualInstall represents a synthetic install instruction.
type VirtualInstall struct {
	Platform string
	Command  string
}

// GenerateVirtualInstallInstructions creates virtual install entries.
func GenerateVirtualInstallInstructions(installs []db.InstallInstruction) []VirtualInstall {
	virtuals := platform.GenerateVirtualInstalls(installs)
	result := make([]VirtualInstall, len(virtuals))
	for i, v := range virtuals {
		result[i] = VirtualInstall{Platform: v.Platform, Command: v.Command}
	}

	return result
}

package install

import (
	"fmt"

	"troveler/db"
	"troveler/internal/platform"
)

type CommandInfo struct {
	Platform  string
	Command   string
	IsDefault bool
}

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

func RenderCommand(platformID, command string) string {
	return fmt.Sprintf("%s: %s", platformID, command)
}

type Selector = platform.Selector

func NewPlatformSelector(cliOverride, configOverride, fallback, toolLanguage string) *Selector {
	return platform.NewSelector(cliOverride, configOverride, fallback, toolLanguage)
}

func FilterCommands(
	installs []db.InstallInstruction, platformID string, toolLanguage string,
) ([]db.InstallInstruction, bool) {
	return platform.FilterDBInstalls(installs, platformID, toolLanguage)
}

func SelectDefaultCommand(
	commands []db.InstallInstruction, usedFallback bool, detectedOS string,
) *db.InstallInstruction {
	return platform.SelectDefaultDBInstalls(commands, usedFallback, detectedOS)
}

func TransformToMise(command string) string {
	return platform.TransformToMise(command)
}

type VirtualInstall struct {
	Platform string
	Command  string
}

func GenerateVirtualInstallInstructions(installs []db.InstallInstruction) []VirtualInstall {
	virtuals := platform.GenerateVirtualInstalls(installs)
	result := make([]VirtualInstall, len(virtuals))
	for i, v := range virtuals {
		result[i] = VirtualInstall{Platform: v.Platform, Command: v.Command}
	}

	return result
}

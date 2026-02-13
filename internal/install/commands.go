// Package install provides install command handling and platform detection.
package install

import (
	"fmt"
	"strings"

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

// BackendType represents a supported mise backend
type BackendType string

const (
	BackendGo     BackendType = "go"
	BackendCargo  BackendType = "cargo"
	BackendNPM    BackendType = "npm"
	BackendPipx   BackendType = "pipx"
	BackendGithub BackendType = "github"
)

// VirtualInstall represents a generated install method
type VirtualInstall struct {
	Platform string
	Command  string
}

// GenerateVirtualInstallInstructions creates virtual mise install methods from raw install commands
// It identifies commands that can be transformed to mise format and generates virtual entries
// with the platform prefixed with "mise:" (e.g., mise:go, mise:cargo)
func GenerateVirtualInstallInstructions(installs []db.InstallInstruction) []VirtualInstall {
	virtuals := make(map[BackendType]VirtualInstall)

	for _, inst := range installs {
		// Skip if already a mise platform
		if strings.HasPrefix(inst.Platform, "mise:") {
			continue
		}

		transformed := TransformToMise(inst.Command)

		// Only generate virtual if transformation actually changed the command
		if transformed == inst.Command {
			continue
		}

		// Extract backend type from transformed command
		backend := extractBackendType(transformed)
		if backend == "" {
			continue
		}

		// Use the transformed command and create virtual platform name
		virtuals[backend] = VirtualInstall{
			Platform: "mise:" + string(backend),
			Command:  transformed,
		}
	}

	// Convert map to sorted slice for deterministic output
	result := make([]VirtualInstall, 0, len(virtuals))
	order := []BackendType{BackendGo, BackendCargo, BackendNPM, BackendPipx, BackendGithub}
	for _, backend := range order {
		if v, exists := virtuals[backend]; exists {
			result = append(result, v)
		}
	}

	return result
}

// extractBackendType extracts the backend type from a mise command
// e.g., "mise use --global go:github.com/user/repo" -> "go"
func extractBackendType(miseCommand string) BackendType {
	if !strings.HasPrefix(miseCommand, "mise use --global ") {
		return ""
	}

	rest := strings.TrimPrefix(miseCommand, "mise use --global ")

	// Extract backend:package pattern
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) < 2 {
		return ""
	}

	backend := parts[0]

	switch BackendType(backend) {
	case BackendGo, BackendCargo, BackendNPM, BackendPipx, BackendGithub:
		return BackendType(backend)
	default:
		return ""
	}
}

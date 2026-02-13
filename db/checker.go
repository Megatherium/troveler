package db

import (
	"os/exec"
	"regexp"
	"strings"
)

// IsInstalled checks if a tool is installed by examining its install instructions
// and checking if executable name is available on PATH
func IsInstalled(tool *Tool, installs []InstallInstruction) bool {
	if tool == nil || len(installs) == 0 {
		return false
	}

	for _, inst := range installs {
		if cmd := parseToolName(inst.Command); cmd != "" {
			if isCommandAvailable(cmd) {
				return true
			}
		}
	}

	return false
}

// parseToolName extracts executable/package name from an install command
func parseToolName(command string) string {
	command = strings.TrimSpace(command)

	// Common patterns for install commands
	patterns := []struct {
		pattern string
		group   int
	}{
		// npm install <package>
		{`npm\s+(install|i|global)\s+([^\s]+)`, 2},
		// yarn global add <package>
		{`yarn\s+global\s+add\s+([^\s]+)`, 1},
		// pnpm add -g <package>
		{`pnpm\s+add\s+-g\s+([^\s]+)`, 1},
		// pip install <package>
		{`pip(?:3)?\s+install\s+([^\s]+)`, 1},
		// pipx install <package>
		{`pipx\s+install\s+([^\s]+)`, 1},
		// uv tool install <package>
		{`uv\s+tool\s+install\s+([^\s]+)`, 1},
		// cargo install <crate>
		{`cargo\s+install\s+([^\s]+)`, 1},
		// go install <package>
		{`go\s+install\s+([^\s]+)`, 1},
		// brew install <formula>
		{`(?:brew|linuxbrew)\s+install\s+([^\s]+)`, 1},
		// apt install <package>
		{`apt(?:-get)?\s+install\s+([^\s]+)`, 1},
		// pacman -S <package>
		{`pacman\s+-S\s+([^\s]+)`, 1},
		// dnf/yum install <package>
		{`(?:dnf|yum)\s+install\s+([^\s]+)`, 1},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		matches := re.FindStringSubmatch(command)
		if len(matches) > p.group {
			name := matches[p.group]
			// Strip version specifiers like package@1.2.3
			if idx := strings.Index(name, "@"); idx > 0 {
				name = name[:idx]
			}
			// Extract last component from paths (e.g., github.com/user/repo -> repo, org/formulae/name -> name)
			if strings.Contains(name, "/") {
				parts := strings.Split(name, "/")
				if len(parts) > 0 {
					name = parts[len(parts)-1]
				}
			}

			return name
		}
	}

	// If no pattern matched, try to extract from the command
	// Handle cases like "mise use --global npm:package"
	if strings.Contains(command, "mise use") {
		parts := strings.Fields(command)
		for _, part := range parts {
			if strings.Contains(part, ":") {
				nameParts := strings.Split(part, ":")
				if len(nameParts) > 1 {
					return nameParts[1]
				}
			}
		}
	}

	// For direct executable names
	fields := strings.Fields(command)
	if len(fields) > 0 {
		// Return last field, usually package/executable name
		name := fields[len(fields)-1]
		// Strip version specifiers like package@1.2.3
		if idx := strings.Index(name, "@"); idx > 0 {
			name = name[:idx]
		}
		// Extract last component from paths (e.g., owner/repo -> repo)
		if strings.Contains(name, "/") {
			parts := strings.Split(name, "/")
			if len(parts) > 0 {
				name = parts[len(parts)-1]
			}
		}

		return name
	}

	return ""
}

// isCommandAvailable checks if a command is available on PATH
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)

	return err == nil
}

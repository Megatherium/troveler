// Package db provides SQLite-backed storage for tools and install instructions.
package db

import (
	"os/exec"
	"regexp"
	"strings"
)

// IsInstalled checks if a tool is installed by examining its install instructions
// and checking if executable name is available on PATH.
// If an install instruction has ExecutableName set, that is used directly;
// otherwise the executable name is parsed from the command string.
func IsInstalled(tool *Tool, installs []InstallInstruction) bool {
	if tool == nil || len(installs) == 0 {
		return false
	}

	for _, inst := range installs {
		name := resolveExecutableName(inst)
		if name != "" && isCommandAvailable(name) {
			return true
		}
	}

	return false
}

// resolveExecutableName returns the executable name for an install instruction.
// It prefers the explicit ExecutableName field; falls back to parsing the command.
func resolveExecutableName(inst InstallInstruction) string {
	if inst.ExecutableName != "" {
		return inst.ExecutableName
	}

	return parseToolName(inst.Command)
}

// isCommandAvailable checks if a command is available on PATH.
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)

	return err == nil
}

// resolver defines how to extract an executable name from an install command.
type resolver struct {
	pattern *regexp.Regexp
	extract func(matches []string) string
}

// resolvers is the ordered list of command patterns and their extractors.
// Order matters: more specific patterns should come before general ones.
var resolvers = []resolver{
	// --- Sudo-prefixed commands (strip sudo, delegate to inner command) ---
	{regexp.MustCompile(`(?i)^sudo\s+port\s+install\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^sudo\s+ports\s+install\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^sudo\s+snap\s+install\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^sudo\s+dnf\s+install\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^sudo\s+(?:apt|apt-get)\s+install\s+(.+)$`), firstField},

	// --- Go install ---
	// go install host/repo/cmd/...@latest (wildcard - extract repo name)
	{regexp.MustCompile(`(?i)^go\s+install\s+(?:https?://)?[^/]+/([^/@\s]+)/cmd/\.\.\.(?:@\S+)?$`), identity},
	// go install host/user/repo/cmd/...@latest (wildcard with org)
	{regexp.MustCompile(`(?i)^go\s+install\s+(?:https?://)?[^/]+/[^/]+/([^/@\s]+)/cmd/\.\.\.(?:@\S+)?$`), identity},
	// go install github.com/user/repo/cmd/binary@latest
	{regexp.MustCompile(`(?i)^go\s+install\s+\S+/cmd/(\S+?)(?:@\S+)?$`), stripVersion},
	// go install github.com/user/repo/v2/cmd/binary@latest
	{regexp.MustCompile(`(?i)^go\s+install\s+\S+/v\d+/cmd/(\S+?)(?:@\S+)?$`), stripVersion},
	// go install github.com/user/repo@latest (last path segment)
	{regexp.MustCompile(`(?i)^go\s+install\s+(?:https?://)?\S+/(\S+?)(?:@\S+)?$`), stripVersion},

	// --- Cargo ---
	{regexp.MustCompile(`(?i)^cargo\s+binstall\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^cargo\s+install\s+--locked\s+--git\s+\S+\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^cargo\s+install\s+--git\s+\S+\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^cargo\s+install\s+--git=\S+\s+(.+)$`), firstField},
	// cargo install --git <url> (no explicit binary name - extract from URL)
	{regexp.MustCompile(`(?i)^cargo\s+install\s+(?:--locked\s+)?--git\s+https?://[^/]+/[^/]+/(\S+?)(?:\.git)?$`), identity},
	{regexp.MustCompile(`(?i)^cargo\s+install\s+(?:--locked\s+)?(.+)$`), firstField},

	// --- npm / yarn / pnpm ---
	// npm scoped: @scope/package -> extract last segment after /
	{regexp.MustCompile(`(?i)^npm\s+(?:install\s+--global|install\s+-g|i\s+-g|i\s+--global|install|i)\s+@([^/@\s]+/([^\s]+))`), npmScoped},
	// npm unscoped
	{regexp.MustCompile(`(?i)^npm\s+(?:install\s+--global|install\s+-g|i\s+-g|i\s+--global|install|i)\s+([^\s]+)`), stripVersion},
	// npm install daff -g (flag at end)
	{regexp.MustCompile(`(?i)^npm\s+install\s+([^\s]+)\s+-g`), stripVersion},
	{regexp.MustCompile(`(?i)^yarn\s+global\s+add\s+([^\s]+)`), stripVersion},
	{regexp.MustCompile(`(?i)^pnpm\s+(?:add|install|i)\s+-g\s+([^\s]+)`), stripVersion},

	// --- Python ---
	{regexp.MustCompile(`(?i)^(?:python\d*\s+-m\s+)?pip(?:3)?\s+install\s+([^\s]+)`), stripVersion},
	{regexp.MustCompile(`(?i)^pipx\s+install\s+([^\s]+)`), stripVersion},
	// uv tool install [flags...] <package> - extract last argument
	{regexp.MustCompile(`(?i)^uv\s+tool\s+install\s+(.+)`), lastField},
	{regexp.MustCompile(`(?i)^uvx\s+([^\s]+)`), stripVersion},
	{regexp.MustCompile(`(?i)^rye\s+install\s+([^\s]+)`), stripVersion},

	// --- Homebrew / Linuxbrew ---
	{regexp.MustCompile(`(?i)^(?:brew|linuxbrew)\s+install\s+(.+)$`), firstField},

	// --- Debian / Ubuntu ---
	{regexp.MustCompile(`(?i)^(?:apt|apt-get)\s+install\s+(.+)$`), firstField},

	// --- Arch family ---
	{regexp.MustCompile(`(?i)^pacman\s+-S\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^yay\s+-S\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^paru\s+-S\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^paru\s+-Syu\s+(.+)$`), firstField},

	// --- Fedora / RHEL ---
	{regexp.MustCompile(`(?i)^(?:dnf|yum)\s+install\s+(.+)$`), firstField},

	// --- openSUSE ---
	{regexp.MustCompile(`(?i)^zypper\s+install\s+(.+)$`), firstField},

	// --- Alpine ---
	{regexp.MustCompile(`(?i)^apk\s+add\s+(.+)$`), firstField},

	// --- Gentoo ---
	{regexp.MustCompile(`(?i)^emerge\s+(?:--\S+\s+)*(\S+)$`), stripVersion},

	// --- FreeBSD ---
	{regexp.MustCompile(`(?i)^pkg_add\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^pkgin\s+install\s+(.+)$`), firstField},

	// --- macOS MacPorts ---
	{regexp.MustCompile(`(?i)^port\s+install\s+(.+)$`), firstField},

	// --- Nix ---
	{regexp.MustCompile(`(?i)^nix-env\s+-iA?\s+(?:nixos\.|nixpkgs\.)?(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^nix-env\s+-i\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^nix-shell\s+-p\s+(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^nix\s+profile\s+install\s+(?:\S+\s+)*(.+)$`), firstField},

	// --- Windows ---
	{regexp.MustCompile(`(?i)^scoop\s+install\s+(?:\S+/)?(.+)$`), firstField},
	{regexp.MustCompile(`(?i)^choco\s+install\s+(.+?)(?:\s+--\S+)*$`), firstField},
	{regexp.MustCompile(`(?i)^winget\s+install\s+(.+?)(?:\s+--\S+)*$`), firstField},

	// --- Snap ---
	{regexp.MustCompile(`(?i)^snap\s+install\s+(.+)$`), firstField},

	// --- eget (github release downloader) ---
	{regexp.MustCompile(`(?i)^eget\s+(\S+)/(\S+)$`), egetExtract},

	// --- mise ---
	{regexp.MustCompile(`(?i)^mise\s+(?:use|install)\s+(?:--global\s+)?\S+:(\S+)$`), firstField},

	// --- Gem (Ruby) ---
	{regexp.MustCompile(`(?i)^gem\s+install\s+(.+)$`), firstField},

	// --- Cabal (Haskell) ---
	{regexp.MustCompile(`(?i)^cabal\s+(?:install|update)\s+(.+)$`), firstField},

	// --- npx ---
	{regexp.MustCompile(`(?i)^npx\s+([^\s]+)`), stripVersion},

	// --- pkg (Termux / various) ---
	{regexp.MustCompile(`(?i)^pkg\s+install\s+(.+)$`), firstField},

	// --- pkgman (Haiku) ---
	{regexp.MustCompile(`(?i)^pkgman\s+install\s+(.+)$`), firstField},

	// --- curl/wget install scripts (extract tool name from domain) ---
	{regexp.MustCompile(`(?i)^curl\s+-\S*\s+https?://([^./\s]+)\.`), identity},
	{regexp.MustCompile(`(?i)^wget\s+https?://[^/]+/[^\s]*?/([^/\s.]+)`), identity},
}

// parseToolName extracts executable/package name from an install command.
func parseToolName(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}

	for _, r := range resolvers {
		matches := r.pattern.FindStringSubmatch(command)
		if len(matches) >= 2 {
			name := r.extract(matches)
			if name != "" {
				return name
			}
		}
	}

	// Fallback: try "mise use" special case (colon-separated names)
	if strings.Contains(command, "mise use") {
		return extractMisUse(command)
	}

	// Last resort: take the last non-flag field
	return lastNonFlagField(command)
}

// --- Extractor helpers ---

func identity(matches []string) string { return matches[1] }

func firstField(matches []string) string {
	// The captured group may contain multiple space-separated args; take the first.
	field := strings.Fields(matches[1])
	if len(field) == 0 {
		return ""
	}

	return cleanName(field[0])
}

func stripVersion(matches []string) string {
	return cleanName(matches[1])
}

// lastField returns the last space-separated field from the captured group.
func lastField(matches []string) string {
	fields := strings.Fields(matches[1])
	if len(fields) == 0 {
		return ""
	}

	return cleanName(fields[len(fields)-1])
}

func npmScoped(matches []string) string {
	// matches[2] is the inner group after the /, e.g. "cli" from @ast-grep/cli
	if len(matches) >= 3 && matches[2] != "" {
		return cleanName(matches[2])
	}

	// Fallback: full scoped name minus @
	return cleanName(strings.TrimPrefix(matches[1], "@"))
}

func egetExtract(matches []string) string {
	// eget owner/repo -> repo is the executable
	if len(matches) >= 3 {
		return cleanName(matches[2])
	}

	return cleanName(matches[1])
}

// cleanName strips version specifiers and extracts the last path component.
func cleanName(name string) string {
	// Strip version specifiers like package@1.2.3 or package@latest
	if idx := strings.Index(name, "@"); idx > 0 {
		name = name[:idx]
	}

	// Extract last component from paths (github.com/user/repo -> repo)
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]
	}

	return name
}

// extractMisUse handles "mise use --global npm:package" patterns.
func extractMisUse(command string) string {
	parts := strings.Fields(command)
	for _, part := range parts {
		if strings.Contains(part, ":") {
			nameParts := strings.Split(part, ":")
			if len(nameParts) > 1 {
				return nameParts[1]
			}
		}
	}

	return ""
}

// lastNonFlagField returns the last command argument that doesn't start with -.
func lastNonFlagField(command string) string {
	fields := strings.Fields(command)
	for i := len(fields) - 1; i >= 0; i-- {
		if !strings.HasPrefix(fields[i], "-") && !strings.HasPrefix(fields[i], "@") {
			return cleanName(fields[i])
		}
	}

	return ""
}

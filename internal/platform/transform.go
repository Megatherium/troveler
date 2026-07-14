// Package platform provides install command transformation functionality.
package platform

import (
	"regexp"
	"strings"
)

// cargoGitEqRegex matches "--git=<url>" (optionally followed by a binary name).
var cargoGitEqRegex = regexp.MustCompile(`^--git=(\S+)`)

// cargoGitSpaceRegex matches "--git <url>" (optionally followed by a binary name).
var cargoGitSpaceRegex = regexp.MustCompile(`^--git\s+(\S+)`)

// stripCargoFlags removes known cargo verification flags (--locked, --frozen)
// that may precede the crate name or git URL.
func stripCargoFlags(s string) string {
	for {
		if rest, ok := strings.CutPrefix(s, "--locked "); ok {
			s = rest
			continue
		}
		if rest, ok := strings.CutPrefix(s, "--frozen "); ok {
			s = rest
			continue
		}
		return s
	}
}

// TransformToMise converts install commands to mise format.
// Supports transforming:
//   - "go install <package>" → "mise use --global go:<package>"
//   - "cargo install [--locked|--frozen] <crate>" → "mise use --global cargo:<crate>"
//   - "cargo install [--locked|--frozen] --git[=]<url> [<binary>]" → "mise use --global cargo:<url>"
//   - "npm/yarn/pnpm install <package>" → "mise use --global npm:<package>"
//   - "pip/pip3/pipx install <package>" → "mise use --global pipx:<package>"
//   - "eget <repo>" → "mise use --global github:<repo>"
//
// Returns the original command if no transformation applies.
func TransformToMise(command string) string {
	command = strings.TrimSpace(command)

	if after, ok := strings.CutPrefix(command, "go install "); ok {
		pkg := after
		pkg = strings.TrimPrefix(pkg, "https://")

		return "mise use --global go:" + pkg
	}

	if after, ok := strings.CutPrefix(command, "cargo install "); ok {
		stripped := stripCargoFlags(after)

		if matches := cargoGitEqRegex.FindStringSubmatch(stripped); matches != nil {
			return "mise use --global cargo:" + matches[1]
		}
		if matches := cargoGitSpaceRegex.FindStringSubmatch(stripped); matches != nil {
			return "mise use --global cargo:" + matches[1]
		}
		if fields := strings.Fields(stripped); len(fields) > 0 {
			if strings.HasPrefix(fields[0], "--") {
				return command
			}
			return "mise use --global cargo:" + fields[0]
		}
		return command
	}

	npmRegex := regexp.MustCompile(`^(npm install -g|yarn global add|pnpm add -g)\s+(.+)$`)
	if matches := npmRegex.FindStringSubmatch(command); len(matches) > 2 {
		pkg := matches[2]

		return "mise use --global npm:" + pkg
	}

	if strings.HasPrefix(command, "npm install ") && !strings.Contains(command, "-g") {
		pkg := strings.TrimPrefix(command, "npm install ")

		return "mise use --global npm:" + pkg
	}

	pipRegex := regexp.MustCompile(`^(pip install|pip3 install|pipx install)\s+(.+)$`)
	if matches := pipRegex.FindStringSubmatch(command); len(matches) > 2 {
		pkg := matches[2]

		return "mise use --global pipx:" + pkg
	}

	if after, ok := strings.CutPrefix(command, "eget "); ok {
		fields := strings.Fields(after)
		if len(fields) > 0 {
			repo := fields[len(fields)-1]
			repo = strings.TrimPrefix(repo, "github.com/")

			return "mise use --global github:" + repo
		}
	}

	return command
}

// Package platform provides install command transformation functionality.
package platform

import (
	"regexp"
	"strings"
)

// TransformToMise converts install commands to mise format.
// Supports transforming:
//   - "go install <package>" → "mise use --global go:<package>"
//   - "cargo install <crate>" → "mise use --global cargo:<crate>"
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
		crate := after

		return "mise use --global cargo:" + crate
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

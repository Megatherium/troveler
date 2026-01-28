package install

import (
	"regexp"
	"strings"
)

// TransformToMise converts standard install commands to mise-en-place commands
func TransformToMise(command string) string {
	command = strings.TrimSpace(command)

	if strings.HasPrefix(command, "go install ") {
		pkg := strings.TrimPrefix(command, "go install ")
		pkg = strings.TrimPrefix(pkg, "https://")
		return "mise use --global go:" + pkg
	}

	if strings.HasPrefix(command, "cargo install ") {
		crate := strings.TrimPrefix(command, "cargo install ")
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

	return command
}

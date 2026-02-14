package platform

import (
	"strings"
)

var Aliases = map[string]string{
	"pip":   "python (pip)",
	"pipx":  "python (pipx)",
	"uv":    "python (uv)",
	"npm":   "node (npm)",
	"yarn":  "node (yarn)",
	"pnpm":  "node (pnpm)",
	"bun":   "node (bun)",
	"cargo": "rust (cargo)",
}

func Normalize(platform string) string {
	if normalized, ok := Aliases[strings.ToLower(platform)]; ok {
		return normalized
	}

	return platform
}

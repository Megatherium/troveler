// Package platform provides platform ID normalization functionality.
package platform

import (
	"strings"
)

// Aliases maps common shorthand platform names to their canonical forms.
// For example, "pip" normalizes to "python (pip)" for consistent matching.
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

// Normalize converts a platform identifier to its canonical form.
// If the platform has a defined alias, returns the aliased form; otherwise returns the original.
func Normalize(platform string) string {
	if normalized, ok := Aliases[strings.ToLower(platform)]; ok {
		return normalized
	}

	return platform
}

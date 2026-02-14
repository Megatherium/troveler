// Package platform provides platform matching and language detection functionality.
package platform

import (
	"strings"
)

// LanguageToPackageManager maps programming languages to their associated package managers.
// Used for matching install instructions to tools based on their language.
var LanguageToPackageManager = map[string][]string{
	"python":      {"python", "pip", "pipx", "uv"},
	"rust":        {"rust", "cargo"},
	"javascript":  {"javascript", "npm", "yarn", "pnpm", "bun", "deno", "node"},
	"typescript":  {"typescript", "npm", "yarn", "pnpm", "bun", "deno", "node"},
	"go":          {"go"},
	"ruby":        {"ruby", "gem"},
	"perl":        {"perl", "cpan"},
	"haskell":     {"haskell", "cabal", "stack"},
	"csharp":      {"csharp", "dotnet", "nuget"},
	"nim":         {"nim", "nimble"},
	"ocaml":       {"ocaml", "opam"},
	"zig":         {"zig"},
	"common-lisp": {"common-lisp", "quicklisp"},
	"haxe":        {"haxe", "haxelib"},
}

// MatchPlatform determines if an install platform matches the detected OS ID.
// Supports various platform formats including:
//   - Simple OS IDs (e.g., "macos", "ubuntu")
//   - OS:method formats (e.g., "linux:apt", "macos:brew")
//   - Multi-distro methods (e.g., "linux:ubuntu / debian")
//   - Language-based platforms
func MatchPlatform(detectedID string, installPlatform string) bool {
	parts := strings.SplitN(installPlatform, ":", 2)
	platformOS := parts[0]
	platformMethod := ""
	if len(parts) > 1 {
		platformMethod = parts[1]
	}

	if platformOS == installPlatform {
		return platformOS == detectedID
	}

	if platformOS == detectedID {
		return true
	}

	if platformOS == "linux" && platformMethod != "" {
		if platformMethod == detectedID {
			return true
		}
		methods := strings.Split(platformMethod, "/")
		for _, m := range methods {
			m = strings.TrimSpace(m)
			if m == detectedID {
				return true
			}
		}
	}

	if platformOS == "bsd" && platformMethod != "" {
		if platformMethod == detectedID {
			return true
		}
	}

	if detectedID == OSRHEL {
		return platformOS == OSRHEL || platformOS == OSFedora || platformOS == OSCentos
	}
	if detectedID == OSArch {
		return platformOS == OSArch || platformOS == OSManjaro
	}
	if detectedID == OSUbuntu {
		return platformOS == OSUbuntu || platformOS == OSDebian
	}

	return false
}

// MatchLanguage checks if an install platform matches a given programming language.
// Returns true if the platform is the language itself or uses a package manager
// associated with that language (e.g., "python (pip)" matches "python").
func MatchLanguage(language string, installPlatform string) bool {
	language = strings.ToLower(language)
	installPlatform = strings.ToLower(installPlatform)

	if installPlatform == language {
		return true
	}

	if strings.HasPrefix(installPlatform, language+" ") || strings.HasPrefix(installPlatform, language+"(") {
		return true
	}

	if managers, ok := LanguageToPackageManager[language]; ok {
		for _, manager := range managers {
			if installPlatform == manager {
				return true
			}
			if strings.HasPrefix(installPlatform, manager+" ") || strings.HasPrefix(installPlatform, manager+"(") {
				return true
			}
		}
	}

	return false
}

// IsLinuxFamily returns true if the given OS ID belongs to the Linux family.
func IsLinuxFamily(osID string) bool {
	switch osID {
	case OSUbuntu, OSDebian, OSFedora, OSArch, OSManjaro, OSRHEL, OSCentos:
		return true
	default:
		return strings.Contains(strings.ToLower(osID), "linux")
	}
}

// IsMacFamily returns true if the given OS ID belongs to the macOS family (darwin/macos).
func IsMacFamily(osID string) bool {
	osID = strings.ToLower(osID)

	return strings.Contains(osID, OSMacOS) || strings.Contains(osID, OSDarwin)
}

// IsBSDFamily returns true if the given OS ID belongs to the BSD family.
func IsBSDFamily(osID string) bool {
	switch osID {
	case OSFreeBSD, OSOpenBSD, OSNetBSD:
		return true
	default:
		return strings.Contains(strings.ToLower(osID), "bsd")
	}
}

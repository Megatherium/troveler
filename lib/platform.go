package lib

import (
	"strings"
)

var PlatformAliases = map[string]string{
	"pip":   "python (pip)",
	"pipx":  "python (pipx)",
	"uv":    "python (uv)",
	"npm":   "node (npm)",
	"yarn":  "node (yarn)",
	"pnpm":  "node (pnpm)",
	"bun":   "node (bun)",
	"cargo": "rust (cargo)",
}

func NormalizePlatform(platform string) string {
	if normalized, ok := PlatformAliases[strings.ToLower(platform)]; ok {
		return normalized
	}
	return platform
}

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

	if detectedID == "rhel" {
		return platformOS == "rhel" || platformOS == "fedora" || platformOS == "centos"
	}

	if detectedID == "arch" {
		return platformOS == "arch" || platformOS == "manjaro"
	}

	if detectedID == "ubuntu" {
		return platformOS == "ubuntu" || platformOS == "debian"
	}

	return false
}

func MatchLanguage(language string, installPlatform string) bool {
	// Exact match (e.g., "rust" == "rust", "python" == "python")
	if installPlatform == language {
		return true
	}
	// Prefix match (e.g., "python (pip)", "rust (cargo)")
	return strings.HasPrefix(installPlatform, language+" ") || strings.HasPrefix(installPlatform, language+"(")
}

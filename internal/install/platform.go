package install

import (
	"strings"
	
	"troveler/db"
	"troveler/lib"
)

// PlatformSelector handles platform selection logic based on priority
type PlatformSelector struct {
	cliOverride    string
	configOverride string
	fallback       string
	toolLanguage   string
}

// NewPlatformSelector creates a new platform selector
func NewPlatformSelector(cliOverride, configOverride, fallback, toolLanguage string) *PlatformSelector {
	return &PlatformSelector{
		cliOverride:    cliOverride,
		configOverride: configOverride,
		fallback:       fallback,
		toolLanguage:   toolLanguage,
	}
}

// SelectPlatform determines the platform to use based on priority:
// CLI override > config override > OS detection > fallback
func (ps *PlatformSelector) SelectPlatform(detectedOS string) string {
	// Check CLI override first
	if ps.cliOverride != "" {
		return ps.cliOverride
	}

	// Check config override
	if ps.configOverride != "" {
		return ps.configOverride
	}

	// Use OS detection if available
	if detectedOS != "" {
		return detectedOS
	}

	// Fall back to configured fallback
	return ps.fallback
}

// FilterCommands filters install commands based on platform selection
// Returns (matched commands, whether fallback was used)
func FilterCommands(installs []db.InstallInstruction, platform string, toolLanguage string) ([]db.InstallInstruction, bool) {
	var matched []db.InstallInstruction

	if platform == "LANG" {
		// Use language matching
		for _, inst := range installs {
			if lib.MatchLanguage(toolLanguage, inst.Platform) {
				matched = append(matched, inst)
			}
		}
	} else {
		// Use platform matching
		normalizedPlatform := lib.NormalizePlatform(platform)
		for _, inst := range installs {
			if lib.MatchPlatform(normalizedPlatform, inst.Platform) {
				matched = append(matched, inst)
			}
		}
	}

	// Fallback: if no matches found, return all instructions
	// This ensures users always see *something* rather than empty panel
	if len(matched) == 0 {
		return installs, true // true = fallback used
	}

	return matched, false // false = normal match
}

// SelectDefaultCommand returns the best default command
// If fallback was used, tries to pick a sensible default (brew, apt, etc.)
// Otherwise returns the first matched command
func SelectDefaultCommand(commands []db.InstallInstruction, usedFallback bool) *db.InstallInstruction {
	if len(commands) == 0 {
		return nil
	}
	
	// If we used fallback (showing all commands), try to pick a sensible default
	if usedFallback {
		// Priority order for fallback defaults
		preferredPlatforms := []string{"brew", "linux:brew", "macos:brew", "apt", "apt-get", "linux:ubuntu", "linux:debian", "pacman", "linux:arch"}
		for _, preferred := range preferredPlatforms {
			for _, cmd := range commands {
				if strings.Contains(strings.ToLower(cmd.Platform), preferred) {
					return &cmd
				}
			}
		}
		// If no preferred found, don't mark any as default
		return nil
	}
	
	// Normal case: first match is default
	return &commands[0]
}

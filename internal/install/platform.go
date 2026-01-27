package install

import (
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
func FilterCommands(installs []db.InstallInstruction, platform string, toolLanguage string) []db.InstallInstruction {
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

	return matched
}

// SelectDefaultCommand returns the first matched command (highest priority)
func SelectDefaultCommand(commands []db.InstallInstruction) *db.InstallInstruction {
	if len(commands) == 0 {
		return nil
	}
	return &commands[0]
}

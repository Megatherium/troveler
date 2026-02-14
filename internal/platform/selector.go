// Package platform provides platform selection logic for determining which install method to use.
package platform

// Selector handles the priority-based selection of the target platform.
// Selection order: CLI override > config override > detected OS > fallback.
type Selector struct {
	cliOverride    string
	configOverride string
	fallback       string
	toolLanguage   string
}

// NewSelector creates a new Selector with the given configuration options.
// Parameters:
//   - cliOverride: platform specified via CLI flags (highest priority)
//   - configOverride: platform specified in config file
//   - fallback: default platform to use when no OS detected
//   - toolLanguage: programming language of the tool being installed
func NewSelector(cliOverride, configOverride, fallback, toolLanguage string) *Selector {
	return &Selector{
		cliOverride:    cliOverride,
		configOverride: configOverride,
		fallback:       fallback,
		toolLanguage:   toolLanguage,
	}
}

// Select returns the platform ID to use based on the selector's configuration.
// Selection priority: CLI override > config override > detected OS > fallback.
func (s *Selector) Select(detectedOS string) string {
	if s.cliOverride != "" {
		return s.cliOverride
	}

	if s.configOverride != "" {
		return s.configOverride
	}

	if detectedOS != "" {
		return detectedOS
	}

	return s.fallback
}

// SelectPlatform returns the platform ID to use. Alias for Select method.
func (s *Selector) SelectPlatform(detectedOS string) string {
	return s.Select(detectedOS)
}

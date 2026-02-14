package platform

type Selector struct {
	cliOverride    string
	configOverride string
	fallback       string
	toolLanguage   string
}

func NewSelector(cliOverride, configOverride, fallback, toolLanguage string) *Selector {
	return &Selector{
		cliOverride:    cliOverride,
		configOverride: configOverride,
		fallback:       fallback,
		toolLanguage:   toolLanguage,
	}
}

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

func (s *Selector) SelectPlatform(detectedOS string) string {
	return s.Select(detectedOS)
}

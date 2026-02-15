package tui

import (
	"troveler/db"
)

// BatchInstallConfig holds configuration for batch tool installation.
type BatchInstallConfig struct {
	ReuseConfig    bool   // Use same config for all tools
	UseSudo        bool   // Use sudo for all installs
	SudoOnlySystem bool   // Use sudo only for system package managers
	SkipIfBlind    bool   // Skip tools without proper install method
	UseMise        bool   // Use mise for all installs
	SudoPassword   string // Cached sudo password (in memory only)
	ConfigStep     int    // Current step in config wizard (0-4)
}

// BatchInstallProgress tracks the progress of a batch installation.
type BatchInstallProgress struct {
	Tools         []db.SearchResult
	CurrentIndex  int
	Completed     []string // Tool IDs that completed successfully
	Failed        []string // Tool IDs that failed
	Skipped       []string // Tool IDs that were skipped
	CurrentOutput string
	CurrentError  error
	IsComplete    bool
}

// NewBatchInstallConfig creates a new batch install configuration.
func NewBatchInstallConfig() *BatchInstallConfig {
	return &BatchInstallConfig{
		ReuseConfig:    true,
		UseSudo:        false,
		SudoOnlySystem: true,
		SkipIfBlind:    false,
		UseMise:        false,
		ConfigStep:     0,
	}
}

// NewBatchInstallProgress creates a new batch install progress tracker.
func NewBatchInstallProgress(tools []db.SearchResult) *BatchInstallProgress {
	return &BatchInstallProgress{
		Tools:        tools,
		CurrentIndex: 0,
		Completed:    []string{},
		Failed:       []string{},
		Skipped:      []string{},
		IsComplete:   false,
	}
}

// ConfigStepCount returns the total number of config steps.
func (c *BatchInstallConfig) ConfigStepCount() int {
	return 5 // reuse, sudo, sudo-only-system, skip-if-blind, mise
}

// NextStep advances to the next config step.
func (c *BatchInstallConfig) NextStep() bool {
	c.ConfigStep++

	return c.ConfigStep < c.ConfigStepCount()
}

// PrevStep goes back to the previous config step.
func (c *BatchInstallConfig) PrevStep() bool {
	if c.ConfigStep > 0 {
		c.ConfigStep--

		return true
	}

	return false
}

// IsConfigComplete returns true if configuration is complete.
func (c *BatchInstallConfig) IsConfigComplete() bool {
	return c.ConfigStep >= c.ConfigStepCount()
}

// GetCurrentStepTitle returns the title for the current config step.
func (c *BatchInstallConfig) GetCurrentStepTitle() string {
	switch c.ConfigStep {
	case 0:
		return "Reuse configuration for all tools?"
	case 1:
		return "Use sudo for installation?"
	case 2:
		return "Sudo only for system package managers?"
	case 3:
		return "Skip tools without install method?"
	case 4:
		return "Use mise for installation?"
	default:
		return ""
	}
}

// GetCurrentStepOptions returns the options for the current config step.
func (c *BatchInstallConfig) GetCurrentStepOptions() []string {
	switch c.ConfigStep {
	case 0:
		return []string{"Yes - use same config for all", "No - configure per tool"}
	case 1:
		return []string{"Yes - always use sudo", "No - never use sudo"}
	case 2:
		return []string{"Yes - sudo only for apt/dnf/etc", "No - apply sudo setting to all"}
	case 3:
		return []string{"Yes - skip unknown tools", "No - show available options"}
	case 4:
		return []string{"Yes - prefer mise", "No - use native commands"}
	default:
		return []string{}
	}
}

// SetCurrentStepValue sets the value for the current config step.
func (c *BatchInstallConfig) SetCurrentStepValue(optionIndex int) {
	switch c.ConfigStep {
	case 0:
		c.ReuseConfig = optionIndex == 0
	case 1:
		c.UseSudo = optionIndex == 0
	case 2:
		c.SudoOnlySystem = optionIndex == 0
	case 3:
		c.SkipIfBlind = optionIndex == 0
	case 4:
		c.UseMise = optionIndex == 0
	}
}

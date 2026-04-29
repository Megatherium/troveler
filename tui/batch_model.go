package tui

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/install"
	"troveler/internal/platform"
)

// --- Message types for batch install flow -----------------------------------

type batchInstallStartMsg struct {
	tools []db.SearchResult
}

type batchInstallProgressMsg struct {
	toolID  string
	output  string
	err     error
	skipped bool
}

type batchInstallCompleteMsg struct{}

// --- BatchInstallModel ------------------------------------------------------

// BatchInstallModel manages batch install configuration and progress.
type BatchInstallModel struct {
	db       *db.SQLiteDB
	config   *BatchInstallConfig
	progress *BatchInstallProgress
}

// NewBatchInstallModel creates a new BatchInstallModel.
func NewBatchInstallModel(database *db.SQLiteDB) *BatchInstallModel {
	return &BatchInstallModel{db: database}
}

// Config returns the current batch install configuration.
func (bm *BatchInstallModel) Config() *BatchInstallConfig { return bm.config }

// Progress returns the current batch install progress.
func (bm *BatchInstallModel) Progress() *BatchInstallProgress { return bm.progress }

// IsConfigActive reports whether the batch config wizard is open.
func (bm *BatchInstallModel) IsConfigActive() bool { return bm.config != nil }

// IsRunning reports whether a batch install is in progress.
func (bm *BatchInstallModel) IsRunning() bool { return bm.progress != nil }

// StartBatchConfig creates a new batch config and begins the wizard.
func (bm *BatchInstallModel) StartBatchConfig(useMise bool) {
	bm.config = NewBatchInstallConfig()
	if useMise {
		bm.config.UseMise = true
	}
}

// ClearConfig discards the current batch config.
func (bm *BatchInstallModel) ClearConfig() {
	bm.config = nil
}

// SetStep sets the wizard to the given step directly (for tests).
func (bm *BatchInstallModel) SetStep(n int) {
	if bm.config != nil {
		bm.config.ConfigStep = n
	}
}

// SetStepValue records a wizard option choice for the current step.
func (bm *BatchInstallModel) SetStepValue(optionIndex int) {
	if bm.config != nil {
		bm.config.SetCurrentStepValue(optionIndex)
	}
}

// AdvanceStep moves the wizard to the next step. Returns false when the wizard
// is complete (all steps done).
func (bm *BatchInstallModel) AdvanceStep() bool {
	if bm.config == nil {
		return false
	}

	return bm.config.NextStep()
}

// StartInstall begins batch installation of the marked tools.
func (bm *BatchInstallModel) StartInstall(markedTools []db.SearchResult) tea.Cmd {
	bm.progress = NewBatchInstallProgress(markedTools)

	return func() tea.Msg {
		return batchInstallStartMsg{tools: markedTools}
	}
}

// ProcessTool processes the batch tool at the given index, returning a Cmd
// that runs the install command in a background goroutine.
func (bm *BatchInstallModel) ProcessTool(index int) tea.Cmd {
	if bm.progress == nil || index >= len(bm.progress.Tools) {
		return func() tea.Msg { return batchInstallCompleteMsg{} }
	}

	tool := bm.progress.Tools[index]
	cfg := bm.config
	database := bm.db

	return func() tea.Msg {
		installs, err := database.GetInstallInstructions(tool.ID)
		if err != nil || len(installs) == 0 {
			if cfg != nil && cfg.SkipIfBlind {
				return batchInstallProgressMsg{
					toolID:  tool.ID,
					skipped: true,
				}
			}

			return batchInstallProgressMsg{
				toolID: tool.ID,
				err:    fmt.Errorf("no install instructions found"),
			}
		}

		selector := install.NewPlatformSelector("", "", "", tool.Language)
		osInfo, _ := platform.DetectOS()
		detectedOS := ""
		if osInfo != nil {
			detectedOS = osInfo.ID
		}

		result := install.ResolvePlatform(selector, installs, detectedOS, tool.Language)
		filtered := result.Installs
		if result.UsedFallback || len(filtered) == 0 {
			synthMatched, platformID := install.TryResolveLangFallback(installs, result.PlatformID)
			if len(synthMatched) > 0 {
				filtered = synthMatched
			} else if result.UsedFallback && len(filtered) > 0 {
				// Fallback returned candidate installs from other platforms;
				// use them as-is since TryResolveLangFallback couldn't refine.
				_ = platformID
			} else {
				if cfg != nil && cfg.SkipIfBlind {
					return batchInstallProgressMsg{
						toolID:  tool.ID,
						skipped: true,
					}
				}

				return batchInstallProgressMsg{
					toolID: tool.ID,
					err:    fmt.Errorf("no compatible install method"),
				}
			}
		}

		defaultCmd := install.SelectDefaultCommand(filtered, result.UsedFallback, detectedOS)
		cmd := filtered[0].Command
		if defaultCmd != nil {
			cmd = defaultCmd.Command
		}

		if cfg != nil && cfg.UseMise {
			cmd = install.TransformToMise(cmd)
		}

		if cfg != nil && cfg.UseSudo {
			isSystemPM := isSystemPackageManager(filtered[0].Platform)
			if !cfg.SudoOnlySystem || isSystemPM {
				cmd = "sudo " + cmd
			}
		}

		execCmd := exec.Command("sh", "-c", cmd) //nolint:noctx,gosec
		output, err := execCmd.CombinedOutput()

		return batchInstallProgressMsg{
			toolID: tool.ID,
			output: string(output),
			err:    err,
		}
	}
}

// HandleProgress applies a progress update and reports whether the batch is
// finished. When not finished, nextIndex gives the next tool to process.
func (bm *BatchInstallModel) HandleProgress(msg batchInstallProgressMsg) (finished bool, nextIndex int) {
	if bm.progress == nil {
		return true, 0
	}

	if msg.skipped {
		bm.progress.Skipped = append(bm.progress.Skipped, msg.toolID)
	} else if msg.err != nil {
		bm.progress.Failed = append(bm.progress.Failed, msg.toolID)
	} else {
		bm.progress.Completed = append(bm.progress.Completed, msg.toolID)
	}
	bm.progress.CurrentOutput = msg.output
	bm.progress.CurrentError = msg.err
	bm.progress.CurrentIndex++

	if bm.progress.CurrentIndex < len(bm.progress.Tools) {
		return false, bm.progress.CurrentIndex
	}
	bm.progress.IsComplete = true

	return true, 0
}

// HandleComplete marks the batch install as finished.
func (bm *BatchInstallModel) HandleComplete() {
	if bm.progress != nil {
		bm.progress.IsComplete = true
	}
}

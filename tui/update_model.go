package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/update"
)

// UpdateModel manages database update state and the slug wave animation.
type UpdateModel struct {
	service  *update.Service
	running  bool
	slugWave *update.SlugWave
	progress chan update.ProgressUpdate
	cancel   context.CancelFunc
}

// NewUpdateModel creates a new UpdateModel.
func NewUpdateModel(database *db.SQLiteDB) *UpdateModel {
	return &UpdateModel{
		service: update.NewService(database),
	}
}

// IsRunning reports whether an update is in progress.
func (um *UpdateModel) IsRunning() bool {
	return um.running
}

// SlugWave returns the slug wave animation (nil if not started).
func (um *UpdateModel) SlugWave() *update.SlugWave {
	return um.slugWave
}

// Start begins the database update process.
func (um *UpdateModel) Start() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	um.cancel = cancel

	opts := update.Options{
		Limit:    0,
		Progress: um.progress,
	}

	go func() {
		_ = um.service.FetchAndUpdate(ctx, opts)
	}()

	um.running = true
	return um.listen()
}

// listen waits for progress updates and returns a command.
func (um *UpdateModel) listen() tea.Cmd {
	return func() tea.Msg {
		if um.progress == nil {
			return nil
		}
		upd, ok := <-um.progress
		if !ok {
			return updateProgressMsg{Type: "complete"}
		}

		return updateProgressMsg(upd)
	}
}

// tick returns a command that ticks the animation frame.
func (um *UpdateModel) tick() tea.Cmd {
	return tea.Tick(time.Millisecond*33, func(_ time.Time) tea.Msg {
		return slugTickMsg{}
	})
}

// HandleProgress processes a progress update message.
func (um *UpdateModel) HandleProgress(upd update.ProgressUpdate) {
	if upd.Type == "progress" && upd.Total > 0 && um.slugWave == nil {
		um.slugWave = update.NewSlugWave(upd.Total)
	}

	if upd.Type == "slug" && um.slugWave != nil {
		um.slugWave.AddSlug(upd.Slug)
		um.slugWave.IncProcessed()
	}

	if upd.Type == "complete" || upd.Type == "error" {
		um.running = false
	}
}

// HandleTick advances the slug wave animation frame.
func (um *UpdateModel) HandleTick() tea.Cmd {
	if um.running && um.slugWave != nil {
		um.slugWave.AdvanceFrame()
		return um.tick()
	}
	return nil
}

// Close stops the update and cleans up resources.
func (um *UpdateModel) Close() {
	if um.cancel != nil {
		um.cancel()
		um.cancel = nil
	}
	um.running = false
	um.progress = nil
}

// StartProgress prepares the progress channel and returns start commands.
func (um *UpdateModel) StartProgress() tea.Cmd {
	um.progress = make(chan update.ProgressUpdate, 100)
	return tea.Batch(um.Start(), um.tick())
}
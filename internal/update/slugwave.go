package update

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/lipgloss"

	"troveler/pkg/ui"
)

const (
	streamWidth  = 60
	streamHeight = 4
)

var runePalette = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~!@#$%^&*()_+-=")

// SlugEntry represents a slug in the wave
type SlugEntry struct {
	Slug     string
	Position int
	Age      int
	Row      int
}

// SlugWave manages the animated slug wave display
type SlugWave struct {
	slugBuffer []SlugEntry
	bufferMu   sync.Mutex
	step       int
	totalTools int
	processed  int64
	startTime  time.Time
}

// NewSlugWave creates a new slug wave animation
func NewSlugWave(totalTools int) *SlugWave {
	return &SlugWave{
		slugBuffer: make([]SlugEntry, 0, 50),
		totalTools: totalTools,
		startTime:  time.Now(),
	}
}

// AddSlug adds a slug to the wave
func (sw *SlugWave) AddSlug(slug string) {
	sw.bufferMu.Lock()
	entry := SlugEntry{
		Slug:     slug,
		Position: streamWidth - 1 - (len(sw.slugBuffer) % 15),
		Age:      0,
		Row:      rand.IntN(streamHeight),
	}
	sw.slugBuffer = append(sw.slugBuffer, entry)
	if len(sw.slugBuffer) > 30 {
		sw.slugBuffer = sw.slugBuffer[1:]
	}
	sw.bufferMu.Unlock()
}

// IncProcessed increments the processed count
func (sw *SlugWave) IncProcessed() {
	atomic.AddInt64(&sw.processed, 1)
}

// AdvanceFrame advances the animation frame
func (sw *SlugWave) AdvanceFrame() {
	sw.bufferMu.Lock()
	for i := range sw.slugBuffer {
		sw.slugBuffer[i].Age++
	}
	sw.step++
	sw.bufferMu.Unlock()
}

// RenderStream renders the chaotic slug stream
func (sw *SlugWave) RenderStream() string {
	sw.bufferMu.Lock()
	defer sw.bufferMu.Unlock()

	var lines []string
	for row := 0; row < streamHeight; row++ {
		var line strings.Builder
		for col := 0; col < streamWidth; col++ {
			found := false
			for _, entry := range sw.slugBuffer {
				pos := (entry.Position - entry.Age + streamWidth) % streamWidth
				dist := abs(pos - col)
				if dist < 4 && entry.Row == row {
					charIdx := (col + entry.Age + sw.step) % len(entry.Slug)
					char := entry.Slug[charIdx]
					colorIdx := (entry.Age + col + sw.step) % len(ui.GradientColors)
					color := ui.GradientColors[colorIdx]

					style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
					if sw.step%8 < 4 {
						style = style.Bold(true)
					}

					line.WriteString(style.Render(string(char)))
					found = true

					break
				}
			}
			if !found {
				noiseChar := runePalette[(col+sw.step+row*7)%len(runePalette)]
				color := ui.GetGradientColor(col+sw.step, streamWidth*2)
				line.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(color)).
					Render(string(noiseChar)))
			}
		}
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}

// RenderWithProgress renders the stream with progress info
func (sw *SlugWave) RenderWithProgress() string {
	processed := atomic.LoadInt64(&sw.processed)
	percent := float64(processed) / float64(sw.totalTools)
	if percent > 1 {
		percent = 1
	}

	elapsed := time.Since(sw.startTime)
	var eta time.Duration
	if processed > 0 {
		eta = time.Duration(float64(elapsed) / float64(processed) * float64(sw.totalTools-int(processed)))
	}

	status := fmt.Sprintf("%d/%d (%.0f%%)", processed, sw.totalTools, percent*100)
	etaStr := fmt.Sprintf("ETA: %s", eta.Round(time.Second))

	stream := sw.RenderStream()

	return fmt.Sprintf("%s\n\n%s\n\n%s",
		lipgloss.NewStyle().Bold(true).Render(status),
		stream,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(etaStr),
	)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

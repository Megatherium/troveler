package commands

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func TestUpdateUIConcurrentSlugUpdatesAndRendering(t *testing.T) {
	const workers, slugsPerWorker = 4, 200
	u := NewUpdateUI(workers * slugsPerWorker)
	start := make(chan struct{})
	var tasks sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		tasks.Add(1)
		go func() {
			defer tasks.Done()
			<-start
			for slug := 0; slug < slugsPerWorker; slug++ {
				u.AddSlug(fmt.Sprintf("worker-%d-tool-%d", worker, slug))
				u.IncProcessed()
				runtime.Gosched()
			}
		}()
	}
	for renderer := 0; renderer < 2; renderer++ {
		tasks.Add(1)
		go func() {
			defer tasks.Done()
			<-start
			for frame := 0; frame < 20; frame++ {
				assertUpdateFrameShape(t, u.Render(), getTerminalWidth())
				runtime.Gosched()
			}
		}()
	}
	close(start)
	tasks.Wait()

	if got := atomic.LoadInt64(&u.processed); got != workers*slugsPerWorker {
		t.Errorf("processed=%d; want %d", got, workers*slugsPerWorker)
	}
	if final := ansi.Strip(u.Render()); !strings.Contains(final, "800/800 (100%)") {
		t.Errorf("final progress is incorrect: %q", final)
	}
	u.bufferMu.Lock()
	defer u.bufferMu.Unlock()
	if len(u.slugBuffer) != 30 {
		t.Errorf("buffer retained %d slugs; want its bounded capacity of 30", len(u.slugBuffer))
	}
}

func TestUpdateUIConcurrentFrameAdvancementAndRendering(t *testing.T) {
	u := NewUpdateUI(1)
	u.AddSlug("frame-fixture")
	ctx, cancel := context.WithCancel(context.Background())
	exited := make(chan struct{})
	go func() {
		defer close(exited)
		runUpdateUI(ctx, u, make(chan struct{}), true)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-exited:
		case <-time.After(time.Second):
			t.Error("animation loop did not stop after cancellation")
		}
	})

	deadline := time.Now().Add(2 * time.Second)
	for {
		assertUpdateFrameShape(t, u.Render(), getTerminalWidth())
		u.bufferMu.Lock()
		step, age := u.step, u.slugBuffer[0].age
		u.bufferMu.Unlock()
		if age != step {
			t.Fatalf("frame=%d and slug age=%d did not advance together", step, age)
		}
		if step >= 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("animation ticker did not advance two frames")
		}
		runtime.Gosched()
	}
}

func TestUpdateUIStreamRenderingPreservesAnimation(t *testing.T) {
	u := NewUpdateUI(2)
	// A known row/offset makes the visible slug fragment deterministic.
	u.slugBuffer = []slugEntry{{slug: "XYZ", startOffset: -1, row: 1}}
	initial := strings.Split(ansi.Strip(u.renderChaoticStream(12)), "\n")
	if len(initial) != streamHeight || initial[0] != "abcdefghijkl" || initial[1][8:] != "ZXYZ" {
		t.Fatalf("unexpected initial noise or slug placement: %q", initial)
	}
	u.bufferMu.Lock()
	u.step = 1
	u.slugBuffer[0].age = 1
	u.bufferMu.Unlock()
	advanced := strings.Split(ansi.Strip(u.renderChaoticStream(12)), "\n")
	if len(advanced) != streamHeight || advanced[0] != "bcdefghijklm" || advanced[1][7:] != "XYZXY" {
		t.Fatalf("frame advancement changed slug/noise progression: %q", advanced)
	}
	u.IncProcessed()
	if display := ansi.Strip(u.Render()); !strings.Contains(display, "1/2 (50%)") {
		t.Errorf("intermediate progress is incorrect: %q", display)
	}
}

func assertUpdateFrameShape(t *testing.T, display string, width int) {
	t.Helper()
	lines := strings.Split(ansi.Strip(display), "\n")
	if len(lines) != streamHeight+1 {
		t.Errorf("frame has %d lines; want progress plus %d stream rows", len(lines), streamHeight)
		return
	}
	for row, line := range lines[1:] {
		if got := ansi.StringWidth(line); got != width {
			t.Errorf("stream row %d has width %d; want %d", row, got, width)
		}
	}
}

package commands

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"troveler/crawler"
	"troveler/db"
)

var limit int
var logOutput bool

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Crawl terminaltrove.com and update local database",
	Long: `Fetches all tools from terminaltrove.com and stores them in the local SQLite database.
Use --log to show detailed logging output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig(cmd.Context())
		if cfg == nil {
			return fmt.Errorf("config not loaded")
		}

		database, err := db.New(cfg.DSN)
		if err != nil {
			return fmt.Errorf("db init: %w", err)
		}
		defer database.Close()

		currentCount, err := database.ToolCount(context.Background())
		if err != nil {
			return fmt.Errorf("count tools: %w", err)
		}

		fetcher := crawler.NewFetcher()

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Minute)
		defer cancel()

		return runUpdate(ctx, database, fetcher, limit, logOutput, currentCount)
	},
}

func init() {
	UpdateCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit number of tools to fetch (0 for all)")
	UpdateCmd.Flags().BoolVarP(&logOutput, "log", "v", false, "Show verbose logging output")
}

const (
	progressBarWidth = 50
	streamWidth      = 60
	streamHeight     = 4
	totalLines       = 1 + streamHeight + 1
)

var runePalette = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~!@#$%^&*()_+-=")

var gradientColors = []string{
	"#90EE90", "#8BE88C", "#86E288", "#81DC88", "#7CD688",
	"#77D088", "#72CA88", "#6DC488", "#68BE88", "#63B888",
	"#5EB288", "#59AC88", "#54A688", "#4FA088", "#4A9A88",
	"#459488", "#408E88", "#3B8888", "#368288", "#317C88",
	"#2C7688", "#277088", "#226A88", "#1D6488", "#185E88",
	"#135888", "#0E5288", "#094C88", "#044688", "#004088",
}

func getGradientColor(pos int, total int) string {
	if total == 0 {
		return gradientColors[0]
	}
	p := float64(pos) / float64(total)
	if p > 1 {
		p = 1
	}
	idx := int(p * float64(len(gradientColors)-1))
	return gradientColors[idx]
}

type slugEntry struct {
	slug     string
	position int
	age      int
	row      int
}

type UpdateUI struct {
	totalTools int
	processed  int64
	slugBuffer []slugEntry
	bufferMu   sync.Mutex
	startTime  time.Time
	step       int
}

func NewUpdateUI(totalTools int) *UpdateUI {
	return &UpdateUI{
		totalTools: totalTools,
		slugBuffer: make([]slugEntry, 0, 50),
		startTime:  time.Now(),
	}
}

func (u *UpdateUI) AddSlug(slug string) {
	u.bufferMu.Lock()
	entry := slugEntry{
		slug:     slug,
		position: streamWidth - 1 - (len(u.slugBuffer) % 15),
		age:      0,
		row:      rand.IntN(streamHeight),
	}
	u.slugBuffer = append(u.slugBuffer, entry)
	if len(u.slugBuffer) > 30 {
		u.slugBuffer = u.slugBuffer[1:]
	}
	u.bufferMu.Unlock()
}

func (u *UpdateUI) IncProcessed() {
	atomic.AddInt64(&u.processed, 1)
}

func (u *UpdateUI) Render() string {
	processed := atomic.LoadInt64(&u.processed)
	percent := float64(processed) / float64(u.totalTools)
	if percent > 1 {
		percent = 1
	}

	filled := int(percent * float64(progressBarWidth))

	var bar strings.Builder
	for i := 0; i < filled; i++ {
		color := getGradientColor(i, filled)
		bar.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("▇"))
	}
	for i := filled; i < progressBarWidth; i++ {
		bar.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("░"))
	}

	elapsed := time.Since(u.startTime)
	var eta time.Duration
	if processed > 0 {
		eta = time.Duration(float64(elapsed) / float64(processed) * float64(u.totalTools-int(processed)))
	}

	status := fmt.Sprintf(" %d/%d (%.0f%%)", processed, u.totalTools, percent*100)
	etaStr := fmt.Sprintf(" ETA: %s", eta.Round(time.Second))

	stream := u.renderChaoticStream()

	return fmt.Sprintf("\x1b7\x1b[1;1H\x1b[J%s %s %s  %s\n%s\x1b8",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("▐"),
		bar.String(),
		lipgloss.NewStyle().Bold(true).Render(status),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(etaStr),
		stream,
	)
}

func (u *UpdateUI) renderChaoticStream() string {
	var lines []string
	for row := 0; row < streamHeight; row++ {
		var line strings.Builder
		for col := 0; col < streamWidth; col++ {
			found := false
			for _, entry := range u.slugBuffer {
				pos := (entry.position - entry.age + streamWidth) % streamWidth
				dist := abs(pos - col)
				if dist < 4 && entry.row == row {
					charIdx := (col + entry.age + u.step) % len(entry.slug)
					char := entry.slug[charIdx]
					colorIdx := (entry.age + col + u.step) % len(gradientColors)
					color := gradientColors[colorIdx]

					style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
					if u.step%8 < 4 {
						style = style.Bold(true)
					}

					line.WriteString(style.Render(string(char)))
					found = true
					break
				}
			}
			if !found {
				noiseChar := runePalette[(col+u.step+row*7)%len(runePalette)]
				color := getGradientColor(col+u.step, streamWidth*2)
				line.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color(color)).
					Render(string(noiseChar)))
			}
		}
		lines = append(lines, line.String())
	}
	return strings.Join(lines, "\n")
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func runUpdate(ctx context.Context, database *db.SQLiteDB, fetcher *crawler.Fetcher, limit int, logOutput bool, currentCount int) error {
	if !logOutput {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render("Fetching tools from terminaltrove.com..."))
		fmt.Println()
	}

	initialData, err := fetcher.FetchSearchPage(ctx, 1)
	if err != nil {
		return fmt.Errorf("initial fetch: %w", err)
	}

	initialResp, err := crawler.ParseSearchResponse(initialData)
	if err != nil {
		return fmt.Errorf("parse initial: %w", err)
	}

	totalTools := int(initialResp.Found)
	if limit > 0 && limit < totalTools {
		totalTools = limit
	}

	if logOutput {
		fmt.Printf("Found %d tools (limited to %d)\n", int(initialResp.Found), totalTools)
	} else {
		fmt.Printf("Found %d tools\n", totalTools)
	}

	pageResults, err := fetcher.FetchSearchPagesConcurrently(ctx, (totalTools+99)/100)
	if err != nil {
		if logOutput {
			fmt.Printf("Warning: Partial page fetch error: %v\n", err)
		}
	}

	allSlugs := make([]string, 0, totalTools)

	for i := 1; i <= (totalTools+99)/100; i++ {
		data, ok := pageResults[i]
		if !ok {
			continue
		}

		resp, err := crawler.ParseSearchResponse(data)
		if err != nil {
			continue
		}

		for _, item := range resp.Hits {
			if item.Document.Slug != "" {
				allSlugs = append(allSlugs, item.Document.Slug)
				if limit > 0 && len(allSlugs) >= limit {
					break
				}
			}
		}
		if limit > 0 && len(allSlugs) >= limit {
			break
		}
	}

	if len(allSlugs) == 0 {
		if logOutput {
			fmt.Println("No tools found in search pages")
		} else {
			fmt.Println("No tools found.")
		}
		return nil
	}

	if logOutput {
		fmt.Printf("Fetching details for %d tools...\n", len(allSlugs))
	}

	ui := NewUpdateUI(len(allSlugs))

	slugChan := make(chan string, len(allSlugs))
	for _, s := range allSlugs {
		slugChan <- s
	}
	close(slugChan)

	detailChan := make(chan crawler.DetailPage, 100)

	workerCount := 5
	var detailWg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		detailWg.Add(1)
		go func() {
			defer detailWg.Done()
			for slug := range slugChan {
				data, err := fetcher.FetchDetailPage(ctx, slug)
				if err != nil {
					continue
				}

				detail, err := crawler.ParseDetailPage(data)
				if err != nil {
					continue
				}

				ui.IncProcessed()
				ui.AddSlug(slug)

				select {
				case detailChan <- *detail:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		detailWg.Wait()
		close(detailChan)
	}()

	writeChan := make(chan db.InstallInstruction, 200)
	var writeWg sync.WaitGroup

	dbWorkerCount := 3
	for i := 0; i < dbWorkerCount; i++ {
		writeWg.Add(1)
		go func() {
			defer writeWg.Done()
			for inst := range writeChan {
				database.UpsertInstallInstruction(ctx, &inst)
			}
		}()
	}

	ticker := time.NewTicker(33 * time.Millisecond)
	defer ticker.Stop()

	frame := 0
	for {
		select {
		case detail, ok := <-detailChan:
			if !ok {
				detailChan = nil
				break
			}

			database.UpsertTool(ctx, detail.ToTool())

			for _, inst := range detail.ToInstallInstructions() {
				select {
				case writeChan <- inst:
				case <-ctx.Done():
					break
				}
			}

		case <-ticker.C:
			frame++
			ui.bufferMu.Lock()
			for i := range ui.slugBuffer {
				ui.slugBuffer[i].age++
			}
			ui.step = frame
			ui.bufferMu.Unlock()

			if !logOutput {
				fmt.Print(ui.Render())
			}

		case <-ctx.Done():
			close(writeChan)
			writeWg.Wait()
			return ctx.Err()
		}

		if detailChan == nil {
			break
		}
	}

	close(writeChan)
	writeWg.Wait()

	if !logOutput {
		fmt.Println()
	}

	finalCount, _ := database.ToolCount(context.Background())

	if logOutput {
		fmt.Printf("Update complete! Database now has %d tools\n", finalCount)
	} else {
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).Render("✓ ") +
			fmt.Sprintf("Update complete! Database has %d tools (+%d new)",
				finalCount, finalCount-currentCount))
		fmt.Println()
	}

	return nil
}

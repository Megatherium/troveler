package commands

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"troveler/crawler"
	"troveler/db"
)

var limit int

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Crawl terminaltrove.com and update local database",
	Long:  "Fetches all tools from terminaltrove.com and stores them in the local SQLite database.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig(cmd.Context())
		if cfg == nil {
			return fmt.Errorf("config not loaded")
		}

		log.Info("Opening database with DSN: ", cfg.DSN)

		database, err := db.New(cfg.DSN)
		if err != nil {
			return fmt.Errorf("db init: %w", err)
		}
		defer database.Close()

		currentCount, err := database.ToolCount(context.Background())
		if err != nil {
			return fmt.Errorf("count tools: %w", err)
		}

		log.Info("Current tool count: ", currentCount)

		fetcher := crawler.NewFetcher()

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Minute)
		defer cancel()

		return runUpdate(ctx, database, fetcher, limit)
	},
}

func init() {
	UpdateCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit number of tools to fetch (0 for all)")
}

func runUpdate(ctx context.Context, database *db.SQLiteDB, fetcher *crawler.Fetcher, limit int) error {
	log.Info("Fetching search page 1 to determine total pages...")

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
	totalPages := (totalTools + 99) / 100

	log.Info(fmt.Sprintf("Found %d tools (limited to %d)", int(initialResp.Found), totalTools))

	pageResults, err := fetcher.FetchSearchPagesConcurrently(ctx, totalPages)
	if err != nil {
		log.Warn(fmt.Sprintf("Partial page fetch error: %v", err))
	}

	allSlugs := make([]string, 0, totalTools)

	for i := 1; i <= totalPages; i++ {
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
		log.Warn("No tools found in search pages")
		return nil
	}

	log.Info(fmt.Sprintf("Fetching details for %d tools...", len(allSlugs)))

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

	processed := 0
	lastPercent := 0

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

	progressTicker := time.NewTicker(2 * time.Second)
	defer progressTicker.Stop()

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

			processed++
			if len(allSlugs) > 0 {
				percent := (processed * 100) / len(allSlugs)
				if percent >= lastPercent+10 {
					log.Info(fmt.Sprintf("Progress: %d/%d (%d%%)", processed, len(allSlugs), percent))
					lastPercent = percent
				}
			}

		case <-progressTicker.C:
			if detailChan != nil && len(allSlugs) > 0 {
				percent := (processed * 100) / len(allSlugs)
				fmt.Fprintf(os.Stderr, "\rProgress: %d/%d (%d%%)", processed, len(allSlugs), percent)
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

	fmt.Fprintln(os.Stderr, "")

	finalCount, _ := database.ToolCount(context.Background())

	log.Info(fmt.Sprintf("Update complete! Database now has %d tools", finalCount))

	return nil
}

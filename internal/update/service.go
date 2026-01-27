package update

import (
	"context"
	"fmt"
	"sync"

	"troveler/crawler"
	"troveler/db"
)

// Service handles database updates
type Service struct {
	db      *db.SQLiteDB
	fetcher *crawler.Fetcher
}

// NewService creates a new update service
func NewService(database *db.SQLiteDB) *Service {
	return &Service{
		db:      database,
		fetcher: crawler.NewFetcher(),
	}
}

// ProgressUpdate represents an update progress event
type ProgressUpdate struct {
	Type       string // "start", "slug", "progress", "complete", "error"
	Slug       string // Tool slug being processed
	Processed  int    // Number of tools processed
	Total      int    // Total number of tools
	Message    string // Status message
	Error      error  // Error if any
}

// UpdateOptions configures the update
type UpdateOptions struct {
	Limit    int                      // Limit number of tools (0 = all)
	Progress chan<- ProgressUpdate    // Channel for progress updates
}

// FetchAndUpdate fetches all tools and updates the database
func (s *Service) FetchAndUpdate(ctx context.Context, opts UpdateOptions) error {
	// Send start message
	if opts.Progress != nil {
		opts.Progress <- ProgressUpdate{Type: "start", Message: "Fetching tools from terminaltrove.com..."}
	}

	// Fetch all slugs
	slugs, total, err := s.fetchSlugs(ctx, opts.Limit)
	if err != nil {
		if opts.Progress != nil {
			opts.Progress <- ProgressUpdate{Type: "error", Error: err}
		}
		return err
	}

	if len(slugs) == 0 {
		if opts.Progress != nil {
			opts.Progress <- ProgressUpdate{Type: "complete", Message: "No tools found"}
		}
		return nil
	}

	if opts.Progress != nil {
		opts.Progress <- ProgressUpdate{
			Type:    "progress",
			Total:   len(slugs),
			Message: fmt.Sprintf("Found %d tools", len(slugs)),
		}
	}

	// Fetch and process details
	detailChan, errChan := s.fetchDetailsConcurrently(ctx, slugs, opts.Progress)

	// Process details and write to database
	processed := 0
	for detail := range detailChan {
		tool := detail.ToTool()
		if err := s.db.UpsertTool(ctx, tool); err != nil {
			continue
		}

		installs := detail.ToInstallInstructions()
		for _, inst := range installs {
			if err := s.db.UpsertInstallInstruction(ctx, &inst); err != nil {
				continue
			}
		}

		processed++
		if opts.Progress != nil {
			opts.Progress <- ProgressUpdate{
				Type:      "progress",
				Processed: processed,
				Total:     len(slugs),
			}
		}
	}

	// Check for errors
	select {
	case err := <-errChan:
		if err != nil && opts.Progress != nil {
			opts.Progress <- ProgressUpdate{Type: "error", Error: err}
		}
	default:
	}

	if opts.Progress != nil {
		opts.Progress <- ProgressUpdate{
			Type:      "complete",
			Processed: processed,
			Total:     total,
			Message:   fmt.Sprintf("Update complete! Saved %d tools.", processed),
		}
	}

	return nil
}

// fetchSlugs fetches all tool slugs from search pages
func (s *Service) fetchSlugs(ctx context.Context, limit int) ([]string, int, error) {
	initialData, err := s.fetcher.FetchSearchPage(ctx, 1)
	if err != nil {
		return nil, 0, fmt.Errorf("initial fetch: %w", err)
	}

	initialResp, err := crawler.ParseSearchResponse(initialData)
	if err != nil {
		return nil, 0, fmt.Errorf("parse initial: %w", err)
	}

	totalTools := int(initialResp.Found)
	if limit > 0 && limit < totalTools {
		totalTools = limit
	}

	pageResults, err := s.fetcher.FetchSearchPagesConcurrently(ctx, (totalTools+99)/100)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch pages: %w", err)
	}

	var allSlugs []string
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

	return allSlugs, int(initialResp.Found), nil
}

// fetchDetailsConcurrently fetches tool details concurrently
func (s *Service) fetchDetailsConcurrently(ctx context.Context, slugs []string, progress chan<- ProgressUpdate) (<-chan crawler.DetailPage, <-chan error) {
	detailChan := make(chan crawler.DetailPage, 100)
	errChan := make(chan error, 1)

	slugChan := make(chan string, len(slugs))
	for _, slug := range slugs {
		slugChan <- slug
	}
	close(slugChan)

	workerCount := 5
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for slug := range slugChan {
				data, err := s.fetcher.FetchDetailPage(ctx, slug)
				if err != nil {
					continue
				}

				detail, err := crawler.ParseDetailPage(data)
				if err != nil {
					continue
				}

				// Send slug progress update
				if progress != nil {
					progress <- ProgressUpdate{
						Type: "slug",
						Slug: slug,
					}
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
		wg.Wait()
		close(detailChan)
		close(errChan)
	}()

	return detailChan, errChan
}

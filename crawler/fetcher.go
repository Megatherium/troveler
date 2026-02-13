package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	baseURL        = "https://terminaltrove.com"
	resultsPerPage = 100
	maxRetries     = 3
	globalRate     = 20
	burstRate      = 40
)

type Fetcher struct {
	client  *http.Client
	limiter *rate.Limiter
	cache   map[string][]byte
	mu      sync.RWMutex
}

type FetchResult struct {
	URL  string
	Data []byte
	Err  error
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: rate.NewLimiter(rate.Limit(globalRate), burstRate),
		cache:   make(map[string][]byte),
	}
}

func (f *Fetcher) FetchSearchPage(ctx context.Context, page int) ([]byte, error) {
	url := fmt.Sprintf("%s/search?q=*&page=%d&per_page=%d", baseURL, page, resultsPerPage)

	return f.Fetch(ctx, url)
}

func (f *Fetcher) FetchDetailPage(ctx context.Context, slug string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/", baseURL, slug)

	return f.Fetch(ctx, url)
}

func (f *Fetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	f.mu.RLock()
	if data, ok := f.cache[url]; ok {
		f.mu.RUnlock()

		return data, nil
	}
	f.mu.RUnlock()

	if err := f.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	var body []byte
	var fetchErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

		resp, err := f.client.Do(req)
		if err != nil {
			fetchErr = err
			if attempt < maxRetries {
				delay := time.Duration(attempt*attempt) * 100 * time.Millisecond
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
					continue
				}
			}

			return nil, fmt.Errorf("fetch failed after %d attempts: %w", maxRetries, fetchErr)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			if attempt < maxRetries {
				delay := time.Duration(attempt*attempt) * 100 * time.Millisecond
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
					continue
				}
			}

			return nil, fmt.Errorf("status %d", resp.StatusCode)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			if attempt < maxRetries {
				delay := time.Duration(attempt*attempt) * 100 * time.Millisecond
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
					continue
				}
			}

			return nil, fmt.Errorf("read failed: %w", err)
		}

		break
	}

	f.mu.Lock()
	f.cache[url] = body
	f.mu.Unlock()

	return body, nil
}

func (f *Fetcher) FetchSearchPagesConcurrently(ctx context.Context, totalPages int) (map[int][]byte, error) {
	results := make(map[int][]byte)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, totalPages)

	pageChan := make(chan int, totalPages)

	for i := 1; i <= totalPages; i++ {
		pageChan <- i
	}
	close(pageChan)

	workerCount := 5
	for range workerCount {
		wg.Go(func() {
			for page := range pageChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				data, err := f.FetchSearchPage(ctx, page)
				if err != nil {
					errChan <- fmt.Errorf("page %d: %w", page, err)

					continue
				}

				mu.Lock()
				results[page] = data
				mu.Unlock()
			}
		})
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("fetch errors: %v", errors)
	}

	return results, nil
}

func (f *Fetcher) FetchDetailsConcurrently(ctx context.Context, slugs []string) (map[string][]byte, error) {
	results := make(map[string][]byte)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(slugs))

	slugChan := make(chan string, len(slugs))
	for _, s := range slugs {
		slugChan <- s
	}
	close(slugChan)

	workerCount := 5
	for range workerCount {
		wg.Go(func() {
			for slug := range slugChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				data, err := f.FetchDetailPage(ctx, slug)
				if err != nil {
					errChan <- fmt.Errorf("slug %s: %w", slug, err)

					continue
				}

				mu.Lock()
				results[slug] = data
				mu.Unlock()
			}
		})
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("detail fetch errors: %v", errors)
	}

	return results, nil
}

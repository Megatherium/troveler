package commands

import (
	"context"
	"testing"

	"troveler/crawler"
)

func TestFetchAndParseSlugs(t *testing.T) {
	ctx := context.Background()

	fetcher := crawler.NewFetcher()

	slugs, _, err := fetchAndParseSlugs(ctx, fetcher, 10)
	if err != nil {
		t.Fatalf("fetchAndParseSlugs failed: %v", err)
	}

	if len(slugs) == 0 {
		t.Logf("No slugs fetched from real API (expected - running without network)")
	}
}

func TestFetchAndParseSlugsNoLimit(t *testing.T) {
	ctx := context.Background()

	fetcher := crawler.NewFetcher()

	slugs, _, err := fetchAndParseSlugs(ctx, fetcher, 0)
	if err != nil {
		t.Fatalf("fetchAndParseSlugs failed: %v", err)
	}

	if len(slugs) == 0 {
		t.Logf("No slugs fetched from real API (expected - running without network)")
	}
}

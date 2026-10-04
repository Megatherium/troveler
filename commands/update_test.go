package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type searchFixtureFetcher struct {
	t           *testing.T
	ctx         context.Context
	initial     []byte
	pages       map[int][]byte
	initialErr  error
	pagesErr    error
	initialPage []int
	pageCounts  []int
}

func (f *searchFixtureFetcher) FetchSearchPage(ctx context.Context, page int) ([]byte, error) {
	f.t.Helper()
	if ctx != f.ctx {
		f.t.Fatal("initial fetch did not receive the caller's context")
	}
	f.initialPage = append(f.initialPage, page)

	return f.initial, f.initialErr
}

func (f *searchFixtureFetcher) FetchSearchPagesConcurrently(
	ctx context.Context, count int,
) (map[int][]byte, error) {
	f.t.Helper()
	if ctx != f.ctx {
		f.t.Fatal("page fetch did not receive the caller's context")
	}
	f.pageCounts = append(f.pageCounts, count)
	pages := make(map[int][]byte)
	for page := 1; page <= count; page++ {
		if data, ok := f.pages[page]; ok {
			pages[page] = data
		}
	}

	return pages, f.pagesErr
}

// Use literal API field names rather than marshaling the parser's Go types.
func searchFixturePage(t *testing.T, found, start, count int) []byte {
	t.Helper()
	hits := make([]map[string]any, 0, count)
	for index := start; index < start+count; index++ {
		hits = append(hits, map[string]any{"document": map[string]any{
			"slug":             fmt.Sprintf("tool-%03d", index),
			"tool_of_the_week": index%7 == 0,
		}})
	}
	data, err := json.Marshal(map[string]any{"found": found, "hits": hits})
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func TestFetchAndParseSlugs(t *testing.T) {
	cases := []struct {
		name      string
		limit     int
		wantCount int
		wantPages int
	}{
		{name: "no limit", limit: 0, wantCount: 205, wantPages: 3},
		{name: "negative limit", limit: -1, wantCount: 205, wantPages: 3},
		{name: "within first page", limit: 2, wantCount: 2, wantPages: 1},
		{name: "exact first page", limit: 100, wantCount: 100, wantPages: 1},
		{name: "cross page boundary", limit: 101, wantCount: 101, wantPages: 2},
		{name: "within last page", limit: 201, wantCount: 201, wantPages: 3},
		{name: "limit above found", limit: 999, wantCount: 205, wantPages: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pages := map[int][]byte{
				3: searchFixturePage(t, 205, 200, 5),
				2: searchFixturePage(t, 205, 100, 100),
				1: searchFixturePage(t, 205, 0, 100),
			}
			fetcher := &searchFixtureFetcher{t: t, ctx: t.Context(), initial: pages[1], pages: pages}
			slugs, weekly, total, err := fetchAndParseSlugs(fetcher.ctx, fetcher, tc.limit)
			if err != nil {
				t.Fatal(err)
			}
			wantSlugs := make([]string, 0, tc.wantCount)
			wantWeekly := make(map[string]bool)
			for index := range tc.wantCount {
				slug := fmt.Sprintf("tool-%03d", index)
				wantSlugs = append(wantSlugs, slug)
				wantWeekly[slug] = index%7 == 0
			}
			if !slices.Equal(slugs, wantSlugs) {
				t.Fatalf("slugs = %v, want %v in page order", slugs, wantSlugs)
			}
			if !reflect.DeepEqual(weekly, wantWeekly) {
				t.Fatalf("weekly flags = %v, want %v", weekly, wantWeekly)
			}
			if total != 205 {
				t.Fatalf("total = %d, want original found count 205", total)
			}
			if !slices.Equal(fetcher.initialPage, []int{1}) || !slices.Equal(fetcher.pageCounts, []int{tc.wantPages}) {
				t.Fatalf("fetch calls = initial %v, batches %v; want page 1 then %d pages",
					fetcher.initialPage, fetcher.pageCounts, tc.wantPages)
			}
		})
	}
}

func TestFetchAndParseSlugsSkipsEmptySlugs(t *testing.T) {
	page := []byte(`{"found":4,"hits":[
		{"document":{"slug":"","tool_of_the_week":true}},
		{"document":{"slug":"alpha","tool_of_the_week":false}},
		{"document":{"slug":"beta","tool_of_the_week":true}},
		{"document":{"slug":"gamma","tool_of_the_week":false}}
	]}`)
	fetcher := &searchFixtureFetcher{t: t, ctx: t.Context(), initial: page, pages: map[int][]byte{1: page}}
	slugs, weekly, total, err := fetchAndParseSlugs(fetcher.ctx, fetcher, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(slugs, []string{"alpha", "beta"}) ||
		!reflect.DeepEqual(weekly, map[string]bool{"alpha": false, "beta": true}) || total != 4 {
		t.Fatalf("got slugs %v, weekly flags %v, total %d; want alpha/beta only, both flags and found 4",
			slugs, weekly, total)
	}
}

func TestFetchAndParseSlugsEmptyResponse(t *testing.T) {
	page := searchFixturePage(t, 0, 0, 0)
	fetcher := &searchFixtureFetcher{t: t, ctx: t.Context(), initial: page}
	slugs, weekly, total, err := fetchAndParseSlugs(fetcher.ctx, fetcher, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(slugs) != 0 || len(weekly) != 0 || total != 0 {
		t.Fatalf("empty response returned slugs %v, weekly flags %v, total %d", slugs, weekly, total)
	}
	if !slices.Equal(fetcher.initialPage, []int{1}) || !slices.Equal(fetcher.pageCounts, []int{0}) {
		t.Fatalf("empty response fetch calls = initial %v, batches %v; want page 1 then zero pages",
			fetcher.initialPage, fetcher.pageCounts)
	}
}

func TestFetchAndParseSlugsFailures(t *testing.T) {
	fetchErr := errors.New("fixture fetch failed")
	cases := []struct {
		name       string
		initial    []byte
		initialErr error
		pagesErr   error
		wantStage  string
		wantPages  []int
	}{
		{name: "initial fetch", initialErr: fetchErr, wantStage: "initial fetch: "},
		{name: "initial parse", initial: []byte(`{"found":`), wantStage: "parse initial: "},
		{
			name: "page fetch", initial: searchFixturePage(t, 205, 0, 100), pagesErr: fetchErr,
			wantStage: "fetch pages: ", wantPages: []int{3},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fetcher := &searchFixtureFetcher{
				t: t, ctx: t.Context(), initial: tc.initial, initialErr: tc.initialErr, pagesErr: tc.pagesErr,
			}
			slugs, weekly, total, err := fetchAndParseSlugs(fetcher.ctx, fetcher, 0)
			if err == nil || !strings.HasPrefix(err.Error(), tc.wantStage) {
				t.Fatalf("error = %v, want stage %q", err, tc.wantStage)
			}
			if tc.initialErr != nil || tc.pagesErr != nil {
				if !errors.Is(err, fetchErr) {
					t.Fatalf("error %v does not wrap the fetch failure", err)
				}
			} else {
				var syntaxErr *json.SyntaxError
				if !errors.As(err, &syntaxErr) {
					t.Fatalf("error %v does not wrap the JSON syntax failure", err)
				}
			}
			if slugs != nil || weekly != nil || total != 0 {
				t.Fatalf("failure leaked results: slugs %v, flags %v, total %d", slugs, weekly, total)
			}
			if !slices.Equal(fetcher.initialPage, []int{1}) || !slices.Equal(fetcher.pageCounts, tc.wantPages) {
				t.Fatalf("failure fetch calls = initial %v, batches %v; want page 1 then %v",
					fetcher.initialPage, fetcher.pageCounts, tc.wantPages)
			}
		})
	}
}

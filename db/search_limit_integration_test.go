package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestSearchLimitAfterInstalledFiltering(t *testing.T) {
	database := setupSearchLimitDB(t, 2000)
	installed := map[int]bool{1001: true, 1002: true, 1100: true}
	for i := range installed {
		addSearchLimitInstall(t, database, i)
	}
	for _, i := range []int{550, 1450} {
		if err := database.AddTag(fmt.Sprintf("%06d", i), "cli"); err != nil {
			t.Fatal(err)
		}
	}
	installedFilter := &Filter{Type: FilterField, Field: "installed", Value: "true"}
	orFilter := &Filter{Type: FilterOr, Left: installedFilter, Right: &Filter{
		Type: FilterField, Field: "tag", Value: "cli",
	}}

	for _, tc := range []struct {
		name      string
		filter    *Filter
		limit     int
		sortField string
		sortOrder string
		want      []int
	}{
		{"sparse ascending", installedFilter, 2, "name", "ASC", []int{1001, 1002}},
		{"sparse descending", installedFilter, 2, "name", "DESC", []int{1100, 1002}},
		{"exhaustion before limit", installedFilter, 10, "name", "ASC", []int{1001, 1002, 1100}},
		{"OR ascending", orFilter, 3, "name", "ASC", []int{550, 1001, 1002}},
		{"OR descending", orFilter, 3, "name", "desc", []int{1450, 1100, 1002}},
		{"equal sort keys ascending", installedFilter, 3, "tagline", "ASC", []int{1001, 1002, 1100}},
		{"equal sort keys descending", installedFilter, 3, "tagline", "DESC", []int{1001, 1002, 1100}},
		{"mixed case equal sort keys", installedFilter, 3, "language", "ASC", []int{1001, 1002, 1100}},
		{"date sort", installedFilter, 3, "date_published", "DESC", []int{1001, 1002, 1100}},
		{"zero limit", installedFilter, 0, "name", "ASC", []int{1001, 1002, 1100}},
		{"negative limit", installedFilter, -1, "name", "ASC", []int{1001, 1002, 1100}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			results, err := database.Search(ctx, SearchOptions{
				Filter: tc.filter, Query: "fixture", Limit: tc.limit,
				SortField: tc.sortField, SortOrder: tc.sortOrder,
			})
			if err != nil {
				t.Fatal(err)
			}
			var want []string
			for _, i := range tc.want {
				want = append(want, fmt.Sprintf("%06d", i))
			}
			var got []string
			for _, result := range results {
				got = append(got, result.ID)
				if result.Installed != installedID(installed, result.ID) {
					t.Errorf("%s: unexpected installed status %v", result.ID, result.Installed)
				}
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("matching IDs = %v, want %v", got, want)
			}
		})
	}

	for _, filter := range []*Filter{nil, {Type: FilterField, Field: "installed", Value: "false"}} {
		for _, limit := range []int{1500, 0, -1} {
			t.Run(fmt.Sprintf("filter=%v/limit=%d", filter, limit), func(t *testing.T) {
				results, err := database.Search(context.Background(), SearchOptions{
					Filter: filter, Limit: limit, SortField: "name",
				})
				if err != nil {
					t.Fatal(err)
				}
				var want []string
				for i := range 2000 {
					if filter != nil && installed[i] {
						continue
					}
					want = append(want, fmt.Sprintf("%06d", i))
					if limit > 0 && len(want) == limit {
						break
					}
				}
				var got []string
				for _, result := range results {
					got = append(got, result.ID)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("got %d IDs, want %d in exact order", len(got), len(want))
				}
			})
		}
	}
}

func TestSearchInstalledUnlimitedBeyondFormerCap(t *testing.T) {
	database := setupSearchLimitDB(t, 100001)
	addSearchLimitInstall(t, database, 100000)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	results, err := database.Search(ctx, SearchOptions{
		Filter: &Filter{Type: FilterField, Field: "installed", Value: "true"},
		Limit:  0, SortField: "name",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].ID != "100000" || !results[0].Installed {
		t.Fatalf("expected installed tool 100000 beyond the old cap, got %v", results)
	}
}

func TestSearchLimitCanceledContext(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := database.Search(ctx, SearchOptions{
		Filter: &Filter{Type: FilterField, Field: "installed", Value: "true"}, Limit: 2,
	})
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func setupSearchLimitDB(t *testing.T, count int) *SQLiteDB {
	t.Helper()
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "fixture-installed"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	database := setupTestDB(t)
	t.Cleanup(func() { checkClose(t, database) })
	_, err := database.getDB().ExecContext(context.Background(), `
		WITH RECURSIVE numbers(i) AS (SELECT 0 UNION ALL SELECT i+1 FROM numbers WHERE i+1 < ?)
		INSERT INTO tools (id, slug, name, tagline, description, language, license, date_published, code_repository)
		SELECT printf('%06d', i), printf('%06d', i), printf('%06d', i), 'fixture', '',
			CASE WHEN i%2=0 THEN 'Rust' ELSE 'rust' END, '', '2026-01-01', '' FROM numbers
	`, count)
	if err != nil {
		t.Fatal(err)
	}
	return database
}

func addSearchLimitInstall(t *testing.T, database *SQLiteDB, i int) {
	t.Helper()
	id := fmt.Sprintf("%06d", i)
	if err := database.UpsertInstallInstruction(context.Background(), &InstallInstruction{
		ID: "install-" + id, ToolID: id, Platform: "linux", ExecutableName: "fixture-installed",
	}); err != nil {
		t.Fatal(err)
	}
}

func installedID(installed map[int]bool, id string) bool {
	for i := range installed {
		if id == fmt.Sprintf("%06d", i) {
			return true
		}
	}
	return false
}

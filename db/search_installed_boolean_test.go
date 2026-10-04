package db_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"troveler/db"
	"troveler/internal/search"
)

func TestSearchInstalledBooleanExpressions(t *testing.T) {
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "fixture-installed"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Error(err)
		}
	})

	fixtures := []struct {
		id        string
		language  string
		tagged    bool
		installed bool
	}{
		{"a", "Go", true, true},
		{"b", "Rust", true, true},
		{"c", "Go", true, false},
		{"d", "Rust", true, false},
		{"e", "Rust", false, true},
		{"f", "Go", false, false},
		{"g", "Rust", false, false},
	}
	installedByID := make(map[string]bool)
	for _, fixture := range fixtures {
		tool := &db.Tool{
			ID: fixture.id, Slug: fixture.id, Name: fixture.id,
			Language: fixture.language, Tagline: "fixture " + fixture.id,
		}
		binary := "fixture-missing"
		if fixture.installed {
			binary = "fixture-installed"
		}
		installs := []db.InstallInstruction{{
			ID: "install-" + fixture.id, Platform: "linux", ExecutableName: binary,
		}}
		// Include a tool without install instructions: it must count as uninstalled.
		if fixture.id == "g" {
			installs = nil
		}
		if err := database.SaveToolSnapshot(context.Background(), tool, installs); err != nil {
			t.Fatal(err)
		}
		if fixture.tagged {
			if err := database.AddTag(tool.Slug, "cli"); err != nil {
				t.Fatal(err)
			}
		}
		installedByID[fixture.id] = fixture.installed
	}

	for _, tc := range []struct {
		query string
		want  []string
	}{
		{"installed=true", []string{"a", "b", "e"}},
		{"installed=false", []string{"c", "d", "f", "g"}},
		{"!installed=true", []string{"c", "d", "f", "g"}},
		{"!installed=false", []string{"a", "b", "e"}},
		{"!!installed=true", []string{"a", "b", "e"}},
		{"installed=true|language=go", []string{"a", "b", "c", "e", "f"}},
		{"language=go|installed=true", []string{"a", "b", "c", "e", "f"}},
		{"installed=false|language=go", []string{"a", "c", "d", "f", "g"}},
		{"!(installed=true|language=go)", []string{"d", "g"}},
		{"!(installed=true&language=go)", []string{"b", "c", "d", "e", "f", "g"}},
		{"(installed=true|language=go)&tag=CLI", []string{"a", "b", "c"}},
		{"(!installed=true&tag=cli)|(installed=true&language=rust)", []string{"b", "c", "d", "e"}},
		{"installed=true&installed=false", nil},
		{"installed=true|installed=false", []string{"a", "b", "c", "d", "e", "f", "g"}},
		{"installed=true&installed=true", []string{"a", "b", "e"}},
		{"installed=true&!installed=true", nil},
		{"INSTALLED=true", []string{"a", "b", "e"}},
		{"installed=1", []string{"a", "b", "e"}},
		{"installed=0", []string{"c", "d", "f", "g"}},
		{"(installed=true|language=go)&tagline=FIXTURE_b", []string{"b"}},
		{"(installed=true|language=go)&name=%", []string{"a", "b", "c", "e", "f"}},
		{"fixture_b (installed=true|language=go)", []string{"b"}},
		{"fixture_d (installed=true|language=go)", nil},
	} {
		t.Run(tc.query, func(t *testing.T) {
			filter, query, warning, err := search.ParseFilters(tc.query)
			if err != nil || warning != "" || filter == nil {
				t.Fatalf("ParseFilters: filter=%v, warning=%q, err=%v", filter, warning, err)
			}
			results, err := database.Search(context.Background(), db.SearchOptions{
				Filter: filter, Query: query, Limit: 100, SortField: "name",
			})
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, result := range results {
				got = append(got, result.ID)
				if result.Installed != installedByID[result.ID] {
					t.Errorf("%s: installed=%v, want %v", result.ID, result.Installed, installedByID[result.ID])
				}
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("matching IDs = %v, want %v", got, tc.want)
			}
		})
	}
}

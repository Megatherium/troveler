package update

import (
	"context"
	"fmt"
	"testing"

	"troveler/db"
)

type fixtureFetcher struct {
	detail []byte
}

func (f *fixtureFetcher) FetchSearchPage(context.Context, int) ([]byte, error) {
	return []byte(`{"found":1,"hits":[{"document":{"slug":"example"}}]}`), nil
}

func (f *fixtureFetcher) FetchSearchPagesConcurrently(ctx context.Context, _ int) (map[int][]byte, error) {
	page, err := f.FetchSearchPage(ctx, 1)
	return map[int][]byte{1: page}, err
}

func (f *fixtureFetcher) FetchDetailPage(context.Context, string) ([]byte, error) {
	return f.detail, nil
}

func TestFetchAndUpdateRefreshesExistingTool(t *testing.T) {
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	fetcher := &fixtureFetcher{}
	service := &Service{db: database, fetcher: fetcher}
	ctx := context.Background()
	var originalID string
	for i, name := range []string{"Original", "Refreshed"} {
		fetcher.detail = []byte(fmt.Sprintf(
			`<script type="application/ld+json">{"@graph":[{"@type":"SoftwareApplication","name":%q,"url":"https://terminaltrove.com/example/"}]}</script><div id="install" data-install="{&quot;brew&quot;:&quot;brew install %s&quot;}"></div>`,
			name, name,
		))
		if err := service.FetchAndUpdate(ctx, Options{}); err != nil {
			t.Fatal(err)
		}
		tools, err := database.GetAllTools(ctx)
		if err != nil || len(tools) != 1 || tools[0].Name != name {
			t.Fatalf("incorrect tools: %v, error = %v", tools, err)
		}
		if i == 0 {
			originalID = tools[0].ID
			if err := database.AddTag("example", "favorite"); err != nil {
				t.Fatal(err)
			}
		}
		if tools[0].ID != originalID {
			t.Fatalf("tool identity changed: %q != %q", tools[0].ID, originalID)
		}
		installs, err := database.GetInstallInstructions(originalID)
		if err != nil || len(installs) != 1 || installs[0].Command != "brew install "+name {
			t.Fatalf("incorrect installs: %v, error = %v", installs, err)
		}
	}
	tags, err := database.GetTags("example")
	if err != nil || len(tags) != 1 || tags[0] != "favorite" {
		t.Fatalf("lost tags: %v, error = %v", tags, err)
	}
}

func TestFetchAndUpdateReportsWriteFailure(t *testing.T) {
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	fetcher := &fixtureFetcher{detail: []byte(`<script type="application/ld+json">{"@graph":[{"@type":"SoftwareApplication","name":"Example","url":"https://terminaltrove.com/example/"}]}</script>`)}
	service := &Service{db: database, fetcher: fetcher}
	progress := make(chan ProgressUpdate, 10)
	if err := service.FetchAndUpdate(context.Background(), Options{Progress: progress}); err == nil {
		t.Fatal("expected a database write failure")
	}
	foundError := false
	for len(progress) > 0 {
		message := <-progress
		if message.Type == "complete" {
			t.Fatal("failed update reported success")
		}
		if message.Type == "error" && message.Error != nil {
			foundError = true
		}
	}
	if !foundError {
		t.Fatal("failed update did not emit an error event")
	}
}

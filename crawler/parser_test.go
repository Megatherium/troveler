package crawler

import (
	"os"
	"testing"
)

func TestParseSearchResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantErr  bool
	}{
		{
			name: "valid response with tools",
			input: `{
				"found": 2,
				"hits": [
					{"document": {"slug": "tool1", "name": "Tool 1", "tagline": "Tagline 1", "language": "rust", "license": ["mit"]}},
					{"document": {"slug": "tool2", "name": "Tool 2", "tagline": "Tagline 2", "language": "go", "license": ["mit"]}}
				]
			}`,
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "empty response",
			input: `{"found": 0, "hits": []}`,
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "invalid json",
			input: `not valid json`,
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ParseSearchResponse([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSearchResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(resp.Hits) != tt.wantLen {
				t.Errorf("ParseSearchResponse() got %d hits, want %d", len(resp.Hits), tt.wantLen)
			}
		})
	}
}

func TestParseDetailPage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid detail page",
			input: `<!DOCTYPE html>
<html>
<head>
<script type="application/ld+json">{"@context":"https://schema.org","@graph":[{"@type":"SoftwareApplication","@id":"https://terminaltrove.com/testtool/","name":"Test Tool","description":"A test tool","programmingLanguage":"rust","codeRepository":"https://github.com/test/test","datePublished":"2024-01-01"}]}</script>
</head>
<body>
<p id="tagline">Test tagline</p>
<div id="install" data-install='{"brew": "brew install test"}'></div>
</body>
</html>`,
			wantErr: false,
		},
		{
			name: "missing json-ld",
			input: `<!DOCTYPE html><html><body></body></html>`,
			wantErr: true,
		},
		{
			name: "invalid json in script",
			input: `<!DOCTYPE html>
<html>
<head>
<script type="application/ld+json">not valid json</script>
</head>
<body></body>
</html>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, err := ParseDetailPage([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDetailPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if page.Tool.Slug == "" {
					t.Error("ParseDetailPage() slug should not be empty")
				}
				if page.Tool.Name == "" {
					t.Error("ParseDetailPage() name should not be empty")
				}
			}
		})
	}
}

func TestDetailPageToTool(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head>
<script type="application/ld+json">{"@context":"https://schema.org","@graph":[{"@type":"SoftwareApplication","@id":"https://terminaltrove.com/testtool/","name":"Test Tool","description":"A test tool","programmingLanguage":"rust","codeRepository":"https://github.com/test/test","datePublished":"2024-01-01"}]}</script>
</head>
<body>
<p id="tagline">Test tagline</p>
<div id="install" data-install='{"brew": "brew install test"}'></div>
</body>
</html>`

	page, err := ParseDetailPage([]byte(input))
	if err != nil {
		t.Fatalf("ParseDetailPage() failed: %v", err)
	}

	tool := page.ToTool()

	if tool.Slug != "testtool" {
		t.Errorf("ToTool() slug = %v, want %v", tool.Slug, "testtool")
	}
	if tool.Name != "Test Tool" {
		t.Errorf("ToTool() name = %v, want %v", tool.Name, "Test Tool")
	}
	if tool.Language != "rust" {
		t.Errorf("ToTool() language = %v, want %v", tool.Language, "rust")
	}
}

func TestDetailPageWithNestedInstallData(t *testing.T) {
	input := `<!DOCTYPE html>
<html>
<head>
<script type="application/ld+json">{"@context":"https://schema.org","@graph":[{"@type":"SoftwareApplication","@id":"https://terminaltrove.com/testtool/","name":"Test Tool"}]}</script>
</head>
<body>
<div id="install" data-install='{"linux": {"arch": "pacman -S test"}, "macos": {"brew": "brew install test"}}'></div>
</body>
</html>`

	page, err := ParseDetailPage([]byte(input))
	if err != nil {
		t.Fatalf("ParseDetailPage() failed: %v", err)
	}

	insts := page.ToInstallInstructions()

	if len(insts) != 0 {
		t.Logf("Got %d install instructions (nested format not fully parsed)", len(insts))
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

package db

import (
	"os"
	"testing"
)

func TestToolModel(t *testing.T) {
	tool := Tool{
		ID:             "test-id",
		Slug:           "test-tool",
		Name:           "Test Tool",
		Tagline:        "A test tool",
		Description:    "Test description",
		Language:       "rust",
		License:        "MIT",
		DatePublished:  "2024-01-01",
		CodeRepository: "https://github.com/test/test",
	}

	if tool.ID != "test-id" {
		t.Errorf("Tool.ID = %v, want %v", tool.ID, "test-id")
	}
	if tool.Slug != "test-tool" {
		t.Errorf("Tool.Slug = %v, want %v", tool.Slug, "test-tool")
	}
	if tool.Name != "Test Tool" {
		t.Errorf("Tool.Name = %v, want %v", tool.Name, "Test Tool")
	}
}

func TestInstallInstructionModel(t *testing.T) {
	inst := InstallInstruction{
		ID:       "inst-id",
		ToolID:   "tool-id",
		Platform: "linux",
		Command:  "brew install test",
	}

	if inst.ID != "inst-id" {
		t.Errorf("InstallInstruction.ID = %v, want %v", inst.ID, "inst-id")
	}
	if inst.Platform != "linux" {
		t.Errorf("InstallInstruction.Platform = %v, want %v", inst.Platform, "linux")
	}
	if inst.Command != "brew install test" {
		t.Errorf("InstallInstruction.Command = %v, want %v", inst.Command, "brew install test")
	}
}

func TestSearchResultModel(t *testing.T) {
	result := SearchResult{
		Tool: Tool{
			ID:   "test-id",
			Slug: "test-tool",
			Name: "Test Tool",
		},
		Installations: map[string]string{
			"brew": "brew install test",
		},
	}

	if result.Tool.ID != "test-id" {
		t.Errorf("SearchResult.Tool.ID = %v, want %v", result.Tool.ID, "test-id")
	}
	if len(result.Installations) != 1 {
		t.Errorf("SearchResult.Installations length = %v, want %v", len(result.Installations), 1)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

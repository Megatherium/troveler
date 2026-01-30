package search

import (
	"testing"

	"troveler/db"
)

func TestParseFiltersNoFilters(t *testing.T) {
	ast, searchTerm, err := ParseFilters("bat")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ast != nil {
		t.Errorf("expected nil AST, got %+v", ast)
	}
	if searchTerm != "bat" {
		t.Errorf("expected searchTerm 'bat', got '%s'", searchTerm)
	}
}

func TestParseFiltersSimpleField(t *testing.T) {
	ast, searchTerm, err := ParseFilters("name=bat")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != "" {
		t.Errorf("expected empty searchTerm, got '%s'", searchTerm)
	}

	if ast.Type != db.FilterField {
		t.Errorf("expected db.FilterField, got %v", ast.Type)
	}
	if ast.Field != "name" {
		t.Errorf("expected field 'name', got '%s'", ast.Field)
	}
	if ast.Value != "bat" {
		t.Errorf("expected value 'bat', got '%s'", ast.Value)
	}
}

func TestParseFiltersAnd(t *testing.T) {
	ast, _, err := ParseFilters("name=bat&language=rust")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterAnd {
		t.Errorf("expected db.FilterAnd, got %v", ast.Type)
	}
	if ast.Left.Type != db.FilterField || ast.Left.Field != "name" || ast.Left.Value != "bat" {
		t.Errorf("unexpected left filter: %+v", ast.Left)
	}
	if ast.Right.Type != db.FilterField || ast.Right.Field != "language" || ast.Right.Value != "rust" {
		t.Errorf("unexpected right filter: %+v", ast.Right)
	}
}

func TestParseFiltersOr(t *testing.T) {
	ast, _, err := ParseFilters("name=bat|name=batcat")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterOr {
		t.Errorf("expected db.FilterOr, got %v", ast.Type)
	}
	if ast.Left.Type != db.FilterField || ast.Left.Field != "name" || ast.Left.Value != "bat" {
		t.Errorf("unexpected left filter: %+v", ast.Left)
	}
	if ast.Right.Type != db.FilterField || ast.Right.Field != "name" || ast.Right.Value != "batcat" {
		t.Errorf("unexpected right filter: %+v", ast.Right)
	}
}

func TestParseFiltersWithParentheses(t *testing.T) {
	ast, _, err := ParseFilters("(name=git|tagline=git)&language=go")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterAnd {
		t.Errorf("expected db.FilterAnd, got %v", ast.Type)
	}

	// Left should be OR with parentheses
	if ast.Left.Type != db.FilterOr {
		t.Errorf("expected db.FilterOr on left, got %v", ast.Left.Type)
	}

	// Right should be language filter
	if ast.Right.Type != db.FilterField || ast.Right.Field != "language" || ast.Right.Value != "go" {
		t.Errorf("unexpected right filter: %+v", ast.Right)
	}
}

func TestParseFiltersInstalled(t *testing.T) {
	ast, searchTerm, err := ParseFilters("installed=true")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != "" {
		t.Errorf("expected empty searchTerm, got '%s'", searchTerm)
	}

	if ast.Type != db.FilterField {
		t.Errorf("expected db.FilterField, got %v", ast.Type)
	}
	if ast.Field != "installed" {
		t.Errorf("expected field 'installed', got '%s'", ast.Field)
	}
	if ast.Value != "true" {
		t.Errorf("expected value 'true', got '%s'", ast.Value)
	}
}

func TestParseFiltersTagline(t *testing.T) {
	ast, _, err := ParseFilters("tagline=cli")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterField {
		t.Errorf("expected db.FilterField, got %v", ast.Type)
	}
	if ast.Field != "tagline" {
		t.Errorf("expected field 'tagline', got '%s'", ast.Field)
	}
	if ast.Value != "cli" {
		t.Errorf("expected value 'cli', got '%s'", ast.Value)
	}
}

func TestParseFiltersMissingClosingParen(t *testing.T) {
	_, _, err := ParseFilters("(name=bat")

	if err == nil {
		t.Error("expected error for missing closing parenthesis")
	}
}

func TestParseFiltersInvalidSyntax(t *testing.T) {
	ast, searchTerm, err := ParseFilters("name bat")

	// "name bat" without = is treated as a search term, not an error
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ast != nil {
		t.Errorf("expected nil AST for plain search, got %+v", ast)
	}
	if searchTerm != "name bat" {
		t.Errorf("expected searchTerm 'name bat', got '%s'", searchTerm)
	}
}

package search

import (
	"testing"

	"troveler/db"
)

const (
	filterFieldLanguage  = "language"
	testToolBat          = "bat"
	testFilterCLI        = "cli"
	testFilterTag        = "tag"
	testFilterName       = "name"
	filterFieldInstalled = "installed"
	filterValueTrue      = "true"
	testSearchModel      = "model"
)

func TestParseFiltersNoFilters(t *testing.T) {
	ast, searchTerm, _, err := ParseFilters(testToolBat)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ast != nil {
		t.Errorf("expected nil AST, got %+v", ast)
	}
	if searchTerm != testToolBat {
		t.Errorf("expected searchTerm 'bat', got '%s'", searchTerm)
	}
}

func TestParseFiltersSimpleField(t *testing.T) {
	ast, searchTerm, _, err := ParseFilters("name=bat")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != "" {
		t.Errorf("expected empty searchTerm, got '%s'", searchTerm)
	}

	if ast.Type != db.FilterField {
		t.Errorf("expected db.FilterField, got %v", ast.Type)
	}
	if ast.Field != testFilterName {
		t.Errorf("expected field 'name', got '%s'", ast.Field)
	}
	if ast.Value != testToolBat {
		t.Errorf("expected value 'bat', got '%s'", ast.Value)
	}
}

func TestParseFiltersAnd(t *testing.T) {
	ast, _, _, err := ParseFilters("name=bat&language=rust")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterAnd {
		t.Errorf("expected db.FilterAnd, got %v", ast.Type)
	}
	if ast.Left.Type != db.FilterField || ast.Left.Field != testFilterName || ast.Left.Value != testToolBat {
		t.Errorf("unexpected left filter: %+v", ast.Left)
	}
	if ast.Right.Type != db.FilterField || ast.Right.Field != filterFieldLanguage || ast.Right.Value != "rust" {
		t.Errorf("unexpected right filter: %+v", ast.Right)
	}
}

func TestParseFiltersOr(t *testing.T) {
	ast, _, _, err := ParseFilters("name=bat|name=batcat")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterOr {
		t.Errorf("expected db.FilterOr, got %v", ast.Type)
	}
	if ast.Left.Type != db.FilterField || ast.Left.Field != testFilterName || ast.Left.Value != testToolBat {
		t.Errorf("unexpected left filter: %+v", ast.Left)
	}
	if ast.Right.Type != db.FilterField || ast.Right.Field != "name" || ast.Right.Value != "batcat" {
		t.Errorf("unexpected right filter: %+v", ast.Right)
	}
}

func TestParseFiltersWithParentheses(t *testing.T) {
	ast, _, _, err := ParseFilters("(name=git|tagline=git)&language=go")

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
	if ast.Right.Type != db.FilterField || ast.Right.Field != filterFieldLanguage || ast.Right.Value != "go" {
		t.Errorf("unexpected right filter: %+v", ast.Right)
	}
}

func TestParseFiltersInstalled(t *testing.T) {
	ast, searchTerm, _, err := ParseFilters("installed=true")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != "" {
		t.Errorf("expected empty searchTerm, got '%s'", searchTerm)
	}

	if ast.Type != db.FilterField {
		t.Errorf("expected db.FilterField, got %v", ast.Type)
	}
	if ast.Field != filterFieldInstalled {
		t.Errorf("expected field 'installed', got '%s'", ast.Field)
	}
	if ast.Value != filterValueTrue {
		t.Errorf("expected value 'true', got '%s'", ast.Value)
	}
}

func TestParseFiltersTagline(t *testing.T) {
	ast, _, _, err := ParseFilters("tagline=cli")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterField {
		t.Errorf("expected db.FilterField, got %v", ast.Type)
	}
	if ast.Field != "tagline" {
		t.Errorf("expected field 'tagline', got '%s'", ast.Field)
	}
	if ast.Value != testFilterCLI {
		t.Errorf("expected value 'cli', got '%s'", ast.Value)
	}
}

func TestParseFiltersMissingClosingParen(t *testing.T) {
	ast, searchTerm, _, err := ParseFilters("(name=bat")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ast != nil {
		t.Errorf("expected nil AST for malformed filter, got %+v", ast)
	}
	if searchTerm != "(name=bat" {
		t.Errorf("expected searchTerm '(name=bat', got '%s'", searchTerm)
	}
}

func TestParseFiltersInvalidSyntax(t *testing.T) {
	ast, searchTerm, _, err := ParseFilters("name bat")

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

func TestParseFiltersNotSimple(t *testing.T) {
	ast, _, _, err := ParseFilters("!installed=true")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterNot {
		t.Errorf("expected db.FilterNot, got %v", ast.Type)
	}
	if ast.Left == nil || ast.Left.Type != db.FilterField {
		t.Errorf("expected FilterField child, got %+v", ast.Left)
	}
	if ast.Left.Field != filterFieldInstalled || ast.Left.Value != filterValueTrue {
		t.Errorf("expected installed=true, got field=%s value=%s", ast.Left.Field, ast.Left.Value)
	}
}

func TestParseFiltersNotWithParens(t *testing.T) {
	ast, _, _, err := ParseFilters("!(language=go|language=rust)")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterNot {
		t.Errorf("expected db.FilterNot, got %v", ast.Type)
	}
	if ast.Left.Type != db.FilterOr {
		t.Errorf("expected FilterOr child, got %v", ast.Left.Type)
	}
}

func TestParseFiltersNotWithAnd(t *testing.T) {
	ast, _, _, err := ParseFilters("!installed=true&language=go")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterAnd {
		t.Errorf("expected db.FilterAnd, got %v", ast.Type)
	}
	if ast.Left.Type != db.FilterNot {
		t.Errorf("expected FilterNot on left, got %v", ast.Left.Type)
	}
	if ast.Right.Type != db.FilterField || ast.Right.Field != filterFieldLanguage {
		t.Errorf("expected language field on right, got %+v", ast.Right)
	}
}

func TestParseFiltersDoubleNot(t *testing.T) {
	ast, _, _, err := ParseFilters("!!installed=true")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterNot {
		t.Errorf("expected db.FilterNot, got %v", ast.Type)
	}
	if ast.Left.Type != db.FilterNot {
		t.Errorf("expected inner FilterNot, got %v", ast.Left.Type)
	}
	if ast.Left.Left.Type != db.FilterField {
		t.Errorf("expected FilterField at deepest level, got %v", ast.Left.Left.Type)
	}
}

func TestParseFiltersTag(t *testing.T) {
	ast, _, _, err := ParseFilters("tag=cli")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterField {
		t.Errorf("expected db.FilterField, got %v", ast.Type)
	}
	if ast.Field != testFilterTag {
		t.Errorf("expected field 'tag', got '%s'", ast.Field)
	}
	if ast.Value != testFilterCLI {
		t.Errorf("expected value 'cli', got '%s'", ast.Value)
	}
}

func TestParseFiltersTagAndTag(t *testing.T) {
	ast, _, _, err := ParseFilters("tag=cli&tag=fuzzy")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterAnd {
		t.Errorf("expected db.FilterAnd, got %v", ast.Type)
	}
	if ast.Left.Type != db.FilterField || ast.Left.Field != testFilterTag || ast.Left.Value != testFilterCLI {
		t.Errorf("unexpected left filter: %+v", ast.Left)
	}
	if ast.Right.Type != db.FilterField || ast.Right.Field != testFilterTag || ast.Right.Value != "fuzzy" {
		t.Errorf("unexpected right filter: %+v", ast.Right)
	}
}

func TestParseFiltersTagOrTag(t *testing.T) {
	ast, _, _, err := ParseFilters("tag=cli|tag=tui")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterOr {
		t.Errorf("expected db.FilterOr, got %v", ast.Type)
	}
	if ast.Left.Type != db.FilterField || ast.Left.Field != testFilterTag || ast.Left.Value != testFilterCLI {
		t.Errorf("unexpected left filter: %+v", ast.Left)
	}
	if ast.Right.Type != db.FilterField || ast.Right.Field != testFilterTag || ast.Right.Value != "tui" {
		t.Errorf("unexpected right filter: %+v", ast.Right)
	}
}

func TestParseFiltersNotTag(t *testing.T) {
	ast, _, _, err := ParseFilters("!tag=experimental")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterNot {
		t.Errorf("expected db.FilterNot, got %v", ast.Type)
	}
	if ast.Left == nil || ast.Left.Type != db.FilterField {
		t.Errorf("expected FilterField child, got %+v", ast.Left)
	}
	if ast.Left.Field != "tag" || ast.Left.Value != "experimental" {
		t.Errorf("expected tag=experimental, got field=%s value=%s", ast.Left.Field, ast.Left.Value)
	}
}

func TestParseFiltersTagWithLanguage(t *testing.T) {
	ast, _, _, err := ParseFilters("language=go&tag=ai")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterAnd {
		t.Errorf("expected db.FilterAnd, got %v", ast.Type)
	}
	if ast.Left.Type != db.FilterField || ast.Left.Field != "language" || ast.Left.Value != "go" {
		t.Errorf("unexpected left filter: %+v", ast.Left)
	}
	if ast.Right.Type != db.FilterField || ast.Right.Field != "tag" || ast.Right.Value != "ai" {
		t.Errorf("unexpected right filter: %+v", ast.Right)
	}
}

func TestParseFiltersComplexTagExpression(t *testing.T) {
	ast, _, _, err := ParseFilters("(tag=cli|tag=tui)&!tag=experimental")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ast.Type != db.FilterAnd {
		t.Errorf("expected db.FilterAnd at root, got %v", ast.Type)
	}

	if ast.Left.Type != db.FilterOr {
		t.Errorf("expected db.FilterOr on left, got %v", ast.Left.Type)
	}

	if ast.Right.Type != db.FilterNot {
		t.Errorf("expected db.FilterNot on right, got %v", ast.Right.Type)
	}
}

func TestParseFiltersSearchTermWithFilter(t *testing.T) {
	ast, searchTerm, _, err := ParseFilters("model installed=true")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != testSearchModel {
		t.Errorf("expected searchTerm 'model', got '%s'", searchTerm)
	}
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}
	if ast.Type != db.FilterField || ast.Field != filterFieldInstalled || ast.Value != filterValueTrue {
		t.Errorf("expected installed=true filter, got %+v", ast)
	}
}

func TestParseFiltersFilterWithSearchTerm(t *testing.T) {
	ast, searchTerm, _, err := ParseFilters("installed=true model")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != testSearchModel {
		t.Errorf("expected searchTerm 'model', got '%s'", searchTerm)
	}
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}
	if ast.Type != db.FilterField || ast.Field != filterFieldInstalled || ast.Value != filterValueTrue {
		t.Errorf("expected installed=true filter, got %+v", ast)
	}
}

func TestParseFiltersMultipleFiltersWithSearchTerm(t *testing.T) {
	ast, searchTerm, _, err := ParseFilters("model name=bat & language=rust")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != testSearchModel {
		t.Errorf("expected searchTerm 'model', got '%s'", searchTerm)
	}
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}
	if ast.Type != db.FilterAnd {
		t.Errorf("expected FilterAnd at root, got %v", ast.Type)
	}
}

func TestParseFiltersSearchTermWithComplexFilter(t *testing.T) {
	ast, searchTerm, _, err := ParseFilters("model (name=bat|tagline=cli)&!language=rust")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != testSearchModel {
		t.Errorf("expected searchTerm 'model', got '%s'", searchTerm)
	}
	if ast == nil {
		t.Fatal("expected non-nil AST")
	}
	if ast.Type != db.FilterAnd {
		t.Errorf("expected FilterAnd at root, got %v", ast.Type)
	}
}

func TestParseFiltersWarningForOperatorOnly(t *testing.T) {
	_, searchTerm, warning, err := ParseFilters("&")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != "&" {
		t.Errorf("expected searchTerm '&', got '%s'", searchTerm)
	}
	if warning == "" {
		t.Fatal("expected non-empty warning")
	}
	if warning != "Malformed filter \"&\" - using filter expression as search term" {
		t.Errorf("unexpected warning: %s", warning)
	}
}

func TestParseFiltersWarningForMalformedFilterWithSearchTerm(t *testing.T) {
	_, searchTerm, warning, err := ParseFilters("model &")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != testSearchModel {
		t.Errorf("expected searchTerm 'model', got '%s'", searchTerm)
	}
	if warning == "" {
		t.Fatal("expected non-empty warning")
	}
	if warning != "Malformed filter \"&\" - using search term: \"model\"" {
		t.Errorf("unexpected warning: %s", warning)
	}
}

func TestParseFiltersWarningForInvalidEquals(t *testing.T) {
	_, searchTerm, warning, err := ParseFilters("=")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != "=" {
		t.Errorf("expected searchTerm '=', got '%s'", searchTerm)
	}
	if warning == "" {
		t.Fatal("expected non-empty warning")
	}
	if warning != "Malformed filter \"=\" - using filter expression as search term" {
		t.Errorf("unexpected warning: %s", warning)
	}
}

func TestParseFiltersNoWarningForValidQueries(t *testing.T) {
	testCases := []string{
		"bat",
		"name=bat",
		"name=bat&language=rust",
		"model installed=true",
		"installed=true model",
	}

	for _, tc := range testCases {
		_, _, warning, err := ParseFilters(tc)
		if err != nil {
			t.Fatalf("unexpected error for '%s': %v", tc, err)
		}
		if warning != "" {
			t.Errorf("expected no warning for '%s', got: %s", tc, warning)
		}
	}
}

func TestParseFiltersWarningForMultiOperator(t *testing.T) {
	_, searchTerm, warning, err := ParseFilters("&|")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searchTerm != "&|" {
		t.Errorf("expected searchTerm '&|', got '%s'", searchTerm)
	}
	if warning == "" {
		t.Fatal("expected non-empty warning")
	}
	if warning != "Malformed filter \"&|\" - using filter expression as search term" {
		t.Errorf("unexpected warning: %s", warning)
	}
}

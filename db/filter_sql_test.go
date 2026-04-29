package db

import (
	"strings"
	"testing"
)

const testTagCLISQL = "cli"

func TestBuildFilterSQLTag(t *testing.T) {
	filter := &Filter{
		Type:  FilterField,
		Field: "tag",
		Value: testTagCLISQL,
	}

	clause, args := buildFilterSQL(filter)

	if !strings.Contains(clause, "EXISTS") {
		t.Errorf("expected EXISTS in clause, got %s", clause)
	}
	if !strings.Contains(clause, "tool_tags") {
		t.Errorf("expected tool_tags in clause, got %s", clause)
	}
	if !strings.Contains(clause, "tag_name") {
		t.Errorf("expected tag_name in clause, got %s", clause)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
	if args[0] != testTagCLISQL {
		t.Errorf("expected arg 'cli', got %v", args[0])
	}
}

func TestBuildFilterSQLTagCaseInsensitive(t *testing.T) {
	filter := &Filter{
		Type:  FilterField,
		Field: "tag",
		Value: "CLI",
	}

	_, args := buildFilterSQL(filter)

	if args[0] != "cli" {
		t.Errorf("expected tag value normalized to lowercase 'cli', got %v", args[0])
	}
}

func TestBuildFilterSQLTagNot(t *testing.T) {
	filter := &Filter{
		Type: FilterNot,
		Left: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "experimental",
		},
	}

	clause, args := buildFilterSQL(filter)

	if !strings.Contains(clause, "NOT") {
		t.Errorf("expected NOT in clause, got %s", clause)
	}
	if !strings.Contains(clause, "EXISTS") {
		t.Errorf("expected EXISTS in clause, got %s", clause)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestBuildFilterSQLTagAndTag(t *testing.T) {
	filter := &Filter{
		Type: FilterAnd,
		Left: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "cli",
		},
		Right: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "fuzzy",
		},
	}

	clause, args := buildFilterSQL(filter)

	if !strings.Contains(clause, "AND") {
		t.Errorf("expected AND in clause, got %s", clause)
	}
	if strings.Count(clause, "EXISTS") != 2 {
		t.Errorf("expected 2 EXISTS in clause, got %s", clause)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestBuildFilterSQLTagOrTag(t *testing.T) {
	filter := &Filter{
		Type: FilterOr,
		Left: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "cli",
		},
		Right: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "tui",
		},
	}

	clause, args := buildFilterSQL(filter)

	if !strings.Contains(clause, "OR") {
		t.Errorf("expected OR in clause, got %s", clause)
	}
	if strings.Count(clause, "EXISTS") != 2 {
		t.Errorf("expected 2 EXISTS in clause, got %s", clause)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestBuildFilterSQLTagWithLanguage(t *testing.T) {
	filter := &Filter{
		Type: FilterAnd,
		Left: &Filter{
			Type:  FilterField,
			Field: "language",
			Value: "go",
		},
		Right: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "ai",
		},
	}

	clause, args := buildFilterSQL(filter)

	if !strings.Contains(clause, "AND") {
		t.Errorf("expected AND in clause, got %s", clause)
	}
	if !strings.Contains(clause, "language") {
		t.Errorf("expected language in clause, got %s", clause)
	}
	if !strings.Contains(clause, "EXISTS") {
		t.Errorf("expected EXISTS for tag in clause, got %s", clause)
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestBuildFilterSQLNot(t *testing.T) {
	filter := &Filter{
		Type: FilterNot,
		Left: &Filter{
			Type:  FilterField,
			Field: "language",
			Value: "go",
		},
	}

	clause, args := buildFilterSQL(filter)

	if !strings.Contains(clause, "NOT") {
		t.Errorf("expected NOT in clause, got %s", clause)
	}
	if !strings.Contains(clause, "language") {
		t.Errorf("expected language in clause, got %s", clause)
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestBuildFilterSQLNotWithOr(t *testing.T) {
	filter := &Filter{
		Type: FilterNot,
		Left: &Filter{
			Type: FilterOr,
			Left: &Filter{
				Type:  FilterField,
				Field: "language",
				Value: "go",
			},
			Right: &Filter{
				Type:  FilterField,
				Field: "language",
				Value: "rust",
			},
		},
	}

	clause, _ := buildFilterSQL(filter)

	if !strings.Contains(clause, "NOT") {
		t.Errorf("expected NOT in clause, got %s", clause)
	}
	if !strings.Contains(clause, "OR") {
		t.Errorf("expected OR in clause, got %s", clause)
	}
}

func TestGetInstalledFilterInfoNotNegated(t *testing.T) {
	filter := &Filter{
		Type: FilterNot,
		Left: &Filter{
			Type:  FilterField,
			Field: "installed",
			Value: "true",
		},
	}

	value, negated := getInstalledFilterInfo(filter, false)

	if value != "true" {
		t.Errorf("expected value 'true', got '%s'", value)
	}
	if !negated {
		t.Errorf("expected negated=true")
	}
}

func TestGetInstalledFilterInfoDoubleNegation(t *testing.T) {
	filter := &Filter{
		Type: FilterNot,
		Left: &Filter{
			Type: FilterNot,
			Left: &Filter{
				Type:  FilterField,
				Field: "installed",
				Value: "true",
			},
		},
	}

	value, negated := getInstalledFilterInfo(filter, false)

	if value != "true" {
		t.Errorf("expected value 'true', got '%s'", value)
	}
	if negated {
		t.Errorf("expected negated=false (double negation)")
	}
}

func TestInstalledFilterValueNot(t *testing.T) {
	filter := &Filter{
		Type: FilterNot,
		Left: &Filter{
			Type:  FilterField,
			Field: "installed",
			Value: "true",
		},
	}

	wantInstalled, negated := installedFilterValue(filter)

	if !wantInstalled {
		t.Errorf("expected wantInstalled=true, got false")
	}
	if !negated {
		t.Errorf("expected negated=true")
	}
}

func TestHasInstalledInOrContextTrue(t *testing.T) {
	// OR(installed=true, language=go) — installed in OR context
	filter := &Filter{
		Type: FilterOr,
		Left: &Filter{
			Type:  FilterField,
			Field: "installed",
			Value: "true",
		},
		Right: &Filter{
			Type:  FilterField,
			Field: "language",
			Value: "go",
		},
	}

	if !hasInstalledInOrContext(filter) {
		t.Errorf("expected hasInstalledInOrContext=true for OR(installed, language)")
	}
}

func TestHasInstalledInOrContextFalse(t *testing.T) {
	// AND(installed=true, language=go) — installed NOT in OR context
	filter := &Filter{
		Type: FilterAnd,
		Left: &Filter{
			Type:  FilterField,
			Field: "installed",
			Value: "true",
		},
		Right: &Filter{
			Type:  FilterField,
			Field: "language",
			Value: "go",
		},
	}

	if hasInstalledInOrContext(filter) {
		t.Errorf("expected hasInstalledInOrContext=false for AND(installed, language)")
	}
}

func TestHasInstalledInOrContextNil(t *testing.T) {
	if hasInstalledInOrContext(nil) {
		t.Errorf("expected hasInstalledInOrContext=false for nil")
	}
}

func TestHasInstalledInOrContextNestedOr(t *testing.T) {
	// AND(OR(installed=true, language=go), tag=cli) — installed in nested OR
	filter := &Filter{
		Type: FilterAnd,
		Left: &Filter{
			Type: FilterOr,
			Left: &Filter{
				Type:  FilterField,
				Field: "installed",
				Value: "true",
			},
			Right: &Filter{
				Type:  FilterField,
				Field: "language",
				Value: "go",
			},
		},
		Right: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "cli",
		},
	}

	if !hasInstalledInOrContext(filter) {
		t.Errorf("expected hasInstalledInOrContext=true for nested OR")
	}
}

package db

import "testing"

func TestOverfetchLimitWithPositiveRequested(t *testing.T) {
	// 10 * 4 = 40, under the cap
	got := overfetchLimit(10)
	if got != 40 {
		t.Errorf("expected 40, got %d", got)
	}
}

func TestOverfetchLimitCapped(t *testing.T) {
	// 200 * 4 = 800, capped at 500
	got := overfetchLimit(200)
	if got != 500 {
		t.Errorf("expected 500 (capped), got %d", got)
	}
}

func TestOverfetchLimitAtCap(t *testing.T) {
	// 125 * 4 = 500, exactly at cap
	got := overfetchLimit(125)
	if got != 500 {
		t.Errorf("expected 500, got %d", got)
	}
}

func TestOverfetchLimitZeroReturnsFallback(t *testing.T) {
	got := overfetchLimit(0)
	if got != installedNoLimitFallback {
		t.Errorf("expected %d (no-limit fallback), got %d", installedNoLimitFallback, got)
	}
}

func TestOverfetchLimitNegativeReturnsFallback(t *testing.T) {
	got := overfetchLimit(-1)
	if got != installedNoLimitFallback {
		t.Errorf("expected %d (no-limit fallback), got %d", installedNoLimitFallback, got)
	}
}

func TestOverfetchLimitOne(t *testing.T) {
	got := overfetchLimit(1)
	if got != 4 {
		t.Errorf("expected 4, got %d", got)
	}
}

func TestInstalledFilterValueTrue(t *testing.T) {
	filter := &Filter{
		Type:  FilterField,
		Field: "installed",
		Value: "true",
	}

	wantInstalled, negated := installedFilterValue(filter)

	if !wantInstalled {
		t.Errorf("expected wantInstalled=true")
	}
	if negated {
		t.Errorf("expected negated=false")
	}
}

func TestInstalledFilterValueOne(t *testing.T) {
	filter := &Filter{
		Type:  FilterField,
		Field: "installed",
		Value: "1",
	}

	wantInstalled, negated := installedFilterValue(filter)

	if !wantInstalled {
		t.Errorf("expected wantInstalled=true for value '1'")
	}
	if negated {
		t.Errorf("expected negated=false")
	}
}

func TestInstalledFilterValueFalse(t *testing.T) {
	filter := &Filter{
		Type:  FilterField,
		Field: "installed",
		Value: "false",
	}

	wantInstalled, negated := installedFilterValue(filter)

	if wantInstalled {
		t.Errorf("expected wantInstalled=false for value 'false'")
	}
	if negated {
		t.Errorf("expected negated=false")
	}
}

func TestInstalledFilterValueNil(t *testing.T) {
	wantInstalled, negated := installedFilterValue(nil)

	if wantInstalled {
		t.Errorf("expected wantInstalled=false for nil filter")
	}
	if negated {
		t.Errorf("expected negated=false for nil filter")
	}
}

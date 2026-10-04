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

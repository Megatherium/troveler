package install

import (
	"testing"

	"troveler/db"
)

const testPlatformBrew = "brew"

func TestResolvePlatform_DetectedOSMatches(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: "linux:arch", Command: "pacman -S tool"},
		{ID: "2", Platform: testPlatformBrew, Command: "brew install tool"},
	}

	selector := NewPlatformSelector("", "", testPlatformBrew, "")
	result := ResolvePlatform(selector, installs, "arch", "")

	if result.UsedFallback {
		t.Error("Expected no fallback when detected OS matches")
	}

	if len(result.Installs) != 1 {
		t.Fatalf("Expected 1 install, got %d", len(result.Installs))
	}

	if result.Installs[0].Platform != "linux:arch" {
		t.Errorf("Expected linux:arch platform, got %s", result.Installs[0].Platform)
	}
}

func TestResolvePlatform_FallbackRetriedWhenOSYieldsNoMatch(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: testPlatformBrew, Command: "brew install tool"},
		{ID: "2", Platform: "go", Command: "go install"},
	}

	// detectedOS="arch" won't match testPlatformBrew or "go", so fallback=testPlatformBrew should be tried
	selector := NewPlatformSelector("", "", testPlatformBrew, "")
	result := ResolvePlatform(selector, installs, "arch", "")

	if result.UsedFallback {
		t.Error("Expected usedFallback=false because fallback_platform resolved successfully")
	}

	if len(result.Installs) != 1 {
		t.Fatalf("Expected 1 install from fallback, got %d", len(result.Installs))
	}

	if result.PlatformID != testPlatformBrew {
		t.Errorf("Expected platformID=brew, got %s", result.PlatformID)
	}

	if result.Installs[0].Platform != testPlatformBrew {
		t.Errorf("Expected brew platform, got %s", result.Installs[0].Platform)
	}
}

func TestResolvePlatform_FallbackNotRetriedWhenCLIOverrideUsed(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: testPlatformBrew, Command: "brew install tool"},
		{ID: "2", Platform: "go", Command: "go install"},
	}

	// CLI override takes priority, fallback should NOT be tried even if no match
	selector := NewPlatformSelector("windows", "", testPlatformBrew, "")
	result := ResolvePlatform(selector, installs, "arch", "")

	if !result.UsedFallback {
		t.Error("Expected usedFallback=true because CLI override yielded no matches")
	}

	// Should return all installs as fallback, not the filtered brew ones
	if len(result.Installs) != 2 {
		t.Fatalf("Expected 2 installs (all as fallback), got %d", len(result.Installs))
	}
}

func TestResolvePlatform_FallbackNotRetriedWhenConfigOverrideUsed(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: testPlatformBrew, Command: "brew install tool"},
		{ID: "2", Platform: "go", Command: "go install"},
	}

	// Config override takes priority over detected OS, fallback should NOT be tried
	selector := NewPlatformSelector("", "windows", testPlatformBrew, "")
	result := ResolvePlatform(selector, installs, "arch", "")

	if !result.UsedFallback {
		t.Error("Expected usedFallback=true because config override yielded no matches")
	}

	if len(result.Installs) != 2 {
		t.Fatalf("Expected 2 installs (all as fallback), got %d", len(result.Installs))
	}
}

func TestResolvePlatform_FallbackLangRetriedWhenOSYieldsNoMatch(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: "go", Command: "go install"},
		{ID: "2", Platform: testPlatformBrew, Command: "brew install tool"},
	}

	// detectedOS="arch" won't match, fallback="lang" should use language matching
	selector := NewPlatformSelector("", "", "lang", "go")
	result := ResolvePlatform(selector, installs, "arch", "go")

	if result.UsedFallback {
		t.Error("Expected usedFallback=false because lang fallback resolved successfully")
	}

	if len(result.Installs) != 1 {
		t.Fatalf("Expected 1 install from lang fallback, got %d", len(result.Installs))
	}

	if result.Installs[0].Platform != "go" {
		t.Errorf("Expected go platform, got %s", result.Installs[0].Platform)
	}
}

func TestResolvePlatform_NoFallbackWhenFallbackEmpty(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: testPlatformBrew, Command: "brew install tool"},
	}

	// No fallback configured, detected OS doesn't match
	selector := NewPlatformSelector("", "", "", "")
	result := ResolvePlatform(selector, installs, "arch", "")

	if !result.UsedFallback {
		t.Error("Expected usedFallback=true when no fallback configured and OS doesn't match")
	}

	if result.PlatformID != "arch" {
		t.Errorf("Expected platformID=arch, got %s", result.PlatformID)
	}
}

func TestResolvePlatform_FallbackAlsoYieldsNoMatch(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: "winget", Command: "winget install tool"},
	}

	// detectedOS="arch" doesn't match, fallback=testPlatformBrew also doesn't match
	selector := NewPlatformSelector("", "", testPlatformBrew, "")
	result := ResolvePlatform(selector, installs, "arch", "")

	if !result.UsedFallback {
		t.Error("Expected usedFallback=true when both OS and fallback yield no matches")
	}

	if result.PlatformID != "arch" {
		t.Errorf("Expected platformID=arch (original), got %s", result.PlatformID)
	}
}

func TestResolvePlatform_EmptyDetectedOS(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: testPlatformBrew, Command: "brew install tool"},
	}

	// Empty detectedOS: Selector already returns fallback, no retry needed
	selector := NewPlatformSelector("", "", testPlatformBrew, "")
	result := ResolvePlatform(selector, installs, "", "")

	if result.UsedFallback {
		t.Error("Expected no fallback when fallback_platform matches directly")
	}

	if result.PlatformID != testPlatformBrew {
		t.Errorf("Expected platformID=brew, got %s", result.PlatformID)
	}
}

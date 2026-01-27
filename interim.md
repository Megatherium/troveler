# Interim Status - Troveler Refactoring

## Completed Work

### 1. Tagline Column Width Feature ✅ (Committed: `9d47ed0`)

**Implementation:**
- Added `--width/-w <width>` flag to search command
- Added `[search]` section to config file with `tagline_width` setting
- Default width: 50 characters
- Priority: CLI flag > config file value > default (50)

**Config Example:**
```toml
[search]
tagline_width = 30

[install]
fallback_platform = "LANG"
always_run = false
use_sudo = "ask"
```

**Usage Examples:**
```bash
# Default width (50 chars)
troveler search "cli tool"

# Custom width (40 chars)
troveler search "cli tool" --width 40

# Use config file value (30 chars from config above)
troveler search "cli tool" --width 0
```

**Files Changed:**
- `config/config.go`: Added `SearchConfig` struct
- `commands/search.go`: Added width flag, updated truncation logic
- `config/config_test.go`: Added comprehensive config loading tests
- `commands/search_test.go`: Added tagline truncation tests

**Tests:** All passing

---

## In Progress: Platform Override Feature ⚠️

### What We're Implementing

**Goal:** Add `-o/--override <platform>` flag to install command that:
1. Overrides autodetection when specified
2. Allows `platform_override = "LANG"` in config to use language matching
3. Maintains priority system: CLI > platform_override > fallback_platform > OS detection > default

### Challenges Encountered

#### 1. Complex Priority Logic
The platform selection has many conditions:
- CLI argument (highest priority)
- platform_override config setting (medium priority - includes LANG support)
- fallback_platform config setting (lower priority)
- OS detection (lowest priority)
- Language matching (when platform_override = "LANG")

#### 2. Duplicate/Conflicting Logic
Current code has:
- Multiple branches checking same conditions
- Platform normalization logic scattered
- Language matching logic mixed with platform matching
- Duplicate fallback code in multiple places

#### 3. Syntax Errors
The implementation attempts had:
- Missing braces in conditional branches
- Duplicate `else` clauses causing parser confusion
- Variable shadowing issues

#### 4. Test Complexity
Need comprehensive tests for:
- Priority ordering (CLI > override > fallback > OS)
- LANG fallback with language matching
- Language matching with various install platforms
- Edge cases (empty values, invalid platforms)

### Current State

**What Works:**
- Config file parsing with `platform_override` field ✅
- Flag registration for `-o/--override` ✅
- Test cases written (attempted to verify logic)

**What Doesn't Work:**
- Complex platform selection logic has syntax errors
- Tests not passing due to logic complexity
- Duplicate code causing maintenance issues

### Root Cause Analysis

The platform selection logic is trying to handle too many concerns in one function:
1. Priority resolution
2. Platform normalization
3. Language matching
4. OS detection fallback
5. Error handling for no matches

This needs to be broken down into smaller, focused functions.

---

## Recommended Approach

### 1. Refactor Platform Selection Logic

Create separate functions for each concern:

```go
// priority.go
func SelectPlatform(cliArg string, override string, fallback string, detectedOS string) string

// matching.go
func MatchInstallPlatform(desiredPlatform string, installs []InstallInstruction) (string, []InstallInstruction)

// language.go
func MatchLanguagePlatform(toolLanguage string, installs []InstallInstruction) (string, []InstallInstruction)

// normalization.go
func NormalizePlatform(platform string) string (already exists in lib/)
```

### 2. Simplify Priority Logic

```go
func selectPlatform(cliArg string, override string, fallback string, detectedOS string) string {
	// Priority: CLI > Override > Fallback > OS detection > default
	if cliArg != "" {
		return NormalizePlatform(cliArg)
	}
	if override != "" {
		return NormalizePlatform(override)
	}
	if fallback != "" {
		return NormalizePlatform(fallback)
	}
	return detectedOS
}
```

### 3. Handle LANG Special Case

When `override = "LANG"`:
1. Use tool's language
2. Call language matching function
3. Return the matched platform name (e.g., "go (pip)", "cargo")

This needs to return the actual platform name from install instructions, not just "LANG".

---

## Technical Issues Blocking Progress

### Install.go Syntax Errors
- Line ~120-200: Complex nested conditionals with missing braces
- Line ~90-120: Duplicate platform selection logic
- Parser getting confused with `else if` chains

### Test Failures
- `override_test.go`: Created test file but couldn't implement due to syntax errors
- Tests don't validate actual behavior, only mock logic

### Dependencies
- Needs `lib.MatchLanguage` to be examined
- Needs `lib.MatchPlatform` to be examined
- Needs to understand install platform naming conventions

---

## Next Steps (When Context Restarts)

### Immediate:
1. **Fix install.go syntax errors**
   - Resolve nested conditional logic
   - Remove duplicate platform selection code
   - Simplify to single clear function

2. **Implement clean platform selection**
   ```go
   func selectPlatform(...) string {
       // Simple, linear priority check
   }
   ```

3. **Add comprehensive tests**
   - Test priority ordering
   - Test LANG fallback behavior
   - Test language matching
   - Test edge cases

### Medium-term:
4. **Consider extraction** - Move platform matching to separate file
   - Install platform matching logic into `lib/platformmatch.go`
   - Language matching into `lib/languagematch.go`

### Questions to Resolve:
1. **How should `platform_override = "LANG"` work?**
   - Should it use tool's language to match install platforms?
   - Should it return the matched platform name from installs (e.g., "go (pip)")?
   - Or should it just match by language name (go -> matches "go (pip)")?

2. **What happens when no match?**
   - If LANG fallback is set but no language matches, should it fall through to OS detection?
   - Or should it show all install commands?

3. **Platform normalization**: Does `lib.NormalizePlatform()` need updating?
   - Currently handles: "macos", "linux:arch", etc.
   - Does it handle "go (pip)" style formats?
   - Does it handle "python (pip)" style formats?

---

## Files Modified (Not Committed)

- `config/config.go`: Added `PlatformOverride` field ✅
- `commands/install.go`: Attempted `-o/--override` flag and platform selection logic (HAS SYNTAX ERRORS) ⚠️
- `commands/override_test.go`: Test file created (NOT WORKING) ⚠️

**Committed:**
- Tagline width feature (commit `9d47ed0`)

**Needs Work:**
- Fix install.go and complete platform override feature
- Clean up test files

---

## Summary

**What's Done:**
1. Tagline column width configuration - fully implemented and tested ✅
2. Platform override - config added, implementation in progress (blocked by complexity)

**What's Needed:**
1. Fix and complete platform override feature
2. Clean implementation and tests
3. Consider architectural improvements (extracting matching logic)

**Time Estimate:**
- Tagline width: ✅ DONE
- Platform override: ~2-4 hours (depending on complexity of existing code)

**Recommendation:**
Complete tagline width feature, commit it cleanly. Then work on platform override as a separate, focused task with clean test-driven approach.

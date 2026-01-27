# Install Fallback Platform Test Cases

## Problem Description

User reported: "fallback_platform = LANG" should default to using tool's language (e.g., "go"), not literally matching "LANG".

## Expected Behavior

When `fallback_platform = "LANG"`:
1. Tool language should be used for platform matching
2. For a Go tool, should match "go" platforms (e.g., "go (pip)", "go", "go modules")
3. For a Rust tool, should match "rust" platforms (e.g., "cargo", "rust")
4. For a Python tool, should match "python" platforms (e.g., "python (pip)", "python", "python pip")

## Priority Order

Platform selection priority should be:

1. **Command-line argument** (highest priority)
   ```bash
   troveler install qo macos  # Uses macos
   ```

2. **fallback_platform = "LANG"**
   - Use tool's language to match install platforms
   - Example: Tool is Go → match "go (pip)", "go", "go modules"
   ```toml
   [install]
   fallback_platform = "LANG"
   ```

3. **fallback_platform = <specific platform>**
   - Use specified platform directly
   ```toml
   [install]
   fallback_platform = "macos"
   ```

4. **fallback_platform = "" (empty)**
   - Use OS detection
   - Detect current OS and match install platforms

5. **Default** (no config)
   - Use OS detection

## Current Behavior (BUG)

Current code in `commands/install.go`:
```go
if fallbackPlatform == "LANG" {
    fallback = tool.Language
    for _, inst := range installs {
        if lib.MatchLanguage(fallback, inst.Platform) {
            matched = append(matched, inst)
        }
    }
    platform = fallback  // BUG: sets platform to "go" instead of matched install platform
}
```

The bug is: `platform = fallback` sets the platform to the language string (e.g., "go"), but this platform name may not exist in install instructions!

## Expected Fix

When `fallback_platform = "LANG"`:
```go
if fallbackPlatform == "LANG" {
    language := tool.Language
    for _, inst := range installs {
        if lib.MatchLanguage(language, inst.Platform) {
            matched = append(matched, inst)
            platform = inst.Platform  // Use the matched install's platform name
            break
        }
    }
}
```

## Test Scenarios

### Scenario 1: fallback_platform = "LANG", tool = Go
- Config: `fallback_platform = "LANG"`
- Tool language: "go"
- Expected: Match install platforms containing "go" (e.g., "go (pip)", "go modules")
- Display: Install commands for Go language

### Scenario 2: fallback_platform = "LANG", tool = Rust
- Config: `fallback_platform = "LANG"`
- Tool language: "rust"
- Expected: Match install platforms containing "rust" (e.g., "cargo", "rust")
- Display: Install commands for Rust language

### Scenario 3: fallback_platform = "macos", tool = Go
- Config: `fallback_platform = "macos"`
- Tool language: "go"
- Expected: Match install platforms for "macos" (e.g., "brew")
- Display: Install commands for macOS platform

### Scenario 4: No fallback, OS = linux:arch
- Config: no fallback_platform set
- Detected OS: "linux:arch"
- Expected: Match install platforms for "linux:arch" (e.g., "yay", "pacman")
- Display: Install commands for Arch Linux platform

## Platform Matching Rules

Language-based matching (via `lib.MatchLanguage`):
- "go" matches: "go (pip)", "go", "go modules", "go:.*"
- "rust" matches: "cargo", "rust", "rust:.*"
- "python" matches: "python (pip)", "python", "python pip", "pip"
- "node" matches: "npm", "yarn", "node", "node:.*"

Platform-based matching (via `lib.MatchPlatform`):
- "macos" matches: "brew", "macos"
- "linux:arch" matches: "yay", "pacman", "arch"
- "ubuntu" matches: "apt", "ubuntu"
- "fedora" matches: "dnf", "fedora"
- "debian" matches: "apt", "debian"

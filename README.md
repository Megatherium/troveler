# 🔍 Troveler - Terminal Tool Discovery TUI

A blazing-fast Terminal User Interface for browsing and installing CLI tools from [terminaltrove.com](https://terminaltrove.com).

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8.svg)

## ✨ Features

- 🎨 **Beautiful TUI** - 4-panel layout with gradient colors and smooth animations
- 🔍 **Live Search** - 150ms debounced search with instant filtering
- ⌨️  **Keyboard-Driven** - Full keyboard navigation (k/j, h/l, Tab, Alt+keys)
- 📊 **Smart Sorting** - Sort by name, tagline, or language with visual indicators
- 🚀 **One-Click Install** - Execute install commands directly from the TUI
- 🎯 **Auto-Selection** - Info/install panels update as you navigate
- 🌍 **Platform-Aware** - Detects your OS and shows relevant install commands
- 📦 **Offline First** - Local database for fast searches

## 🚀 Quick Start

### Installation

```bash
# Using Go
go install github.com/yourusername/troveler@latest

# Or clone and build
git clone https://github.com/yourusername/troveler.git
cd troveler
go build -o troveler .
```

### First Run

```bash
# Update database (fetch all tools from terminaltrove.com)
troveler update

# Launch TUI
troveler tui

# Or just run troveler (if default_to_tui is enabled in config)
troveler
```

## 🧪 Testing

### Integration Tests with Docker

Troveler includes comprehensive integration tests using Docker to verify functionality across different scenarios:

```bash
# Build the Docker image
cd integration
docker build -t troveler-test .

# Run all integration tests
docker run --rm troveler-test

# Run specific tests interactively
docker run -it --rm troveler-test /bin/sh
./run_tests.sh
```

The test suite includes:
- Basic search functionality
- Install command display (apk, go, cargo, npm, etc.)
- Toolchain verification (mise integration)
- Search filters with complex queries
- Sudo flow as non-root user
- Batch install (dry run)

Dockerfile features:
- Multi-stage build with golang:1.25-alpine
- Pre-installed mise with Go toolchain
- Test user with passwordless sudo
- Pre-populated database for offline testing
- CGO enabled for SQLite support

## ⌨️ Keybindings

### Global

- **Tab** - Cycle through panels (Search → Tools → Install → Search)
- **Alt+Q** - Quit
- **?** - Show help modal
- **Alt+U** - Update database
- **ESC** - Close modals / Clear search

### Search Panel

- **Type** - Live search (debounced)
- **Enter** - Trigger immediate search
- **ESC** - Clear search

### Tools Panel

- **k / ↑** - Move up
- **j / ↓** - Move down
- **h / ←** - Previous column
- **l / →** - Next column
- **Alt+S** - Sort by selected column (▲/▼)
- **Enter** - Select tool and jump to install panel

### Install Panel

- **k / ↑** - Previous command
- **j / ↓** - Next command
- **Alt+I** - Execute selected install command

## 📖 CLI Commands

```bash
# Launch TUI
troveler tui

# Search tools (CLI mode)
troveler search <query>
troveler search language=go
troveler search "name=bat | name=batcat"
troveler search "(name=git|tagline=git)&language=go"

# Show tool info
troveler info <tool-slug>

# Get install commands
troveler install <tool-slug>

# Update database
troveler update

# Shell completion
troveler completion [bash|zsh|fish]
```

### Search Filters

Search supports powerful field-based filtering with the following syntax:

- **Field filters**: Use `field=value` to filter on specific fields
  - `name=bat` - Search for tools with name matching "bat"
  - `tagline=cli` - Search for tools with "cli" in tagline
  - `language=go` - Filter by programming language
  - `installed=true` - Show only installed tools
  - `installed=false` - Show only uninstalled tools

- **Boolean operators**: Combine filters with `&` (AND) and `|` (OR)
  - `name=git&language=go` - Tools named "git" AND written in Go
  - `name=bat|name=batcat` - Tools named "bat" OR "batcat"

- **NOT operator**: Negate filters with `!`
  - `!installed=true` - Show uninstalled tools
  - `!language=go` - Exclude Go tools
  - `!(language=go|language=rust)` - Exclude Go and Rust tools

- **Parentheses**: Group expressions for complex queries
  - `(name=git|tagline=git)&language=go` - (name=git OR tagline=git) AND language=go

- **Wildcards**: Search terms automatically use glob matching (*foo* matches "foo")
  - Field filters also use glob matching (e.g., `tagline=cli` matches "*cli*")

- **Fallback**: If no `=` is found, query is treated as a general search term

**Examples**:
```bash
# Simple field filter
troveler search language=rust

# Multiple filters with AND
troveler search "language=go&tagline=cli"

# Multiple filters with OR
troveler search "name=bat|name=batcat"

# Negate a filter with NOT
troveler search "!installed=true"
troveler search "!language=go"

# Complex query with parentheses
troveler search "(name=git|tagline=git)&language=go"

# Exclude multiple languages
troveler search "!(language=go|language=rust)"

# Filter by installed status
troveler search installed=true

# Combine filters with sort
troveler search language=python --sort name
troveler search installed=true --limit 20
```

## ⚙️ Configuration

Config file: `~/.config/troveler/config.toml`

```toml
# Database settings
db_path = "~/.local/share/troveler/troveler.db"
default_to_tui = false  # Launch TUI when running 'troveler' with no args

# Install behavior
[install]
fallback_platform = "LANG"  # Use language-based matching by default
platform_override = ""      # Force specific platform (e.g., "fedora")
always_run = false          # Auto-execute commands (dangerous!)

# TUI settings
[tui]
theme = "default"
tagline_max_width = 80
gradient_colors = ["#FF00FF", "#00FFFF", "#FF69B4", "#7B68EE", "#A9A9A9"]

# Search settings
[search]
tagline_width = 80
```

## 🎨 TUI Layout

```
┌─────────────────────────────────┬─────────────────────────────────┐
│ 🔍 Search                       │ ℹ️  Info                        │
│ [type to search...]             │ Tool Name v1.0.0                │
├─────────────────────────────────┤ A fantastic CLI tool            │
│ 📊 Tools (42)                   │                                 │
│ Name          Tagline    Lang   │ Description: ...                │
│ > tool1       Fast CLI   Go     │ Language: Go                    │
│   tool2       Smart TUI  Rust   │ License: MIT                    │
│   tool3       Web scrape Python ├─────────────────────────────────┤
│   ...                           │ 🔧 Install                      │
│                                 │ > brew: brew install tool1      │
│                                 │   cargo: cargo install tool1    │
│                                 │                                 │
│                                 │ [Alt+I to execute]              │
└─────────────────────────────────┴─────────────────────────────────┘
Status: Tab to navigate | ? for help | Alt+Q to quit
```

## 🏗️ Architecture

```
troveler/
├── cmd/              # CLI commands (search, info, install, update)
├── config/           # Configuration management
├── crawler/          # terminaltrove.com scraper
├── db/              # SQLite database layer
├── integration/      # Dockerfile and integration tests
├── internal/        # Business logic
│   ├── search/      # Search service
│   ├── install/     # Platform selection & command filtering
│   └── info/        # Tool info formatting
├── tui/             # Terminal UI (bubbletea)
│   ├── panels/      # Search, Tools, Info, Install panels
│   └── styles/      # Lipgloss styles & gradients
└── lib/             # Shared utilities
```

## 🐛 Known Issues & Fixes

### Install Panel Overflow (Fixed)

- **Issue**: Many install entries pushed windows off screen
- **Fix**: Dynamic layout - install panel grows, info panel shrinks

### Platform Selection (Fixed)

- **Issue**: Wrong default platform (e.g., FreeBSD for Fedora users)
- **Fix**: Smart fallback with sensible defaults (brew, apt, pacman priority)

### HTML Entities (Fixed)

- **Issue**: Commands contained `&#39;` instead of `'`
- **Fix**: HTML entity decoding during crawl

## 🤝 Contributing

Contributions welcome! Please:

1. Fork the repo
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Commit with conventional commits (`feat:`, `fix:`, `docs:`)
4. Push and open a PR

## 📝 License

MIT License - See LICENSE file for details

## 🙏 Acknowledgments

- [terminaltrove.com](https://terminaltrove.com) - Amazing CLI tool discovery site
- [Charmbracelet](https://charm.sh) - Beautiful TUI libraries (bubbletea, bubbles, lipgloss)
- [Cobra](https://cobra.dev) - CLI framework

## 🔗 Links

- **Website**: [terminaltrove.com](https://terminaltrove.com)
- **Issues**: [GitHub Issues](https://github.com/yourusername/troveler/issues)
- **Changelog**: See git log for detailed history

---

**Made with ❤️  by the open source community**

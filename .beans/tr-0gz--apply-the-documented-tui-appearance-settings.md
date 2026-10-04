---
# tr-0gz
title: Apply the documented TUI appearance settings
status: completed
type: bug
priority: normal
created_at: 2026-10-03T20:23:09Z
updated_at: 2026-10-04T14:12:10Z
parent: tr-twk
---

Config loads tui.theme, tui.tagline_max_width and tui.gradient_colors, but no production consumer applies these settings. Wire supported options into the TUI and validate values, or explicitly remove unsupported settings and document compatibility.

## Acceptance criteria

Show that custom palette/theme and tagline width affect rendering as documented, that defaults remain usable, and invalid configuration is handled clearly.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

TUIConfig.Resolve applies the existing gradient/40-column defaults and validates supported themes, nonnegative configured widths and #RRGGBB palettes. Config.Load identifies the config file on validation failures; Run also rejects invalid hand-constructed configuration before starting the terminal. NewModel configures its tools panel from these settings and exposes validation failures instead of silently ignoring them.

Tools panels retain copied local palettes: gradient uses the built-in or configured row cycle, custom requires a configured palette, and default uses one color (white or the first configured color). All normal, selected, marked and marked-selected rows use the selected palette while retaining existing backgrounds. Tagline columns are capped by both configured maximum and existing panel availability, with safe grapheme/terminal-column truncation, including widths 1-3. Full Info-panel taglines remain unchanged.

README explains row-theme semantics, color syntax, cap behavior, invalid settings and a custom TOML example. AGENTS.md records local-palette and Unicode requirements. Added fourteen loader validation/control cases, five RGB rendering scenarios across four row states, nine width/Unicode cases and programmatic startup validation. Initial regressions reproduced eighteen failing cases before the implementation.

## Validation

- Config validation and actual NewModel rendering regressions: PASS.
- RGB assertions cover row palettes/cycling, independent model snapshots and normal/selected/marked/marked-selected states, including the corresponding visible backgrounds and mark glyph.
- Width assertions cover configured/default/terminal-limited/large/1-3-column limits, CJK and combining/flag graphemes, with valid UTF-8 and one physical line per row.
- go test -race ./... with TestFetchAndParseSlugs excluded: PASS.
- go vet ./... and go build: PASS.
- Core production golangci-lint (govet/staticcheck/unused/ineffassign): 0 issues.
- git diff --check and beans check --json: PASS.
- Real agent-tui sessions: custom palette plus 12-column cap, default white plus 8-column cap, and built-in gradient plus 40-column cap all render successfully.
- WebSocket live_preview_stream output events preserve the actual ANSI RGB codes; assertions verified both custom colors, default white and built-in green. Screenshots confirmed text caps and the visible mark, and sessions/isolated daemon were cleaned up. Snapshot and ANSI artifacts are in /tmp/troveler-appearance-review/.

Known live crawler exclusions remain tr-4ek and the checked-in linter config repair remains tr-75b. README db_path correction remains separate in tr-zfh.

## Review status

Implementation and self-review are ready. Stopped for independent user review. This fix is uncommitted and must not be pushed or followed by another fix until approved.

Independent review is tracked in tr-5zr. Resume with review feedback before committing or starting another fix. Separate low-priority follow-up tr-8z7 records existing byte-based truncation of names/languages; this fix changes only headers and taglines.

## Approved landing

Independent review tr-5zr approved without refinements. The user confirmed all green and authorized this separate commit and push.

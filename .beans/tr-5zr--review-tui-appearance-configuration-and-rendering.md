---
# tr-5zr
title: Review TUI appearance configuration and rendering
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T14:05:24Z
updated_at: 2026-10-04T14:12:10Z
parent: tr-twk
---

Review tr-0gz before its separate commit. The user requires a stop after every bugfix for independent feedback.

## Scope

- config/tui.go and config/config.go: resolve TUI defaults and validate themes, widths and #RRGGBB palette entries, with path-aware config errors.
- tui/model.go and tui/tui.go: apply settings to each tools panel and reject invalid programmatic config before terminal startup.
- tui/panels/tools.go: copied panel-local palettes, all row states, configured tagline cap and Unicode/terminal-column-safe header/tagline truncation.
- config/tui_validation_test.go and tui/appearance_test.go: loader validation/control cases, actual RGB rendering and background assertions, palette independence/cycling, width and Unicode regressions, startup rejection.
- README.md and one AGENTS.md lesson: theme semantics, color format, width caps and validation.

## Review focus

Confirm gradient keeps the built-in palette unless colors are supplied; custom requires colors and cycles them; default uses a single color (white or the first configured entry). Palettes must remain copied per panel, without mutating global CLI/UI styles or other models. Config slice mutation after model construction must not alter the rendered palette. All selected/marked backgrounds and the mark glyph must remain visible.

Default width is still 40, with zero interpreted as that default. Positive small widths must safely truncate headers/taglines, and terminal availability must constrain larger configured values. Unicode graphemes and wide characters must remain valid/aligned. The full tagline in Info is unchanged. Appearance controls the tools table; other panel accents retain their established styles.

Unknown themes, negative widths, malformed colors and custom themes without a palette must report clear errors. Load includes the selected config path and returns nil; Run validates hand-constructed config before terminal initialization. Existing config and DSN behavior must remain green.

Initial tests reproduced eighteen failing cases. Current coverage includes fourteen loader cases, five RGB scenarios across four row states, nine width cases and an invalid manual startup case. Assertions check actual ANSI RGB and background values, row marks, exact column widths and valid UTF-8, rather than only internal fields.

## Validation

Full race suite with existing live crawler exclusions, vet, build, core production lint (zero issues), whitespace and Beans checks pass. Focused race rendering tests also pass after strengthening mark/background assertions.

Real agent-tui sessions exercised the built CLI with a seeded SQLite database: custom two-color palette and width 12, default white rows and width 8, built-in gradient and width 40. Screenshots verified text truncation and the visible mark; WebSocket live_preview_stream output events verified the actual RGB escapes. Snapshot/ANSI artifacts are in /tmp/troveler-appearance-review/. Owned sessions and the isolated daemon were cleaned up.

Known crawler/linter exceptions remain tr-4ek/tr-75b. README db_path correction is tr-zfh. AGENTS.md contains unrelated user Beans migration edits; stage only the new TUI appearance lesson when landing this fix. Leave migration files and unrelated untracked files outside scope.

- [x] Review the complete implementation and meaningful regression coverage.
- [x] Record approval or actionable feedback for tr-0gz.
- [x] Verify refinements before authorizing the separate commit.

Follow-up tr-8z7 records the existing byte-based truncation of tool names and language labels. Only that follow-up bean was created; name/language rendering is outside the appearance fix.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Theme semantics: gradient uses the built-in palette only when no colors are supplied and the supplied colors otherwise; custom requires at least one color (rejected by Resolve otherwise); default renders a single color — white when none configured, the first configured entry otherwise (all five theme scenarios asserted via actual 38;2;R;G;B escapes under a TrueColor profile, not internal fields).
- Palette independence and aliasing: ConfigureAppearance snapshots every palette with append([]string(nil), ...) into panel-local storage; the panel initializes with a copy of the built-in gradient so rowColors is never empty (no modulo-by-zero). The test explicitly mutates cfg.TUI.GradientColors[0] after model construction AND builds a second model with a different palette, then asserts the first model's rendered RGB values are unchanged. Rows no longer read styles.GradientColors, leaving the shared CLI/UI styles untouched.
- Row states preserved: selected, marked-selected and marked-blurred backgrounds (exact 48;2;... values) and the visible mark glyph are asserted alongside the palette in every scenario.
- Width semantics: default 40 with zero interpreted as default; min(panel-derived, configured) applies the terminal constraint; nine cases cover configured/default/terminal-limited/large/1-2-3 column (dot tails)/wide-CJK/grapheme-cluster taglines with exact ansi.StringWidth assertions, valid UTF-8, one physical line per row, and "..." suffixes.
- Truncation safety: headers and taglines truncate via ansi.Truncate (display-width aware); the sub-three-column tail degrades to dots. Byte-based name/language truncation is pre-existing and correctly deferred to tr-8z7.
- Validation and lifecycle: Load wraps Resolve errors with the selected config path and returns nil; Run rejects hand-constructed invalid config before tea.NewProgram (tested with nil db, proving no terminal/database touch). Info panel's full tagline and all other panel accents are untouched by the diff.
- Loader coverage: fourteen cases asserting both the offending tui.* field and the config path for unknown themes, negative widths, custom-without-palette, and five malformed color forms, plus six valid variants.

Gates independently reproduced: focused appearance/validation suites, go vet, go build, git diff --check, full go test -race ./... with the tr-4ek exclusions — all green. The implementer's agent-tui live sessions and snapshot artifacts provide complementary end-to-end evidence.

tr-0gz is unblocked for its separate commit (code + beans together).

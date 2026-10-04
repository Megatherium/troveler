---
# tr-bob
title: Honor XDG configuration and data directories
status: completed
type: bug
priority: normal
created_at: 2026-10-03T20:23:08Z
updated_at: 2026-10-04T13:23:05Z
parent: tr-twk
---

config/config.go hardcodes HOME/.config and HOME/.local/share, ignoring XDG_CONFIG_HOME and XDG_DATA_HOME. Select the environment-specified XDG locations with the explicit HOME fallbacks required by AGENTS.md.

## Acceptance criteria

Set custom XDG config/data directories and verify default config lookup and DSN selection. Cover unset variables and the existing TROVELER_DSN override.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

Default config lookup now uses absolute XDG_CONFIG_HOME, and the default SQLite DSN uses absolute XDG_DATA_HOME. The settings resolve independently. Unset, empty or relative values use the existing HOME/.config and HOME/.local/share fallbacks; valid XDG roots are evaluated before HOME lookup and therefore work without HOME.

Preserved explicit config path selection, config dsn precedence and the highest-priority TROVELER_DSN override. A missing config at a valid custom XDG root uses normal defaults without loading a HOME config. Existing working-directory fallbacks when both HOME and usable XDG settings are unavailable remain unchanged. Load still has no directory-creation side effects; first-run directory initialization remains tr-b9u.

Added 13 loader-level cases with distinct actual config files and exact DSN expectations: both custom roots, independent config/data roots, empty/unset values, relative values, mixed valid/invalid roots, paths with spaces, valid XDG roots without HOME, missing custom config, configured DSNs, environment-over-file/default DSNs and explicit config precedence. Nine cases failed before implementation; all 13 now pass.

README documents expanded POSIX shell locations, precedence and fallback behavior. AGENTS.md records path-resolution requirements. Absolute-path and empty-variable semantics were verified against the primary XDG specification: https://specifications.freedesktop.org/basedir/latest/.

## Validation

- go test ./config -count=1: PASS.
- go test -race ./... -skip '^TestFetchAndParseSlugs' -count=1 -timeout 60s: PASS.
- go vet ./... and go build: PASS.
- golangci-lint run --no-config --enable-only govet,staticcheck,unused,ineffassign --tests=false --timeout 2m ./...: 0 issues.
- git diff --check and beans check --json: PASS.

Known live-site crawler exclusions remain tracked in tr-4ek; the incompatible checked-in linter configuration remains tracked in tr-75b. The README's existing db_path sample correction remains the separate tr-zfh issue.

## Review status

Implementation and self-review are ready. Stopped for independent user review. This fix remains uncommitted and must not be pushed or followed by another fix until approved.

Independent review is tracked in tr-68b. Resume with review feedback before committing or starting another bugfix.

## Approved landing

Independent review tr-68b approved without refinements. The user confirmed all green and authorized this separate commit and push.

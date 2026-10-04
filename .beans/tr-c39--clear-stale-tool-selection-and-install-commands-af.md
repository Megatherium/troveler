---
# tr-c39
title: Clear stale tool selection and install commands after an empty search
status: completed
type: bug
priority: high
created_at: 2026-10-03T20:23:08Z
updated_at: 2026-10-04T06:35:57Z
parent: tr-twk
---

tui/update_handlers.go clears the tool list on an empty result but retains selectedTool, installs and both dependent panels. The previous install command remains actionable after a search returns no tools. Clear all dependent selection state together; also prevent stale commands when instruction loading fails.

## Acceptance criteria

Populate a real selection, then deliver empty search results and verify cleared selection, info, instructions and install actions. Cover instruction lookup failure after moving to another tool.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

Centralized TUI selection updates so stored instructions and the install panel are cleared before loading a new tool. Empty search results clear selectedTool, details, instructions, commands, fallback/language metadata and batch marks. Instruction lookup errors retain the new tool's metadata, keep commands disabled and report a contextual error. A successful lookup with no instructions keeps the install panel cleared. Install message handlers ignore queued normal/mise requests when the selection or its commands are cleared.

Added eight regression cases populated through a real SQLite database and the public Update flow. They cover nil/allocated empty results with and without batch marks, populated-search and cursor changes with lookup failure or no instructions, cleared info/install output, hotkeys from every active panel, direct panel action messages, queued messages after clearing commands, and recovery after a later populated search. All eight failed before the fix and now pass. Updated two existing install-message routing tests to populate actual commands and verify valid installs remain available.

## Validation

- go test ./tui ./tui/panels -count=1: PASS.
- go test -race ./... -skip '^TestFetchAndParseSlugs' -count=1 -timeout 60s: PASS.
- go vet ./... and go build: PASS.
- golangci-lint run --no-config --enable-only govet,staticcheck,unused,ineffassign --tests=false --timeout 2m ./...: 0 issues.
- git diff --check: PASS.

The two known live-site crawler tests remain excluded (tr-4ek); the incompatible checked-in lint configuration remains tracked in tr-75b. Tests inspect emitted command messages and queue state without executing fixture installs.

## Follow-ups

- tr-t9s: Reject queued install commands from superseded tool selections.
- tr-9qe: Keep batch-mark counts consistent with nonempty filtered searches.

These are separately tracked bugs with their own review/commit gates.

## Review status

Implementation and self-review are ready. Stopped for user review; no commit or push of this fix until approved.

Independent review is tracked in tr-drc. Resume with review feedback before committing or starting another bugfix.

## Approved landing

Independent review tr-drc approved the fix without refinements. The user confirmed "All green. Commit and continue." Landing the code, tests, documentation and affected Beans together in a separate Conventional Commit.

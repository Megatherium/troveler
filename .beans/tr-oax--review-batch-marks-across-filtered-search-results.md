---
# tr-oax
title: Review batch marks across filtered search results
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T07:11:17Z
updated_at: 2026-10-04T07:15:17Z
parent: tr-twk
---

Review tr-9qe before its separate commit. The user requires a pause after every bugfix for independent feedback.

## Scope

- tui/panels/tools.go: GetMarkedCount counts current-result membership, matching GetMarkedTools; API comments clarify mark retention and visible batch targets.
- tui/panels/tools_marks_test.go: meaningful mark/unmark/filter/restore/clear lifecycle, with exact IDs and counts.
- tui/filtered_marks_test.go: eight Model.Update cases covering Alt+I/Alt+M and hidden-only, single-tool fallback, mixed visible/hidden marks and filter restoration; batch cases verify displayed wizard counts, queued target IDs and progress state after completing the wizard. Separate empty-search case proves hidden marks never reappear afterward.
- README.md and one AGENTS.md lesson: document nonempty mark retention, visible counts/targets, empty-search clearing and single-tool fallback.

## Review focus

Verify displayed status and wizard counts, shortcut dispatch and actual targets agree. Confirm nonempty filtering retains hidden IDs, visible marks remain usable, restoring results restores retained marks, and empty results clear all marks. Check that hidden-only marks neither open an empty wizard nor block the selected tool's normal/mise command. Tests must never execute fixture install commands.

## Validation

Six of eight shortcut cases and the panel lifecycle failed before the change; all ten cases now pass. Focused TUI/panel tests, the full race suite with existing live-site exclusions, vet, build, core production lint (0 issues), whitespace and Beans integrity checks pass. Outstanding gates are tracked in tr-4ek and tr-75b.

AGENTS.md also contains the user's uncommitted Beans migration edits; only the new batch-mark lesson belongs to this fix. Other migration changes and unrelated untracked files remain outside scope.

- [x] Review the complete implementation and regression coverage.
- [x] Record approval or actionable feedback for tr-9qe.
- [x] Verify any required refinements before authorizing the separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Displayed status, wizard counts, shortcut dispatch and actual targets agree: the production change is exactly one method — GetMarkedCount now counts marks present in current results, matching GetMarkedTools. Grep-verified every consumer: status bar (view_panels), wizard count (ViewBatchConfig), Alt+I/M dispatch, batch targets, and per-row rendering all flow through the two now-consistent methods; no raw len(markedTools) reads remain anywhere.
- Nonempty filtering retains hidden IDs: SetTools keeps the marks map (comment updated to document retention); the panel lifecycle test proves a hidden mark survives filtering (IsMarked) and reappears in GetMarkedTools/GetMarkedCount when results are restored.
- Visible marks remain usable: marking and unmarking operate on visible rows while hidden marks persist (tested mid-sequence).
- Empty results clear all marks: the tr-c39 ClearMarks interplay is regression-tested end-to-end — hidden mark retained after a nonempty filter, cleared by an empty search, and never reappearing after results return.
- Hidden-only marks neither open an empty wizard nor block the selected tool: with zero visible marks, Alt+I/M fall through to the single-tool path; the test asserts no wizard, no config-active, no status-bar advertisement, and — when the current tool has instructions — the correct single install command is emitted (hidden marks do not block it).
- Batch path double-safety: startBatchInstall independently no-ops on zero visible targets.
- Test safety: the batch wizard is driven to completion and the emitted batchInstallStartMsg is inspected (target IDs, executing state, progress targets) but never delivered, so no ProcessTool runs and no fixture install command executes.
- Test quality: exact IDs, counts, status-bar text, wizard title count and mise mode asserted across 8 dispatch scenarios plus the empty-search case; all driven through real Model.Update key flows with a real SQLite database.

All state mutations remain on the Update thread; the change introduces no concurrency surface.

Gates independently reproduced: focused TUI/panels tests, go vet, go build, git diff --check, full go test -race ./... with the tr-4ek exclusions — all green.

tr-9qe is unblocked for its separate commit (code + beans together).

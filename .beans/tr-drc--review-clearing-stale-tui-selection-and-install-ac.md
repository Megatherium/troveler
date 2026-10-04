---
# tr-drc
title: Review clearing stale TUI selection and install actions
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T05:11:45Z
updated_at: 2026-10-04T06:35:57Z
parent: tr-twk
---

Review tr-c39 before its separate commit. The user requires a pause after each bugfix for independent review.

## Scope

- tui/update_handlers.go: one selection update path clears dependent state before instruction lookup, handles empty results, reports lookup errors, and ignores install messages when selection/commands are cleared. Empty results also clear batch marks.
- tui/panels/install.go: Clear resets commands, cursor, language and fallback state.
- tui/selection_state_test.go: eight real-database regression cases covering empty results, marked/unmarked tools, search/cursor selection changes, lookup failures, absent instructions, disabled global/direct/queued install actions, panel contents and populated-search recovery.
- tui/update_test.go: two positive install-routing tests now populate real commands and check valid installs still queue normally.
- README.md and one AGENTS.md lesson: empty-search and instruction-error behavior.
- Follow-up Beans tr-t9s/tr-9qe and epic inventory: separate queued-message identity and hidden-mark count issues, with no implementation in this fix.

## Review focus

Verify selection, instructions, info/install panels and fallback state clear together; failures show the new metadata and an error with no old commands; empty results cannot open batch configuration or old info; normal/mise hotkeys and queued messages remain disabled while commands are cleared; later populated results restore functionality. Confirm state mutations remain on Update's thread and fixture shell commands are never executed.

The queued-message guard deliberately checks the current availability of selection/commands. Selection identity after switching to another tool with valid commands is separately tracked in tr-t9s. Hidden batch marks after nonempty filtering are tracked in tr-9qe.

## Validation

Focused TUI/panel tests, full race suite with known live-site exclusions, go vet, build, production core lint (0 issues) and git diff --check pass. All eight new regression cases failed before the fix and now pass. Existing exclusions remain tr-4ek and tr-75b.

AGENTS.md contains the user's uncommitted Beans migration edits; only the new TUI-selection lesson belongs to this fix. Other migration changes and unrelated untracked files are outside scope.

- [x] Review the complete implementation and regression coverage.
- [x] Record approval or actionable feedback for tr-c39.
- [x] Verify any required refinements before authorizing the separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Clearing together: setSelectedTool is the single mutation path for selection state (verified no other assignment sites exist); it nils installs, clears the install panel, and clears the info panel for a nil tool before any load. InstallPanel.Clear is symmetric with SetTool (commands, cursor, toolLanguage, usedFallback); InfoPanel.Clear renders the empty prompt.
- Lookup failure path: new tool metadata shown, no old commands (install panel cleared before lookup), contextual m.err set and rendered through the existing error banner/modal.
- Empty results cannot act: marks cleared so Alt+I/M find neither marks nor commands; the info modal requires a non-nil selection; startBatchInstall no-ops with zero marked tools.
- Guard placement: the selectedTool+HasCommands check lives in handleInstallExecute/handleInstallExecuteMise — the single chokepoint all message sources flow through (global hotkeys, panel delegation, queued messages), so a message queued before the state cleared is inert. Selection identity after switching is correctly deferred to tr-t9s.
- Recovery: a later populated search rebuilds selection, instructions, and a working install command (asserted with exact fixture command equality).
- Update-thread discipline: instruction lookup is synchronous inside Update; no state mutated inside tea.Cmd goroutines (project lesson respected).
- Test safety: fixture commands are never executed — tests assert emitted command messages, modal flags, executing state, and queue behavior; the two updated routing tests check cmd non-nil without calling it.
- Test quality: eight cases drive the public Update flow against a real SQLite database with a deterministic fallback platform, covering nil vs allocated empty slices, marked and unmarked variants, search and cursor paths for both failure and zero-instruction tools, hotkeys from every panel, direct panel messages, and recovery. Non-vacuous throughout.

Non-blocking observation: m.err from a failed lookup persists until escape clears it — the same lifecycle shared by search and install errors (pre-existing pattern, not introduced by this fix).

Gates independently reproduced: focused TUI/panels tests, go vet, go build, git diff --check, full go test -race ./... with the tr-4ek exclusions — all green.

tr-c39 is unblocked for its separate commit (code + beans together).

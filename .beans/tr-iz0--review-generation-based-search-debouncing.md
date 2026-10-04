---
# tr-iz0
title: Review generation-based search debouncing
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T07:22:39Z
updated_at: 2026-10-04T07:54:24Z
parent: tr-twk
---

Review tr-lgx before its separate commit. The user requires a pause after every bugfix for independent feedback.

## Scope

- tui/panels/search.go: generation-tagged timer expiries, immutable query/version snapshots, Update-thread validation, duplicate expiry rejection, immediate Enter/Escape invalidation and queued-trigger validation.
- tui/update.go: timer messages route to the search panel regardless of the active panel.
- tui/update_handlers.go: reject superseded triggers before setting searching or scheduling database work.
- tui/update_keys.go: forward Escape as a real tea.KeyMsg; the previous keybinding object never cleared the input.
- tui/panels/search_debounce_test.go and tui/search_debounce_test.go: 13 behavioral cases covering ordered/reversed rapid edits, repeated query strings, Enter/Escape/Backspace clearing, interval timing, immutable delayed callback data, duplicate timer delivery, panel switch routing and obsolete triggers between timer expiry and database dispatch.
- tui/update_test.go: existing positive routing test obtains a real panel-produced versioned trigger.
- README.md and one AGENTS.md lesson: debounce timing, clear/Enter behavior and Update-thread routing guidance.

## Review focus

Confirm only the latest edit can dispatch a debounced search, including A-to-AB-to-A transitions and delayed background command evaluation. Enter/Escape must dispatch immediately and invalidate older timers; Backspace clearing must invalidate older queries through normal debounce. Timer expiry must work after focus moves. Verify expiry consumption does not duplicate searches and queued triggers are revalidated before database dispatch. Timer/trigger closures must not read mutable panel state. Check Escape routing through the actual Model.Update path, not only the panel API.

In-flight success/error ordering is intentionally assigned to the existing separate issue tr-nif; this fix concerns work not yet dispatched to the database.

## Validation

Nine original regression cases failed before implementation; all 13 cases now pass. Focused TUI/panels tests, full race suite with existing live-site exclusions, go vet, build, core production lint (0 issues), whitespace and Beans integrity checks pass. Existing exclusions/config repair remain tracked in tr-4ek and tr-75b.

AGENTS.md also contains the user's uncommitted Beans migration edits; only the new search-debounce lesson belongs to this fix. Other migration changes and unrelated untracked files remain outside scope.

- [x] Review the complete implementation and regression coverage.
- [x] Record approval or actionable feedback for tr-lgx.
- [x] Verify any required refinements before authorizing the separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Only the latest edit can dispatch: every edit, Enter and Escape increments the search generation; tea.Tick closures snapshot query and generation before running (no panel reads in the background); expiry validation requires the current generation, and consuming an expiry increments the generation so duplicate deliveries are inert. A-to-AB-to-A and repeated query strings are decided by generation, never string equality (tested both orders and both repeat forms).
- Enter/Escape dispatch immediately and invalidate pending timers; Backspace-to-empty flows through normal debounce with the empty query (tested for query state and dispatch sets).
- Timer expiry works after focus moves: SearchDebounceMsg routes through delegateToSearchPanel, which calls the panel Update directly regardless of active panel — verified by a test that tabs to the tools panel before delivering both timers.
- No duplicate searches: expiry is consumed once; immediate keys and later edits each invalidate the older timer/trigger.
- Queued triggers revalidated before database dispatch: handleSearchTriggered calls MatchesSearch before setting searching or scheduling performSearch — covering both the Enter-then-edit and expiry-then-edit windows (the TOCTOU cases), with the real performSearch executed against the in-memory database in tests.
- Closures read no mutable state: SearchTriggeredMsg and SearchDebounceMsg are immutable values built on the Update thread; generation fields are unexported so the Model can only validate via MatchesSearch.
- Escape routed through the real Model.Update path: the fix replaces the key.Binding passed as a message (which the panel's type switch could never match — a genuine latent bug) with a real tea.KeyMsg{Type: KeyEsc}; covered by the model-level Enter/Escape cases.
- Sole-emitter check: SearchTriggeredMsg is constructed only in triggerSearch, so no raw generation-zero messages exist in production; MatchesSearch's nonzero requirement additionally rejects any future legacy construction.

Test quality: 13 behavioral cases with a mini-runtime harness that feeds intermediate messages back through Update, executes real searches, and fails on searchErrorMsg; interval timing asserted with elapsed >= debounce; cursor blink set static for determinism. The stale gopls errors reference pre-change symbols; vet, build and tests compile clean.

Gates independently reproduced: focused debounce suites, go vet, go build, git diff --check, full go test -race ./... with the tr-4ek exclusions — all green.

tr-lgx is unblocked for its separate commit (code + beans together).

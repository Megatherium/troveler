---
# tr-uen
title: Review stale asynchronous search completion handling
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T08:03:50Z
updated_at: 2026-10-04T08:08:05Z
parent: tr-twk
---

Review tr-nif before its separate commit. The user requires a pause after every bugfix for independent feedback.

## Scope

- tui/model.go: register unique request/loading state before background execution; snapshot request/input generations and service; include request origins in both completion types; match only current pending work and unchanged input.
- tui/panels/search.go: expose the input/immediate-search generation to identify in-flight work.
- tui/update_handlers.go: reject obsolete completions before state mutations; consume current requests once; mark displayed search errors and clear only search-owned errors after a valid successful retry, preserving errors.Is behavior.
- tui/search_results_test.go: 24 behavioral cases with real DB searches and closed-DB failures; state snapshots include metadata/results, selection/cursor, instructions, command/install generation, info panel, visible marks, error and searching state.
- tui/selection_state_test.go, tui/filtered_marks_test.go and tui/update_test.go: fixture completions obtain origins from real search commands instead of generation-zero messages; closed-DB instruction-lookup fixture retains its intended failure path.
- README.md and one AGENTS.md lesson: latest request ownership, immediate input invalidation, retry error behavior and command snapshot guidance.

## Review focus

Confirm success and error handlers validate both unique request and unchanged input before changing any state. Repeated query strings, edits before the debounce fires, clearing input and startup work must not bypass the guard. Background commands must capture origins/service before running and avoid mutable model reads. Current results and failures must still work; duplicate completions must be inert. Successful retries must clear search errors without erasing unrelated errors, and retain error-chain behavior. Ensure older fixture tests remain meaningful under the origin guard.

## Validation

Sixteen original cases failed before implementation; all 24 new cases now pass. Focused TUI/panel tests, full race suite with existing live-site exclusions, vet, build, core production lint (0 issues), whitespace and Beans integrity checks pass. Existing exclusions/config repair remain tracked in tr-4ek and tr-75b.

AGENTS.md also contains the user's uncommitted Beans migration edits; only the new search-completion lesson belongs to this fix. Other migration changes and unrelated untracked files remain outside scope.

- [x] Review the complete implementation and regression coverage.
- [x] Record approval or actionable feedback for tr-nif.
- [x] Verify any required refinements before authorizing the separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Dual validation before any state mutation: both completion handlers call matchesSearchRequest first — requiring a nonzero unique request generation equal to the current one, an unchanged input generation from the panel, and searching still true. Repeated query strings are decided by generations, never string equality (proven with a refreshed-snapshot case where identical query strings must not regress newer metadata/instructions).
- No bypass paths: edits before the debounce fires, Escape clearing, backspace-to-empty, and round trips all invalidate in-flight work through the input generation (four edit variants x success/error tested); startup work flows through Init -> performSearch (generation 1, valid while input untouched, and rejected once the user types — verified both directions); origin-less messages are inert by the nonzero check, and searchResultMsg/searchErrorMsg have performSearch's closure as their sole constructor.
- Background capture: performSearch snapshots the request, the query, and the search service reference on the Update thread; the closure reads nothing from the model. The service-swap test executes an old command after the service was replaced and proves it still reports the captured (failing) service's error.
- Consumption and duplicates: searching=false on the first valid completion makes duplicate deliveries fail the guard; tested for both success and error completions, including preservation of a subsequent unrelated error.
- Error ownership: search failures wrap in searchFailure with Unwrap, so errors.Is/As and %v rendering keep working (asserted via errors.Is). A successful valid retry clears only searchFailure-owned errors; unrelated errors set after dispatch survive (tested both ways).
- Spinner lifecycle: rejected completions leave searching untouched (it belongs to the newer request); every dispatched request produces exactly one completion, so no stuck-loading path exists; traced the typing-during-startup-search race by hand.
- Older fixtures remain meaningful: all raw searchResultMsg deliveries were upgraded with searchResultWithOriginForTest (real origins from real commands), and the closed-DB error fixture keeps its genuine failure path.
- Test quality: 24 cases with a deep visibleSearchState snapshot (tools, selection, cursor, instructions, selected command, install generation via InstallRequest, info view, marks, searching, error) compared with reflect.DeepEqual around every stale delivery — non-vacuous throughout; real DB searches and closed-DB failures, no fixture command executed.

Gates independently reproduced: focused completion suites, go vet, go build, git diff --check, full go test -race ./... with the tr-4ek exclusions — all green. Stale gopls errors reference pre-change symbols; compilation and tests are clean.

tr-nif is unblocked for its separate commit (code + beans together).

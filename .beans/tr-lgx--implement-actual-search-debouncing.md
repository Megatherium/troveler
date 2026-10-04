---
# tr-lgx
title: Implement actual search debouncing
status: completed
type: bug
priority: normal
created_at: 2026-10-03T20:23:08Z
updated_at: 2026-10-04T07:54:24Z
parent: tr-twk
---

tui/panels/search.go never assigns searchTimer and schedules an independent sleep command for every input change. Typing ab still emits the obsolete search for a. Use cancellable timers or generation-tagged messages to launch only the newest debounced search.

## Acceptance criteria

Simulate rapid input changes and verify that only the latest query launches a search after the debounce interval. Ensure Enter triggers an immediate search and clearing the input invalidates pending searches.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

Replaced the unused searchTimer and per-edit sleep/search commands with generation-tagged debounce expiry messages. Queries and generations are captured before background execution. SearchPanel.Update drops superseded expiries and consumes a current timer once, then emits a versioned immutable search trigger. Input changes, Enter and Escape advance the generation, so repeated query strings cannot revive old timers. Model validates queued triggers before changing searching state or scheduling database work.

Model.Update routes timer messages to the search panel regardless of focus; switching panels does not discard the latest pending search. Corrected Escape delegation to forward a tea.KeyMsg rather than a keybinding object, allowing input clearing and immediate empty search to invalidate pending work end-to-end.

README documents the 150 ms interval, immediate Enter/Escape, Backspace clearing, repeated-query invalidation and panel switching. AGENTS.md records the Update-thread and routing requirements. The existing positive routing test now uses a real panel-produced trigger with a valid origin.

Added 13 meaningful cases across panel and model tests: rapid A-to-AB-to-A changes with ordered/reversed timer delivery; Enter/Escape/Backspace invalidation; interval timing; immutable delayed callbacks and duplicate expiry rejection; actual database dispatch after a panel switch; immediate Enter/Escape without a delayed duplicate; stale queued Enter with distinct/repeated queries; and a debounce expiry accepted before a subsequent edit invalidates its queued trigger. Nine original regression cases failed before implementation. All cases now pass. Actual database commands run in model tests, but completions are not applied so they do not conflate dispatch with in-flight ordering.

## Validation

- go test ./tui ./tui/panels -count=1: PASS.
- go test -race ./... -skip '^TestFetchAndParseSlugs' -count=1 -timeout 60s: PASS.
- go vet ./... and go build: PASS.
- golangci-lint run --no-config --enable-only govet,staticcheck,unused,ineffassign --tests=false --timeout 2m ./...: 0 issues.
- git diff --check and beans check --json: PASS.

Known live-site crawler exclusions remain tracked in tr-4ek; the incompatible checked-in lint configuration remains tracked in tr-75b. Rejecting out-of-order successes/errors from searches already launched remains the separate bug tr-nif.

## Review status

Implementation and self-review are ready. Stopped for independent user review. This fix remains uncommitted and must not be pushed or followed by another fix until approved.

Independent review is tracked in tr-iz0. Resume with review feedback before committing or starting another bugfix.

## Approved landing

Independent review tr-iz0 approved without refinements. The user confirmed all green and authorized the separate commit and push.

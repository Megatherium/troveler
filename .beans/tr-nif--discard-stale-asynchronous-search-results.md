---
# tr-nif
title: Discard stale asynchronous search results
status: completed
type: bug
priority: normal
created_at: 2026-10-03T20:23:08Z
updated_at: 2026-10-04T08:08:05Z
parent: tr-twk
---

tui/model.go carries query in searchResultMsg, but tui/update_handlers.go ignores it and accepts every completion. Older searches can replace newer results or resurrect a cleared query. Track a request generation and ignore obsolete successes and errors.

## Acceptance criteria

Deliver search completions in reversed order, including repeated query strings, clearing input and stale errors. Verify the latest request controls results and searching/error state.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

Every database search, including Init, now registers a unique request generation and loading state on the main thread before returning a command. Commands capture their query, request identity, input generation and service; background work returns those immutable origins on both success and error messages without reading mutable model fields.

Both completion handlers validate the current pending request and unchanged search-panel generation before any state mutation. Input edits, Enter/Escape and repeated query transitions immediately obsolete older completions, even before another database request launches. Startup requests share the same guard. Completed requests are consumed once, rejecting duplicate results/errors.

Search failures retain their error chain in a small ownership wrapper. A valid successful retry clears only a search-owned error, preserving errors reported by other actions. Existing selection, install-command and batch-mark fixtures now obtain origins from actual search commands so the new guards do not make their tests vacuous. README and AGENTS.md document the behavior and origin requirements.

Added 24 regression/control cases using real database search commands and real closed-database failures: older success/error delivery before/after the newest result; identical-query metadata snapshots; edits, Escape, Backspace and A-to-AB-to-A transitions before the next dispatch; latest-error ownership; successful retries with/without unrelated errors; stale/current startup completions; delayed execution after a newer request is queued; immutable service snapshots; missing-origin messages; and duplicate success/error delivery. State snapshots assert results, selection/cursor, instructions, command, install generation, info panel, visible marks, error and loading state stay unchanged for obsolete responses. Sixteen original cases failed before the fix; all 24 now pass.

## Validation

- go test ./tui ./tui/panels -count=1: PASS.
- go test -race ./... -skip '^TestFetchAndParseSlugs' -count=1 -timeout 60s: PASS.
- go vet ./... and go build: PASS.
- golangci-lint run --no-config --enable-only govet,staticcheck,unused,ineffassign --tests=false --timeout 2m ./...: 0 issues.
- git diff --check and beans check --json: PASS.

Known live-site crawler exclusions remain tracked in tr-4ek; the incompatible checked-in linter configuration remains tracked in tr-75b.

## Review status

Implementation and self-review are ready. Stopped for independent user review. This fix remains uncommitted and must not be pushed or followed by another fix until approved.

Independent review is tracked in tr-uen. Resume with review feedback before committing or starting another bugfix.

## Approved landing

Independent review tr-uen approved without refinements. The user confirmed all green and authorized the separate commit and push.

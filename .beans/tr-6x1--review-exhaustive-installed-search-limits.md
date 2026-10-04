---
# tr-6x1
title: Review exhaustive installed-search limits
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T04:37:34Z
updated_at: 2026-10-04T05:01:29Z
parent: tr-twk
---

Review the tr-3bk search-limit bugfix before its separate commit. The user requires a stop after every bugfix for independent feedback.

## Scope

- db/sqlite_search.go: exhaustive bounded candidate batches with continuation on the case-insensitive sort key and ID; existing Boolean match flags retained; final limit counts matching tools; install queries run after closing candidate rows; PATH cache persists across batches.
- db/models.go: explicit nonpositive-limit contract for direct database searches.
- db/search_limit_integration_test.go: 17 limit/order cases across 2,000 tools plus a 100,001-tool cap regression and cancellation check, using an isolated PATH executable fixture.
- db/sqlite_search_test.go removed: obsolete overfetch formula tests replaced by behavioral integration coverage.
- README.md and one new AGENTS.md lesson: matching limits, CLI default, and database batching contract.

## Review focus

Check sort/continuation collation and tie-breakers for ASC, DESC and equal keys, full Boolean/OR selection in every batch, parameter order on subsequent queries, large and nonpositive limits, termination on exhausted candidates, database cursor release, cross-batch PATH caching and context/error propagation. No schema or CLI-default changes.

## Validation

Full race suite passes with the existing live-site crawler exclusions (tr-4ek). go vet, build, core production lint (0 issues), git diff --check and beans check pass. The checked-in linter configuration remains a separate tr-75b issue. Twelve regression cases failed before this fix; all new cases and the previous 23 Boolean-filter cases pass now. The race-enabled db package, including the 100,001-tool fixture, took 5.369s.

AGENTS.md also contains the user's uncommitted Beans migration edits. Only the newly added installed-filter batching lesson belongs to this fix; other migration files and unrelated untracked files are outside scope.

- [x] Review the complete implementation and regression coverage.
- [x] Record approval or actionable feedback for tr-3bk.
- [x] Verify any required refinements before authorizing the separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Sort/continuation collation and tie-breakers: ORDER BY (`search_sort_value COLLATE NOCASE <dir>, id ASC`) and the cursor predicate (`NOCASE > ?`/`< ?` OR `NOCASE = ? AND id > ?`) use identical collation and tie-break direction for ASC and DESC; strict comparisons make skipped or duplicated rows across batch boundaries impossible. The equal-key cases (2000 tools with identical taglines, alternating Rust/rust languages, identical dates) are exactly the boundary traps and pass with hand-verified expected ID sequences.
- Full Boolean/OR selection in every batch: both CASE match flags are present in every paginated query; OR cases assert exact match sets in both sort directions and the final limit counts matches, not candidates.
- Parameter order on subsequent queries: placeholder text order (installedClause, uninstalledClause, whereClause, cursorClause, LIMIT) matches pageArgs assembly; args are re-copied per iteration, so no cross-iteration slice aliasing.
- Large and nonpositive limits: nonpositive maps to SQLite -1 (unlimited); the 100,001-tool case walks 201 batches and finds the single match beyond the old cap; limit 1500 with installed=false exercises multi-batch Go-side filtering and exhaustion.
- Termination: empty batch returns, partial batch (< batchSize) returns, exact-multiple sizes terminate via one final empty batch. No infinite-loop path found.
- Cursor release: candidates are materialized and rows closed before install-instruction batch queries — required by the single-connection database; the suite would deadlock otherwise.
- Cross-batch PATH caching and context/error propagation: shared cache skips empty names consistently with IsInstalledCached; ctx checked per batch and per candidate; cancellation test asserts context.Canceled.
- No schema or CLI-default changes: the --limit flag default, help text, and the 50 default are untouched and applied upstream (internal/search/search.go applies 50 when limit <= 0; commands/newest.go uses an explicit 1000), so the new 0-means-unlimited database contract is unreachable from CLI and TUI. README documents the contract accurately.
- Deleted db/sqlite_search_test.go contained only the six overfetch formula tests; behavioral sort and installed-limit coverage remains in search_integration_test.go and passes.

Non-blocking note: SQL NULL sort values (impossible via app write paths, which bind Go zero-value strings as '') would stall cursor continuation — theoretical, only reachable by direct external SQL inserts.

Gates independently reproduced: go vet ./..., go build, git diff --check, focused limit/Boolean suites, full go test -race ./... (db package 6.8s including the 100,001-tool fixture) — all green. Stale gopls diagnostics reference pre-change files; compilation and tests are clean.

tr-3bk is unblocked for its separate commit (code + beans together).

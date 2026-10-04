---
# tr-3bk
title: Apply installed-search limits after filtering all candidates
status: completed
type: bug
priority: high
created_at: 2026-10-03T20:23:08Z
updated_at: 2026-10-04T05:01:29Z
parent: tr-twk
---

db/sqlite_search.go uses a fixed over-fetch factor and cap. Matching installed tools beyond the fetched prefix disappear, and an installed OR branch can exceed the requested limit. Return the first requested number of matches in sort order without relying on a guessed installed ratio.

## Acceptance criteria

Seed more nonmatching tools than the current over-fetch prefix, with matches later in the ordering. Verify positive limits, descending order, OR filters, and documented unlimited behavior.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Scope clarification after tr-w1q

The Boolean-expression fix also removes the OR bypass from final trimming, so OR searches obey the requested output limit. The guessed candidate prefix and capped unlimited behavior remain unfixed; retain this bean for correct candidate exhaustion/pagination and the existing acceptance criteria.

## Summary of Changes

Replaced guessed installed-search overfetch limits with 500-candidate batches. Search continues in SQLite sort order until enough full-expression matches are found or candidates are exhausted. Continuation uses the case-insensitive sort value plus an ascending tool-ID tie-breaker for both ASC and DESC sorts. Candidate rows are released before install-instruction batch queries to preserve the single-connection database behavior. PATH availability is cached across all batches in the search.

Removed the overfetch multiplier, maximum and 100,000-row fallback and their implementation-only tests. Direct database searches now consistently treat zero or negative limits as unlimited, including searches without an installed predicate. The search service still supplies the existing default of 50. Documented the database contract, CLI limits and agent batching guidance.

Added deterministic integration coverage with isolated PATH fixtures: 17 exact-ID limit/order cases over 2,000 tools, a 100,001-tool regression beyond the old unlimited cap, and cancellation propagation. Twelve cases failed before the fix; all new cases and the prior 23 Boolean regression cases now pass.

## Validation

- go test -race ./... -skip '^TestFetchAndParseSlugs' -count=1 -timeout 60s: PASS (db package 5.369s, including the 100,001-tool test).
- go vet ./...: PASS.
- go build: PASS.
- golangci-lint run --no-config --enable-only govet,staticcheck,unused,ineffassign --tests=false --timeout 2m ./...: 0 issues.
- git diff --check and beans check --json: PASS.

Known exclusions remain unchanged: HTTP-dependent crawler tests (tr-4ek) and the incompatible checked-in linter configuration (tr-75b).

## Review status

Implementation and self-review are ready. Stopped for user review; this fix is uncommitted and must not be pushed or followed by another fix until approved.

Independent review is tracked in tr-6x1. Resume with review feedback before committing or starting the next bugfix.

## Approved landing

Independent review tr-6x1 approved the fix with no refinements required. The user confirmed "All green. Commit and continue." Code, tests, documentation and affected Beans are being landed in a separate Conventional Commit.

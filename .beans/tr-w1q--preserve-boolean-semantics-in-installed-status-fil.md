---
# tr-w1q
title: Preserve Boolean semantics in installed-status filters
status: completed
type: bug
priority: high
created_at: 2026-10-03T20:23:08Z
updated_at: 2026-10-04T04:30:09Z
parent: tr-twk
---

db/filter_sql.go converts installed filters to 1=1, including under NOT; db/sqlite_search.go skips runtime filtering whenever installed occurs under OR. Consequently installed=true|language=go includes unrelated tools, and !installed=true returns no results. Evaluate the complete expression with runtime installed status, preserving AND, OR and NOT semantics.

## Acceptance criteria

Cover installed-only negation, OR with a non-installed field, nested expressions, repeated and contradictory installed predicates, and combinations with tags. Assert exact matching tool IDs.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

Preserve the full filter AST by evaluating SQL match flags for both possible installed states, keeping candidates matching either state, then choosing the correct flag after the runtime PATH check. Text, case-insensitive LIKE, wildcard, and tag predicates remain in SQLite. Installed field names are recognized case-insensitively. Removed first-leaf extraction and OR-bypass helpers and their implementation-only tests. Updated README examples and AGENTS.md guidance.

Added 23 deterministic parser-to-database regression cases asserting exact IDs and installed status, with an isolated temporary PATH. Before the fix, 13 cases failed; all now pass. Cases cover negation, OR in both operand orders, nested AND/OR/NOT with tags, contradictory/repeated installed leaves, true/false and 1/0 values, uppercase field names, missing install instructions, SQLite wildcard/case behavior, and combined free-text search.

## Validation

- go test -race ./... -skip '^TestFetchAndParseSlugs' -count=1 -timeout 60s: PASS.
- go vet ./...: PASS.
- go build: PASS.
- golangci-lint run --no-config --enable-only govet,staticcheck,unused,ineffassign --tests=false --timeout 2m ./...: 0 issues.
- git diff --check: PASS.

The known live-site crawler tests remain excluded because they depend on an HTTP 403 response (tr-4ek); the checked-in linter configuration remains incompatible with v2 (tr-75b). The fixed candidate over-fetch remains for the separate tr-3bk fix. OR expressions now obey final result trimming as part of applying their Boolean predicate.

## Review status

Implementation and self-review are ready. Paused for user review; no commit or push of this fix until approved.

Independent review is tracked in tr-vks. Resume with review feedback before committing or beginning another fix.

## Approved landing

Independent review tr-vks approved the fix with no refinements required; the user confirmed "All green. Commit and continue." Code, regression tests, documentation and affected Beans are being landed together in a separate Conventional Commit.

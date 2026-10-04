---
# tr-vks
title: Review Boolean installed-status filter semantics
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-03T20:47:18Z
updated_at: 2026-10-04T04:30:09Z
parent: tr-twk
---

Review tr-w1q before commit or push. The user requires a pause after every bugfix for independent review.

## Scope

- db/filter_sql.go: full-AST predicates for both installed states; safe SQL candidate union under NOT; case-insensitive installed field recognition.
- db/sqlite_search.go: SQL match flags and runtime selection of the actual state; removed OR bypass and first-installed-leaf filtering.
- db/search_installed_boolean_test.go: 23 deterministic parser-to-database cases, isolated executable fixture and PATH, exact matching IDs and installed flags.
- db/filter_sql_test.go and db/sqlite_search_test.go: removed tests of obsolete helpers; existing static SQL and overfetch tests retained.
- README.md and one new AGENTS.md lesson: Boolean semantics and examples.

## Validation

Race-enabled suite passes with the two known live-site crawler tests excluded (tr-4ek). go vet, build, production core linter checks and diff whitespace checks pass. Checked-in golangci-lint configuration remains a separate tr-75b issue.

## Review focus

Check full Boolean evaluation with repeated/contradictory installed predicates and nested negation, bound SQL argument order, tag subquery scope, and preservation of SQLite LIKE semantics. Confirm no background/state or schema changes. The overfetch prefix can still hide later matches; that is separately tracked in tr-3bk.

AGENTS.md also contains the user's uncommitted Beans migration edits; only the new installed-filter lesson belongs to this fix. Other migration files and unrelated untracked files are outside this review.

- [x] Review the complete implementation and regression coverage.
- [x] Record approval or actionable feedback for tr-w1q.
- [x] Verify any required refinements before authorizing the separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Full Boolean semantics: the dual `CASE WHEN expr[state]` design evaluates the complete AST in SQLite for both installed states, keeps the union as candidates, and Go selects the flag for the actual PATH-resolved state. Exact AND/OR/NOT semantics by construction; hand-traced all 23 fixture expectations (contradictions to empty set, double negation, De Morgan, repeated predicates, both OR operand orders, tag mixes, free text + filter) against the fixture table before running them.
- Bound SQL argument order: placeholders appear in text as installedClause, uninstalledClause, whereClause, LIMIT; args are appended in exactly that order, and the recursive builder emits left-then-right args matching its parenthesized text. Any misalignment would fail the mixed-arg cases (tagline=FIXTURE_b, tag=CLI) — they pass.
- Tag subquery scope: the correlated EXISTS still binds to the inner select's tools row; duplication inside the two CASE expressions does not change scoping.
- LIKE semantics: wildcards (name=%) and ASCII case-insensitivity (tagline=FIXTURE_b, INSTALLED=true) survive the CASE duplication; buildFieldFilter is untouched.
- No schema, background, or state changes: only db/filter_sql.go, db/sqlite_search.go, and tests.
- Removed tests are exclusively implementation-only tests of the deleted helpers (hasInstalledInOrContext, getInstalledFilterInfo, installedFilterValue) — correct removal per project lessons; static SQL and overfetch behavioral tests retained and passing.

Hazards specifically hunted and cleared:
- Removed empty-WHERE fallback in Search: safe because BuildWhereClause has its own 1=1 fallback (filter_sql.go:41), so installed-only filters with no search term produce valid SQL.
- Installed-only filter with no args: CASE constants 1=1/1=0 produce placeholder-free flags; arg counts stay consistent.
- Stale gopls "undefined" errors at pre-diff line numbers: go vet, build, and the race suite compile clean; LSP cache artifact only.
- buildFilterSQL's candidate-union branch is now unreachable from Search (filter nulled before BuildWhereClause) and only reachable via external callers of the exported BuildWhereClause, whose doc comment warns installed filters need runtime evaluation. Acceptable.

Non-blocking notes: installed=<other values> parses as false (pre-existing true|1 parsing, unchanged); the overfetch prefix hiding later matches remains tracked in tr-3bk as instructed.

Gates independently reproduced: go vet ./..., go build, git diff --check, focused Boolean suite (23/23), full go test -race ./... with the tr-4ek live-site exclusions — all green.

tr-w1q is unblocked for its separate commit (code + beans together).

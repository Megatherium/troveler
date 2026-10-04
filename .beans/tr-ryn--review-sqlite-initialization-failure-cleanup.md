---
# tr-ryn
title: Review SQLite initialization failure cleanup
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T15:12:29Z
updated_at: 2026-10-04T15:16:18Z
parent: tr-twk
---

Review tr-15y before its separate bugfix commit. The user requires a stop after every fix.

## Scope

db/sqlite.go: New immediately transfers its opened handle into initializeSQLite; deferred failure cleanup preserves initialization and close error causes with errors.Join. New db/sqlite_init_test.go: a local connector wraps real SQLite for stage failures and ownership checks. README: three lines about initialization cleanup. AGENTS.md: exactly one new db.New ownership lesson. Eventual commit must include tr-15y and this review Bean and preserve unrelated migration edits and untracked files.

## Review focus

Confirm every failure after sql.Open closes the handle, while successful initialization leaves it open for SQLiteDB.Close. Named return error assignments must reach the deferred guard despite local err variables. Initialization error prefixes and errors.Is must be preserved, including simultaneous close errors. No database schema, migration ordering, connection limit or API behavior beyond cleanup should change.

Reproduce eight failure stages: connection open, ping, foreign-key setup, first table, final index, first and second migration query, and migration ALTER. Verify no retained connections, exactly one physical close where a connection exists, and unusable sql.DB after removing the fault. Successful setup must retain an open connection, enable foreign keys, allow schema/migrated-column writes, and close once upon caller cleanup. Connector must be local to tests without global sql driver replacement or production callbacks.

## Validation

Seven open-connection failures and cleanup-error preservation failed before adding the guard. Targeted regressions and full go test -race ./... -count=1 -timeout 60s now pass, including real SQLite and integration runner tests, with module downloads off and external HTTP proxies refused. Vet, build, config verification, whitespace and Beans integrity pass. Full configured lint still fails on the existing tr-dvn baseline: 248 uncapped diagnostics; comparison by linter, message and filename shows no additions/removals, and the new test has no findings. Report: /tmp/troveler-15y-lint.json.

- [x] Review implementation, real SQLite fixtures and documentation.
- [x] Reproduce relevant quality gates and check existing lint baseline.
- [x] Record approval or actionable refinement feedback.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Ownership and cleanup: New transfers the opened handle into initializeSQLite immediately; the named-return (sqlite, err) deferred guard closes only when err != nil, so success leaves the handle exclusively to SQLiteDB.Close. Each inner failure returns an explicit value (assigning the named err before the guard runs), which is the correct handling of the shadowed-local-err trap — verified by every failure stage actually closing.
- Error preservation: errors.Join(err, db.Close()) keeps the initialization cause first; when Close succeeds the message is unchanged, when Close fails both causes remain reachable via errors.Is (dedicated test with injected ping+close failures asserting both, with exactly one physical close). Prefix checks in the battery confirm "failed to ping db", "failed to enable foreign keys", "failed to create tables" and "failed to run migrations" survive; the WithDB regression suite (which string-matches the ping prefix) still passes.
- Eight failure stages reproduced: connection open (no connection exists — zero physical closes, but the handle itself is closed), ping, foreign-key pragma, first table, final index, both pragma_table_info migration queries, and the ALTER — each asserting nil database, wrapped cause, stage prefix, exactly one physical close where a connection exists, zero open connections, and (after clearing the fault) that the handle remains unusable, proving the *sql.DB was closed rather than merely disconnected.
- Success path: zero closes during initialization, foreign_keys pragma reads 1, schema and migrated-column writes succeed, and exactly one physical close occurs on caller Close.
- Blast radius: no schema, migration ordering, connection-limit or API changes — the diff is the ownership transfer plus the guard; SetMaxOpenConns placement unchanged. The test connector is local (sql.OpenDB with a wrapper around real *sqlite3.SQLiteConn) — no global driver registration or production hooks.
- Gates reproduced: focused init tests and WithDB regressions with -race, full hermetic race suite (proxies refused, downloads off, integration runner included), go vet, build, git diff --check; the new test file contributes zero lint findings (db package findings are pre-existing tr-dvn baseline).

tr-15y is unblocked for its separate commit (db/sqlite.go + db/sqlite_init_test.go + README + one AGENTS.md lesson + tr-15y/tr-ryn beans).

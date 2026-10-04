---
# tr-15y
title: Close SQLite handles when initialization fails
status: completed
type: bug
priority: normal
created_at: 2026-10-04T13:28:09Z
updated_at: 2026-10-04T15:16:18Z
parent: tr-twk
---

db.New in db/sqlite.go opens a sql.DB but returns from Ping, foreign-key setup, createTables and migration errors without closing it. Failed initialization leaves the connection-opener goroutine running; failures after a successful Ping can also retain an open SQLite connection. This is separate from tr-b9u default-directory preparation. Close the newly opened handle on every initialization failure while preserving successful ownership and wrapped errors. Add meaningful failure-path coverage, run relevant gates, update behavior documentation if needed, stop for independent review, and commit separately with this bean reference.

## Acceptance checklist

- [x] Close the newly opened handle on every initialization failure and preserve error causes.
- [x] Verify failure cleanup and successful ownership with meaningful regression tests.
- [x] Run quality gates and document database initialization ownership.
- [x] Stop for independent review before committing.
- [x] Commit and push separately after approval, including this Bean.

## Summary of Changes

db.New immediately hands its opened sql.DB to a private initializeSQLite helper. The helper owns the handle during setup and closes it through a deferred error guard on every unsuccessful return. errors.Join preserves the wrapped initialization cause and any cleanup error. Successful setup transfers ownership to SQLiteDB.Close; connection limits, foreign keys, schema and migrations remain unchanged. README and one AGENTS.md lesson explain failure cleanup and successful ownership.

## Regression coverage

A test-local sql.OpenDB connector wraps real in-memory SQLite, avoiding global driver replacement or test-only production callbacks. Eight deterministic failures cover physical connection open, Ping, foreign-key setup, first schema table, final schema index, both migration queries and migration ALTER. Every failure must return no database, preserve the expected error stage and cause, retain zero connections, close any opened physical connection exactly once, and leave the sql.DB unusable after the injected fault is removed. A cleanup failure must preserve both error causes. Success must remain open, enable foreign keys, accept writes to the schema and migrated column, and close exactly once when the caller requests it.

Before adding cleanup, the seven open-connection failure cases reproduced zero closes, one retained connection and a still-usable handle; the cleanup-error preservation test also failed. All cases pass after repair, including the connection-open failure where no physical connection exists.

## Validation

Full go test -race ./... -count=1 -timeout 60s passes with module downloads disabled and external HTTP proxies refused. go vet ./..., build, whitespace and Beans integrity pass. Full configured lint still fails on the existing 248 uncapped tr-dvn findings, with no added or removed diagnostic signatures and no new test-file findings.

## Review status

Self-review completed. Stopped with this fix uncommitted for independent review; do not land or start another fix until feedback.

Independent review is tracked in tr-ryn. Await review feedback before committing this fix or starting another issue.

Independent review tr-ryn approved without refinements; user authorized the separate commit and push with All green. Commit and continue.

---
# tr-rfx
title: 'Review: preserve tool identity and atomically refresh installs'
status: completed
type: task
priority: high
tags:
    - review
created_at: 2026-10-03T20:28:51Z
updated_at: 2026-10-03T20:39:30Z
parent: tr-twk
---

Review the isolated fix for tr-fxj before it is committed or work begins on the next bug.

Changes: UpsertTool returns the existing stored ID for a known slug. SaveToolSnapshot atomically updates metadata and replaces instructions while preserving tags. Both CLI and TUI persistence paths use it and return database write failures. Documentation is updated.

Files: db/sqlite_tools.go, db/sqlite_refresh.go, db/sqlite_refresh_test.go, commands/update.go, commands/update_persistence_test.go, internal/update/service.go, internal/update/service_persistence_test.go, README.md, and the new crawl-persistence lesson in AGENTS.md. Existing Beans migration changes and unrelated untracked files are outside this fix.

Validation passed:
- Focused persistence regression tests in db, commands and internal/update.
- go test -race ./... -skip '^TestFetchAndParseSlugs' -count=1 -timeout 60s.
- go vet ./... and go build.
- Core production lint (govet, staticcheck, unused, ineffassign): 0 issues.
- git diff --check.

Known baseline limitations: the two skipped live-site tests previously returned HTTP 403; the invalid project linter configuration is tracked separately in tr-75b.

- [x] Review identity preservation, replacement semantics, rollback and failure propagation.
- [x] Provide approval or concrete refinement feedback.

Independent review and user approval are recorded below; the approved fix is ready to commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.** Independent verification, not a rubber stamp.

Verified correct:
- Identity: bare `ON CONFLICT DO UPDATE` + `RETURNING id` resolves slug conflicts to the stored ID while keeping the old update-by-id slug-change semantics. mattn/go-sqlite3 v1.14.33 bundles SQLite >= 3.47, so both bare-target upsert and RETURNING (both 3.35+) are safe.
- Replacement: instructions deleted and re-inserted by the stored ID inside one tx; obsolete commands removed; nil installs wipe correctly (tested).
- Tags: tool rows are updated in place, never deleted, so tool_tags (ON DELETE CASCADE) survives; verified at db, CLI, and TUI layers. ReapplyTags (slug-based, INSERT OR IGNORE) unaffected.
- Rollback: deferred tx.Rollback; caller's Tool mutated only after Commit; regression test covers metadata + instruction rollback and cross-tool non-interference.
- Failure propagation: CLI writeErr read after close(detailDone) (happens-before via channel close, race-clean); TUI emits an error ProgressUpdate and returns; workers select on ctx.Done so defer cancel() prevents goroutine leaks. FK pragma satisfied within the tx (tool upsert precedes instruction inserts).
- Gates independently reproduced: go vet, go build, git diff --check, focused tests and full `go test -race ./... -skip '^TestFetchAndParseSlugs'` all green.

Non-blocking observations (record for follow-up tickets, do NOT hold this fix):
- `UpsertInstallInstruction` is now production-dead (only tui test helpers use it) — remove during tr-519 pipeline consolidation.
- `executable_name` has no production writer; snapshot replacement would silently clear it if one appears — revisit when tr-799 lands.
- Pre-existing, out of scope: runUpdate's `errChan` nil-error path and the ctx.Done/writeErr select race; possibly covered by tr-nzs.

tr-fxj is unblocked to commit (review passed).

## Summary of Changes

User approved the fix with: All green. Commit and continue. Review passed; the fix may be committed and pushed.

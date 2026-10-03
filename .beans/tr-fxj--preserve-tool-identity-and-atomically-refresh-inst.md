---
# tr-fxj
title: Preserve tool identity and atomically refresh install instructions
status: completed
type: bug
priority: high
created_at: 2026-10-03T20:23:07Z
updated_at: 2026-10-03T20:39:30Z
parent: tr-twk
blocked_by:
    - tr-rfx
---

crawler/parser.go assigns a new UUID on every crawl. db/sqlite_tools.go upserts only on id while slug is unique, so refreshing an existing tool fails with UNIQUE constraint failed: tools.slug. CLI and TUI suppress the failure. Preserve the existing database identity when refreshing a slug and replace tool metadata and install instructions atomically, keeping tags and avoiding stale or duplicated instructions.

## Acceptance criteria

Re-crawl an existing slug with changed metadata and commands; verify the original tool ID, updated values, preserved tags, no duplicate instructions, removal of obsolete commands, and rollback on a failed write. Exercise both CLI and TUI persistence paths.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

Preserves existing tool IDs during refresh and saves metadata plus the latest instructions atomically. CLI and TUI surface database write errors. Added regression coverage for existing slugs, instruction replacement, tag preservation and transactional rollback. Checks pass; awaiting user review before committing or starting another bug.

## Review outcome

Approved by independent review tr-rfx and by the user. Landing as a separate fix(update) commit with the regression tests and ticket references.

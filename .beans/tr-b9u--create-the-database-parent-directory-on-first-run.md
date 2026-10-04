---
# tr-b9u
title: Create the database parent directory on first run
status: completed
type: bug
priority: normal
created_at: 2026-10-03T20:23:09Z
updated_at: 2026-10-04T13:32:59Z
parent: tr-twk
---

Neither config/config.go nor db initialization creates the default troveler data directory. With a fresh home/data directory, opening the default DSN fails with unable to open database file. Create the application data directory during initialization with appropriate permissions and report failures clearly.

## Acceptance criteria

Initialize from an empty temporary data directory and successfully open/use the database. Preserve in-memory and explicit DSN behavior and cover directory creation failures.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

Config.Load captures the selected default DSN and its raw parent directory without filesystem writes. Config.EnsureDatabaseDir creates that directory with MkdirAll(0700) only while the loaded default DSN remains selected. WithDB invokes preparation before SQLite initialization; every CLI command and TUI launch uses this helper. Existing permissions remain unchanged, explicit/config/environment/manual DSNs bypass directory preparation, and DSN changes after Load are respected. Preparation uses the originally selected location even if XDG changes afterward. Filesystem failures include the path and preserve the wrapped os.PathError before any callback.

README and AGENTS.md document first-run creation and explicit DSN ownership. Regression tests exercise real WithDB and SQLite: custom nested XDG roots with spaces, HOME/relative-XDG fallbacks, XDG without HOME, existing permissions, saved tools surviving reopen, config/environment/memory-URI/file/manual/changed DSNs and unchanged missing explicit parents. Four fresh-path cases plus the blocked-directory diagnostic failed before implementation. All now pass.

## Validation

- Fifteen command-level regression cases pass, including saved tools surviving a close/reopen.
- Four fresh-path cases and the directory-error case failed before implementation.
- Self-review caught an explicit environment DSN equal to the default path; its new regression failed before clearing default provenance for environment overrides, and passes now.
- Full race suite with TestFetchAndParseSlugs excluded: PASS.
- go vet ./... and go build: PASS.
- Core production golangci-lint (govet/staticcheck/unused/ineffassign): 0 issues.
- git diff --check and beans check --json: PASS.
- Built CLI search from an empty temporary XDG data root: PASS; database created under a 0700 application directory.

Known live-site crawler exclusions remain tr-4ek and the checked-in linter configuration repair remains tr-75b. Existing README db_path correction remains tr-zfh. Newly observed db.New failure-path handle cleanup is filed separately as tr-15y; it was not folded into this fix.

## Review status

Self-review is finished. Stopped for independent review, with the implementation uncommitted. Do not proceed with another bugfix or push this fix until the user approves.

Independent review is tracked in tr-3ir. Resume with review feedback before committing or starting another fix.

## Approved landing

Independent review tr-3ir approved without refinements. The user confirmed all green and authorized this separate commit and push.

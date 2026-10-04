---
# tr-3ir
title: Review first-run database directory initialization
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T13:30:00Z
updated_at: 2026-10-04T13:32:59Z
parent: tr-twk
---

Review tr-b9u before its separate commit. The user requires a pause after every bugfix for independent feedback.

## Scope

- config/config.go: retain loaded default DSN/path provenance, leave Load read-only, prepare default parents with MkdirAll(0700), and report wrapped path errors.
- commands/config.go: prepare directories through actual WithDB before opening SQLite; all CLI/TUI entry points use it.
- commands/db_directory_test.go: fifteen real command/database cases covering fresh roots, persistence across reopen, modes and explicit overrides.
- README.md and one new AGENTS.md lesson: directory creation, permission preservation, error context and explicit DSN ownership.
- tr-15y: a separate follow-up filed for db.New handles left open on initialization errors; no implementation changes to that issue.

## Review focus

Check that custom nested XDG roots, HOME/relative-XDG fallbacks and valid XDG without HOME initialize successfully. Loading alone must stay read-only. Existing permissions must remain unchanged; new directories use 0700, subject to umask. Initialization uses the captured path if environment variables change after loading.

Config-file, environment, memory URI, manual and changed-after-Load DSNs must retain their behavior and bypass directory preparation. An environment DSN equal to the default string is still explicit and bypasses preparation. Existing explicit file DSNs work; missing explicit parents still fail without creation.

Blocked default directories must yield contextual wrapped os.PathError before SQLite or the callback runs, without changing the blocking file. Tests must exercise the real WithDB, write a tool snapshot and read it after reopening the persistent database.

## Validation

Fifteen regression cases pass. Four fresh-path cases and the directory-error regression failed before implementation. An explicit-environment-equals-default edge also failed during self-review and now passes.

Full race suite with existing live crawler exclusions, vet, build, core production lint (zero issues), whitespace checks and Beans integrity pass. A built CLI search from a fresh temporary XDG data root succeeds and creates a real database under a 0700 application directory. Existing exclusions/config repair remain tr-4ek/tr-75b.

AGENTS.md includes unrelated user Beans migration changes; only the new database-initialization lesson belongs to this fix. Leave migration files and other unrelated untracked files outside scope. README db_path correction remains tr-zfh.

- [x] Review the implementation and meaningful regression coverage.
- [x] Record approval or actionable feedback for tr-b9u.
- [x] Verify refinements before authorizing its separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Provenance, not string comparison: Load records defaultDatabaseDSN/defaultDatabaseDir only when the implicit default is applied and clears both when TROVELER_DSN is set — so an environment DSN equal to the default string is structurally explicit and bypasses preparation (dedicated case with a blocking file proving preparation never ran). Config-file DSNs and post-Load mutations (changed cfg.DSN, hand-constructed Config) all leave provenance empty and bypass.
- Initialization uses the captured path: tests change XDG_DATA_HOME after Load and assert the original directory is created, the database file exists there, a tool snapshot persists across a second real WithDB reopen, and the changed-env location is never created.
- Load stays read-only: every fresh-path case asserts the directory does not exist immediately after Load; MkdirAll runs only in EnsureDatabaseDir from WithDB, which is the sole production db.New call site — all nine CLI/TUI entry points route through it (grep-verified).
- Permissions: MkdirAll(0700) for new trees (owner-only bits, umask-stable for common umasks); the existing-directory case pre-chmods to 0750 and asserts MkdirAll left it untouched.
- Coverage of locations: custom nested XDG root containing a space, HOME fallback, relative XDG fallback, and absolute XDG with empty HOME all initialize and persist.
- Blocked defaults fail fast and contextually: a file where the directory belongs yields a wrapped os.PathError containing the full path and "create default database directory", before SQLite opens and before the callback runs, with the blocking file's contents byte-identical afterward. Missing explicit parents still fail at db.New with no directory created (asserted).
- Memory URIs, explicit files, and environment overrides all retain prior behavior; the environment-over-file precedence case also proves the invalid file-DSN parent is never touched.
- Test quality: 15 cases through the real WithDB pipeline with real SQLite writes and reopens — no mocks; the handle-leak-on-init-error follow-up is correctly isolated in tr-15y with no implementation overlap.

Gates independently reproduced: focused WithDB/XDG suites, go vet, go build, git diff --check, full go test -race ./... with the tr-4ek exclusions — all green.

tr-b9u is unblocked for its separate commit (code + beans together).

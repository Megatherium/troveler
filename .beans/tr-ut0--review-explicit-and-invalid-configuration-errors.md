---
# tr-ut0
title: Review explicit and invalid configuration errors
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T13:37:08Z
updated_at: 2026-10-04T13:41:43Z
parent: tr-twk
---

Review tr-z3s before its separate commit. The user requires a stop after each bugfix for independent feedback.

## Scope

- config/config.go: distinguish an explicit request before resolving the default path, read once, allow only os.ErrNotExist for an implicit default, and include the selected path and wrapped cause in read/parse errors.
- config/config_errors_test.go: fifteen actual-loader cases using real files, directories and permissions.
- README.md and one AGENTS.md lesson: optional missing defaults versus explicit and invalid config failures.

## Review focus

Only an absent implicit default may use defaults. Missing explicit paths must fail even when their string equals the default location or TROVELER_DSN is set. Other default filesystem failures must surface; ENOTDIR and permission errors must not be treated as absence.

Directories, nondirectory parents, unreadable files and inaccessible parents must fail at both locations. TOML parse errors must identify the selected path. All failures return nil configuration and preserve errors.Is/errors.As for filesystem/TOML causes. Valid default/explicit files and existing XDG/DSN precedence must remain working. Single-read loading removes the stat/read gap.

The fifteen cases distinguish actual configuration values and exact default DSNs. Eight cases failed before implementation. Permission cases ran under the normal non-root user, with an explicit root-only skip because root bypasses those restrictions.

## Validation

Config tests, existing command-level WithDB regressions, full race suite with live crawler exclusions, vet, build, core production lint (zero issues), whitespace and Beans checks pass. CLI smoke: missing explicit --config plus an in-memory DSN exits nonzero with its path before database initialization; missing implicit default still permits a successful search with the override. Existing crawler/linter exceptions remain tr-4ek/tr-75b.

AGENTS.md also contains the user's uncommitted Beans migration changes. Only the new Config.Load lesson belongs to this fix. Leave migration files and unrelated untracked files outside scope. README db_path example correction stays tr-zfh.

- [x] Review the complete implementation and regression coverage.
- [x] Record approval or actionable feedback for tr-z3s.
- [x] Verify refinements before authorizing the separate commit.

Follow-up tr-xti records the duplicate fatal stderr diagnostic observed in the CLI smoke (Cobra and main.go both print the returned error). Only the follow-up bean was created; error-reporting implementation remains outside this fix.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Only an absent implicit default is optional: the optional flag is captured from the empty argument BEFORE default resolution, so an explicit path that literally equals the resolved default string remains explicit and must exist — every explicit variant in the test passes exactly the default-computed string. TROVELER_DSN does not rescue a missing explicit config (dedicated case).
- Single read removes the stat/read gap: os.ReadFile replaces the Stat-then-DecodeFile sequence; absence is detected via errors.Is(err, os.ErrNotExist) on the actual read, so a file vanishing between a hypothetical stat and read can no longer be silently skipped.
- Non-absence failures surface everywhere: directories (EISDIR), nondirectory parents (ENOTDIR), unreadable files and inaccessible parents (EACCES) all fail for both default and explicit paths — none can masquerade as absence.
- Error quality: every failure includes the selected path in the message, returns a nil Config (no partials), and preserves causes — errors.As(*os.PathError) with exact Path equality, errors.Is(ErrNotExist) and errors.Is(ErrPermission), and errors.As(toml.ParseError) for malformed files.
- Regressions guarded: valid default and explicit files load real values (dsn :memory:, width 79); missing implicit default still yields the exact XDG-derived default DSN and width 50; the prior XDG and WithDB suites remain green alongside.
- Test hygiene: permission fixtures run under the normal user with an explicit root skip (root bypasses chmod 000); fifteen cases assert actual configuration values and exact DSNs, not just error presence.

The duplicate fatal-diagnostic observation observed during CLI smoke is correctly quarantined in tr-xti with no implementation overlap; README db_path stays in tr-zfh.

Gates independently reproduced: focused config/command suites, go vet, go build, git diff --check, full go test -race ./... with the tr-4ek exclusions — all green.

tr-z3s is unblocked for its separate commit (code + beans together).

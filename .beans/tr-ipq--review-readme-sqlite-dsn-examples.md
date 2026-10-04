---
# tr-ipq
title: Review README SQLite DSN examples
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T14:18:21Z
updated_at: 2026-10-04T14:23:51Z
parent: tr-twk
---

Review tr-zfh before its separate documentation commit. The user requires a stop after each bugfix for independent feedback.

## Scope

- README.md: replace ignored db_path/tilde example with supported dsn, explain relative/absolute and literal path semantics, default selection, legacy behavior and highest-priority TROVELER_DSN, and provide an executable shell override example.
- AGENTS.md: one new database-configuration-examples lesson matching the README and existing loader behavior.

## Review focus

The full TOML sample must load unchanged and select troveler.db in the command's working directory. The comments must make this choice explicit and explain omitting dsn to retain the XDG default. Configured SQLite paths are literal, whereas the invoking shell can expand variables in TROVELER_DSN. Explicit parent directories remain caller-managed, as already documented.

Confirm db_path stays ignored without a new compatibility alias. This change must only correct documentation; no production behavior changes or unrelated examples belong to this issue. Existing TUI appearance fields in the full sample must still validate.

## Validation

A manual CLI checker at /tmp/troveler-check-readme-dsn.py extracts the exact README TOML sample, runs the built CLI with isolated HOME/XDG roots, verifies the intended database file instead of the default data root, and reads its tools schema/count through SQLite. The intended-database assertion failed before the documentation change and passes now.

The checker also verifies TROVELER_DSN precedence over the sample, executes the exact README shell command (a temporary PATH alias points to the built binary), verifies shell PWD expansion and checks a newly created alternate database, verifies literal tilde and HOME paths in TOML, and confirms legacy db_path still selects the XDG default. Fixtures are automatically cleaned up. All checks pass.

Existing config tests, whitespace and Beans integrity checks pass. No production code changed; no permanent test was added for this small documentation correction.

AGENTS.md contains unrelated user Beans migration edits. Only the new database-configuration-examples lesson belongs to this fix; leave migration files and unrelated untracked files outside scope.

- [x] Review the documentation and sample validation.
- [x] Record approval or actionable feedback for tr-zfh.
- [x] Verify refinements before authorizing the separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Documentation-only change verified end-to-end against the real loader and CLI, independently of the implementer's checker:

- Extracted the exact README TOML sample (lines 303-322) and ran the built CLI with isolated HOME/XDG roots and --config: the command succeeds (proving the full sample parses and validates, including theme "default" with gradient_colors and both width fields) and creates troveler.db in the command's working directory; the XDG data root is not created. Matches the new sample comment exactly.
- Legacy db_path: swapped the sample's dsn line for the old db_path form and reran — the key is silently ignored (no compatibility alias added) and the XDG default database is selected with parent directories prepared by the existing tr-b9u behavior, exactly as the prose states.
- Precedence and expansion: ran the sample with TROVELER_DSN set to the README's shell-expanded form (file:$PWD/alternate.db...) — alternate.db is created and used, confirming TROVELER_DSN outranks the configured dsn and that expansion happens in the invoking shell, while TOML paths stay literal.
- Scope discipline: git diff touches only README.md and the AGENTS.md migration noise plus exactly one new database-configuration-examples lesson matching the README; no production code, no unrelated examples, and the neighboring tr-b9u directory paragraph (explicit parents caller-managed) is retained.
- Gates: config tests, go vet, build, git diff --check all green. Skipping a permanent test for a pure documentation correction is reasonable; this review itself re-executes the sample.

tr-zfh is unblocked for its separate documentation commit (README + the one AGENTS.md lesson + beans together).

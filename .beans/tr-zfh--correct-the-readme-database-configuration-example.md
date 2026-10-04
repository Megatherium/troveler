---
# tr-zfh
title: Correct the README database configuration example
status: completed
type: bug
priority: low
created_at: 2026-10-03T20:23:09Z
updated_at: 2026-10-04T14:23:51Z
parent: tr-twk
---

README.md documents db_path with a tilde path, while config.Config accepts dsn and does not expand tilde. The sample configuration therefore silently uses the default database. Correct the example and explain DSN and TROVELER_DSN behavior; document any legacy compatibility decision.

## Acceptance criteria

Load the documented sample config in a meaningful check and verify it selects the intended database. Keep README.md and AGENTS.md behavioral guidance consistent.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

README now uses the supported dsn key with a working file:troveler.db?cache=shared&mode=rwc example, explicitly relative to the command working directory. It explains that omitted/empty dsn selects the XDG default, configured TOML paths do not expand tilde/environment variables, absolute file DSNs select fixed locations, and TROVELER_DSN has highest precedence. The copied shell override example demonstrates expansion by the shell. The ignored legacy db_path behavior is documented without adding a compatibility alias or changing application code. AGENTS.md records matching guidance.

## Validation

A reproducible manual CLI check extracts the exact full TOML sample from README, passes it through the real config loader and database initialization, checks the intended troveler.db exists, and reads its tools schema/count via SQLite. This assertion failed with the old sample and passes with the correction. It also verifies a configured-DSN environment override, executes the exact README shell example after deleting the alternate file to avoid a vacuous existence check, verifies tilde and HOME strings remain literal in TOML file paths, and confirms legacy db_path still uses the XDG default. All checks run with isolated temporary HOME/XDG roots and clean up their fixtures. The checker is /tmp/troveler-check-readme-dsn.py.

Existing config tests, git diff --check and beans check --json pass. This is a documentation-only fix; no production code or permanent tests were added.

## Review status

Implementation and self-review are ready. Stopped for independent review; this fix is uncommitted and must not be pushed or followed by another fix until approved.

Independent review is tracked in tr-ipq. Resume with review feedback before committing or starting the next fix.

## Approval and landing

Independent review tr-ipq approved without refinements. The user confirmed All green, commit and continue. Landing this documentation fix in its own commit with both Beans and the matching AGENTS.md lesson.

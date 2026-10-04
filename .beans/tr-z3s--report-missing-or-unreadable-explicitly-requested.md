---
# tr-z3s
title: Report missing or unreadable explicitly requested configuration
status: completed
type: bug
priority: normal
created_at: 2026-10-03T20:23:09Z
updated_at: 2026-10-04T13:41:43Z
parent: tr-twk
---

config.Load silently ignores every os.Stat error, including a nonexistent explicitly supplied config path. Distinguish an absent optional default config from an invalid explicit config and surface other filesystem failures.

## Acceptance criteria

Cover an absent default config, nonexistent explicit config, malformed TOML, and an unreadable or otherwise invalid config path.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

Config.Load remembers whether the caller explicitly selected a file, then reads that path once instead of checking stat and decoding afterward. Only a missing implicit default file uses normal defaults. Every explicit read failure and every other default read failure stops loading with the selected path and a wrapped filesystem cause. Malformed TOML also identifies the selected path and preserves its parse error. All failures return nil configuration before any defaults or TROVELER_DSN overrides.

Fifteen actual-loader cases cover missing default/explicit files, an explicit missing path with DSN override, valid files at both locations, malformed TOML, directories, nondirectory parents, unreadable files and inaccessible parents. The explicit paths equal the default lookup path, ensuring the distinction comes from caller intent. Eight cases failed before implementation; all fifteen now pass. Permission cases ran as the normal non-root user; they skip when root would bypass the restrictions. README and AGENTS.md document the error/optional-default behavior.

## Validation

- go test ./config -count=1: PASS, including all fifteen new cases and existing XDG/config coverage.
- go test ./commands -run ^TestWithDB -count=1: PASS.
- go test -race ./... with TestFetchAndParseSlugs excluded: PASS.
- go vet ./... and go build: PASS.
- Core production golangci-lint (govet/staticcheck/unused/ineffassign): 0 issues.
- git diff --check and beans check --json: PASS.
- Built CLI search with a missing explicit --config and TROVELER_DSN=:memory: exits nonzero with the selected path and never initializes data directories.
- The same CLI search with an absent implicit default config succeeds using the DSN override.

Known live-site crawler exclusions remain tr-4ek and the incompatible checked-in linter config remains tr-75b. README db_path correction stays assigned to tr-zfh.

## Review status

Implementation and self-review are ready. Stopped for independent review; this fix is uncommitted and must not be pushed or followed by another fix before approval.

Independent review is tracked in tr-ut0. Resume with review feedback before committing or starting another fix.

The CLI smoke also reproduced duplicate fatal error diagnostics from Cobra and main.go. Filed separate low-priority follow-up tr-xti; no error-reporting implementation changes were combined with this fix.

## Approved landing

Independent review tr-ut0 approved without refinements. The user confirmed all green and authorized this separate commit and push.

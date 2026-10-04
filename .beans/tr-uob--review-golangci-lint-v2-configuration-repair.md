---
# tr-uob
title: Review golangci-lint v2 configuration repair
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T14:31:17Z
updated_at: 2026-10-04T14:39:46Z
parent: tr-twk
---

Review the isolated configuration repair step of tr-75b before its separate commit. The user requires a stop after every bugfix for independent feedback.

## Scope

.golangci.yml migration; README contributor lint commands/status; exactly one new lint-configuration lesson in AGENTS.md. No Go source changes. Include tr-75b, this review Bean and the tr-dvn follow-up inventory in the eventual commit. Leave unrelated Beans migration edits/files untouched.

## Review focus

Confirm the exact enabled/disabled checks, standard defaults, timeout, Go version, formatter settings and intended exclusions remain active under the v2 schema. Check translated option names, valid current-year header and generated-file behavior. Path patterns must match whole components and affect both linters and formatters. Removed malformed/obsolete fields must not silently disable intended checks.

## Validation

golangci-lint config verify passes with v2.14.0. Isolated manual fixture checker /tmp/troveler-lint-review/check-config.py copies the exact config: clean source passes all configured checks, deliberate blank-assigned errors and long lines fail, current-year headers pass and missing headers fail, test exemptions and bin exclusions including formatters remain effective. It also asserts preserved checks/settings/timeouts/output limits. Temporary fixtures are removed automatically.

Go race suite (only known live crawler tests skipped, tr-4ek), vet, build, whitespace and Beans integrity pass. Full configured project lint exits 1: 248 uncapped diagnostic entries are inventoried in tr-dvn; default output may report fewer due to the retained cap/deduplication. This is explicitly not a green full lint gate. Configuration approval must not close tr-75b: its original full-gate acceptance remains unmet and blocked by tr-dvn. Individual source bugs/refactors require their own tracked and reviewed commits.

No blanket suppressions, new-code baseline, disabled checks or forced zero exit status were added. Assess whether the configuration-only commit is ready while leaving the broader issue open.

- [x] Review migration semantics and documentation.
- [x] Reproduce configuration and fixture validation; confirm project lint findings remain visible.
- [x] Record approval or actionable refinement feedback for the isolated commit step.


## Summary of Changes

**Verdict: APPROVED for the isolated configuration commit — tr-75b remains open as instructed.**

Migration semantics verified line by line:
- Structure: linters-settings -> linters.settings, issues.exclude-rules -> linters.exclusions.rules, top-level limits -> issues.*, skip-dirs/skip-files -> whole-component path regexes ((^|/)component(/|$)) in BOTH linters.exclusions.paths and formatters.exclusions.paths. config verify passes with v2.14.0 (reproduced).
- Real translation repairs, not losses: the v1 nolintlint require-explaining typo is fixed to require-explanation; goconst.ignore-tests became the goconst entry in the _test.go exclusion rule (behavior preserved via a different mechanism); govet printf is subsumed by enable-all; goheader moved from the broken [year]/values form to the native {{ YEAR }} current-year variable; presets: [] faithfully translates exclude-use-default: false; generated: lax retains v1 generated-file behavior.
- Dropped v1 options (unused.check-exported/go-modules, gofmt.rewrite-tests, errcheck.exclude-dirs vendor) were removed upstream in v2 or are covered by path exclusions — no intended check silently disabled; the enable list itself is unchanged (all 27 linters verified present, including goheader, lll, ineffassign).

Fixture battery reproduced in an isolated module with the exact config:
- Real ineffassign case (shadowed assignment) flagged; 133/155-char lines flagged by lll; missing header flagged by goheader while 2026 headers pass; errcheck/revive exempted in _test.go; bin/ excluded from BOTH linters and formatters (an unformatted bin/ fixture produced zero findings).
- Combined runs show fewer issues than isolated runs due to issues.uniq-by-line: true collapsing same-line multi-linter reports — this is retained v1 parity, explicitly documented, not a suppression.

No-suppression confirmed: no baseline, no blanket exclusions, no disabled checks beyond the v1 disable list, no forced exit status. Full configured project lint exits 1 with findings visible across linters (lll, nlreturn, nolintlint, revive, staticcheck, ...), matching the tr-dvn inventory expectation. Standard gates green: go vet, build, git diff --check; Go sources untouched (diff is .golangci.yml + docs only).

Readiness: the configuration-only commit is ready (config + README + the one AGENTS.md lesson + tr-75b, tr-uob, tr-dvn beans). tr-75b must NOT be closed by this approval — its full-gate acceptance remains unmet pending tr-dvn.

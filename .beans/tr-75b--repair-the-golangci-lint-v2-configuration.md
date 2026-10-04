---
# tr-75b
title: Repair the golangci-lint v2 configuration
status: in-progress
type: bug
priority: normal
created_at: 2026-10-03T20:23:09Z
updated_at: 2026-10-04T14:39:46Z
parent: tr-twk
blocked_by:
    - tr-dvn
---

.golangci.yml declares version 2 but includes legacy keys such as skip-dirs, skip-files, linters-settings, exclude-rules and disable-all in invalid locations. golangci-lint config verify fails. Convert the configuration to valid v2 structure and resolve relevant lint failures without suppressing real defects.

## Acceptance criteria

golangci-lint config verify succeeds, configured lint runs successfully, and intended linter settings and exclusions are retained.

- [x] Implement the configuration repair in isolation; resolve source findings through tr-dvn.
- [x] Validate configuration with meaningful lint fixtures and run quality gates; full project lint remains red as documented.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit the configuration repair separately after review, including Beans and ticket references.

## Review scope and remaining acceptance work

The configuration repair is an isolated review/commit step. The original acceptance criterion that the full configured lint run exits zero remains required; it has not been met and must not be checked off or represented as passing. Restoring ignored settings reveals 248 uncapped diagnostics across 16 linters. Remaining triage and individual fixes are tracked in tr-dvn, which blocks completion of this Bean. Keep tr-75b in-progress after approving/committing the configuration step until the full gate passes.

## Summary of Changes

Moved linter settings/exclusion rules into linters, formatter settings/exclusions into formatters, output caps/deduplication into issues, and retained run tests/timeout/Go version. Kept the exact enabled/disabled linter lists and standard defaults, retained the intended test exemptions (including goconst), ST1005/ST1006 exceptions and disabled default exclusion presets. Generated-file behavior stays v1 lax. Directory patterns now match complete path components and include Beans metadata alongside legacy tracker directories; exclusions apply to both linters and formatters. Preserve implicit v1 third_party/builtin/examples exclusions.

Translated misspell ignore-words to ignore-rules and nolintlint require-explaining to require-explanation. Corrected goheader to its built-in current-year template and strip the final newline, verified with a real compliant header. Removed unrecognized gofmt rewrite-tests, unused check-exported/go-modules, errcheck exclude-dirs (vendor remains globally excluded), malformed govet printf bool (enable-all enables printf), and malformed goheader values map. README and AGENTS.md document verification, real lint status and separately reviewed follow-ups.

## Validation

golangci-lint 2.14.0 rejected the original schema. The migrated file passes golangci-lint config verify. An isolated temporary module uses an exact copy of the configuration: clean production code passes the entire configured lint run; deliberate blank-assigned errors and >120-column lines are reported; a valid current-year header passes while a missing header fails; test exclusions suppress only the configured checks and excluded bin code produces no linter or formatter findings. The checker also asserts enabled/disabled checks, settings, timeout, limits and formatter exclusions match intended configuration. Manual checker: /tmp/troveler-lint-review/check-config.py; fixtures are automatically removed.

Race-enabled Go suite passes with only the known live crawler cases excluded (tr-4ek); go vet, build, whitespace and Beans integrity checks pass. Full configured lint still fails on the tracked source findings. No Go source changed and no unrelated lint fix is included.

Stop for independent review before committing this configuration step or implementing another bug/refactor.

- [ ] Full configured project lint exits zero after separately reviewed fixes tracked in tr-dvn.

## Independent review handoff

Review is tracked in tr-uob. Final exact command golangci-lint run ./... exits 1 and reports 223 findings with retained output limits/deduplication (248 uncapped entries). Configuration verification and the isolated clean/negative fixture checks pass. This configuration repair remains uncommitted pending independent review; tr-75b must stay open until tr-dvn restores the full green gate.

## Configuration commit approval

Independent review tr-uob approved the isolated configuration change without refinements; the user confirmed All green, commit and continue. Landing the configuration repair with tr-dvn inventory. This Bean stays in-progress: the full configured lint gate remains unmet.

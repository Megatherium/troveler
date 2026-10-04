---
# tr-g19
title: Review integration success and failure validation
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T15:00:57Z
updated_at: 2026-10-04T15:06:53Z
parent: tr-twk
---

Review tr-nzs before its separate bugfix commit. The user requires a stop after each fix.

## Scope

integration/run_tests.sh: separate status/stdout/stderr capture, meaningful JSON/command checks, bat/btop per-tool dry-run plans and outcomes, non-root exec status propagation, caller PATH and overridable results directory. integration/Dockerfile: jq dependency. New integration/run_tests_test.go: real runner regression battery with isolated command fixtures and real jq. README and exactly one new AGENTS.md integration lesson. Include tr-nzs, this review Bean and follow-ups tr-e79/tr-0k3 in the eventual commit; preserve unrelated user migration edits and files.

## Review focus

Every command content check must require exit zero and validate stdout only. JSON must contain exactly one array and the expected record, filters must reject empty/mixed-language results. Install/mise checks need the actual tool command, not a matching banner/error. Batch must verify bat and btop plans, both completed outcomes and summary, reject failed/skipped tools even with overall exit zero, and not execute those plans. Non-root su must propagate the Troveler status without head and return curl.

Verify 26 deliberate failure cases fail exactly their expected check with 8 unaffected passes, while the valid fixture has 9 passes. Fixture execution must never install packages or change users. Confirm the two local gosec exemptions are limited to safe test-owned paths/executable permissions and ShellCheck SC2016 exemption is intentional expansion in testuser shell. Bash/jq missing on a host causes explicit regression skips; both are dependencies of the image.

## Validation

Targeted regressions failed against the original runner: search exit 23, a failed batch tool and a hidden non-root failure each incorrectly yielded 9 passes. The repaired runner passes the whole battery. Full race Go suite under disabled downloads/refused HTTP proxies, vet, build, ShellCheck, Bash syntax, config verification and whitespace checks pass. Full configured lint stays red on the same 248 uncapped tr-dvn findings, with no integration diagnostics.

Updated Alpine image built successfully with Podman from clean tracked context plus changed integration files and the existing ignored database fixture. Real image localhost/troveler-review-nzs passes all 9 checks and exits zero, including APK install, actual CLI output and su to testuser. Logs: /tmp/troveler-nzs-review/build.log and container-run.log. No application install/search behavior was changed. Docker README context and helper DSN/build bugs remain separate tr-e79/tr-0k3 follow-ups.

- [x] Review runner semantics, fixtures and documentation.
- [x] Reproduce regression and real container validation.
- [x] Record approval or actionable refinement feedback.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Status/stdout/stderr separation: run_command captures stdout only, redirects stderr to a file, and records the exit status under set -o pipefail; every content check is gated on the command's success, and stderr is used solely for failure diagnostics. The search_stderr_only fixture (valid JSON on stderr, empty stdout, exit 0) fails — proving stdout-only validation.
- JSON predicates: jq -s requires exactly one top-level array containing the expected record (multi-array and wrong-record fixtures rejected); the filter check requires nonempty and all-Go records (empty and mixed-language fixtures rejected).
- Actual commands, not banners: install/mise checks use fully anchored table-row regexes with the real command text (apk add btop, dlvhdr/diffnav@version, mise use --global cargo:bat@?) — banner-only, error-text-mentioning-command, and wrong-repo fixtures all rejected.
- Batch semantics: both bat and btop apk plans required verbatim (grep -Fxq), both completed outcomes plus the exact summary, negative assertion rejects any Failed/Skipped even with overall exit zero; no --run flag so plans are never executed (fixture declines sudo/blind/mise with --reuse-config true).
- Non-root: su - testuser -c '... exec troveler ...' propagates Troveler's exit status with no truncating pipeline; SC2016 disabled intentionally so PATH/XDG expand in testuser's shell; wrong-record and nonzero-exit fixtures rejected, valid run returns the curl record.
- Battery reproduced: all 26 fault cases fail exactly their expected check while the other 8 stay green (asserted counts), and the valid fixture yields 9 passes with exit zero. Fixtures are test-owned 0700 scripts in a private temp dir with fake sudo/apk/su (exec passthrough, same user) — no package installation, user change, or network; jq is real. gosec exemptions limited to the two test-owned paths; bash/jq absence produces explicit skips; jq added to the Dockerfile deps.
- Real container reproduced independently: re-ran the implementer's locally built localhost/troveler-review-nzs image — all 9 checks pass, exit 0, including real APK install, real CLI output, and su to testuser.
- Gates reproduced: go test -race ./integration (3.9s), full race suite with refused proxies (integration package included), bash -n, shellcheck (clean), git diff --check; integration package lints clean in isolation (the two scoped nolint gosec comments only).

tr-nzs is unblocked for its separate commit (runner + Dockerfile + test + README + one AGENTS.md lesson + tr-nzs, tr-g19, tr-e79, tr-0k3 beans).

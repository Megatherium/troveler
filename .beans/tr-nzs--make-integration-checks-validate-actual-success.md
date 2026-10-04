---
# tr-nzs
title: Make integration checks validate actual success
status: completed
type: bug
priority: normal
created_at: 2026-10-03T20:23:09Z
updated_at: 2026-10-04T15:06:53Z
parent: tr-twk
---

integration/run_tests.sh often matches words in combined stdout/stderr without checking command success, and the non-root test can pass on head's exit status even when troveler fails. Validate exit codes and meaningful expected behavior instead of banners or error text.

## Acceptance criteria

Intentionally failing troveler commands cause the relevant checks to fail. Search/filter tests check result content, batch tests check per-tool outcomes, and the non-root pipeline propagates failure.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

The integration runner captures command status/stdout separately from stderr and requires successful execution before content checks. Search and non-root checks use jq to validate a single JSON array containing the expected bat/curl record; filter search must return a nonempty array whose languages are all Go. Install displays require actual command rows for btop/APK and diffnav/Go; mise requires the complete bat cargo command. The Go version check verifies both exit status and version format.

Batch dry run explicitly declines sudo, skip-blind and mise prompts, checks both bat/btop APK plans, each completed tool and the summary, and rejects failed/skipped outcomes even when the CLI exits zero. Replaced gomi in this Alpine success fixture because the stored gomi instructions have no Alpine method. No --run flag is used. Non-root search execs Troveler through su with no head pipeline and validates returned JSON. The runner honors caller PATH (image already provides mise), supports RESULTS_DIR for isolated runs, creates the results directory, uses/cleans a temporary stderr file and disables colors for assertions.

The image now installs jq. README/AGENTS.md describe meaningful checks, Bash/jq regression prerequisites, results output and the container APK network requirement. No CLI application behavior is changed. Existing Docker documentation context and helper DSN/build oversights are filed separately in tr-e79 and tr-0k3.

## Regression Coverage

New integration/run_tests_test.go runs the real Bash script with temporary PATH command fixtures and real jq, without installing packages, changing users or accessing a database/network. Valid commands must produce 9 passes. Each of 26 faults must produce exactly its expected failed check and 8 other passes: nonzero commands with valid-looking output, stderr-only output (including exit zero), empty/wrong/multiple search JSON values, wrong command rows, bad mise output, banner-only/failed/skipped/wrong-tool/missing-command batch results, wrong-language filters and failed/wrong non-root results. Bash/jq absence explicitly skips these host regression checks; the integration image includes both.

Before the semantic fix, targeted regression cases proved the original runner incorrectly reported 9 passes for a failed search, a failed batch tool and a non-root Troveler failure hidden by head. All new cases pass after repair. Two narrow gosec exemptions in the regression fixture explain owner-only executable files and reading its fixed private temporary result path; one ShellCheck exemption explains intentional expansion in the target user shell. No global linter configuration was changed.

## Validation

Full go test -race ./... -count=1 -timeout 60s passes, including the new integration package, under disabled module downloads and refused external HTTP proxies. Vet, build, shellcheck, bash -n, lint config validation, whitespace and Beans integrity pass. Full configured lint still reports the same 248 uncapped diagnostics from tr-dvn, with all per-linter counts unchanged and no integration findings.

Built the updated Dockerfile with Podman using a clean temporary copy of tracked HEAD plus changed integration files and the existing ignored database fixture. Image localhost/troveler-review-nzs built successfully; podman run --rm localhost/troveler-review-nzs returns zero with all 9 real checks passing, including APK installation, actual CLI JSON/install output, both batch plans and real su testuser search. Build and run logs are /tmp/troveler-nzs-review/build.log and container-run.log. Containers are removed; local image remains available for review.

## Review status

Implementation and self-review are ready. Stopped uncommitted for independent feedback before landing or starting another fix.

Independent review is tracked in tr-g19. Resume with review feedback before committing or starting the next fix.

Independent review tr-g19 approved without refinements; user authorized landing with All green. Commit and continue. Landing as a separate conventional bugfix commit with both review and follow-up Beans.

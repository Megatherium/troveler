---
# tr-vnt
title: Review deterministic crawler slug discovery tests
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T14:45:09Z
updated_at: 2026-10-04T14:47:48Z
parent: tr-twk
---

Review tr-4ek before its separate commit. The user requires a stop after every bugfix.

## Scope

commands/update.go: private two-method searchPageFetcher seam and parameter type only; existing concrete crawler implements it. commands/update_test.go: replace live-site tests with scripted API fixtures using the real parser. README and exactly one new AGENTS.md lesson: default suite needs no live-site crawler access; smoke tests must be opt-in. Include tr-4ek, this review Bean and follow-up tr-6mk in the eventual commit. Leave unrelated user Beans migration files/edits outside scope.

## Review focus

Ensure fixture checks assert actual content and call contracts, not merely nonempty output. Verify realistic page sizes, deterministic ordering, metadata for both Boolean values, original found totals, all limit boundaries, blank slug handling, empty response, context propagation, wrapped fetch and initial parse errors, and no leaked partial results on reported errors. Production callers must compile without adapters or runtime changes.

The later-page missing/parse behavior is a separate existing defect tracked in tr-6mk; this fix must preserve production semantics and not silently bundle it. The fixture fetcher tests orchestration and real parsing, not the HTTP worker implementation or retry transport. No live-site smoke test is retained because the old checks were vacuous.

## Validation

Full race-enabled Go suite passes without any skip filter, with GOPROXY/GOSUMDB off and HTTP/HTTPS/ALL proxy http://127.0.0.1:1, NO_PROXY empty. Vet, build, config validation, focused final race tests and whitespace checks pass. Full configured lint still exits 1 with the same 248 uncapped diagnostics, tracked in tr-dvn; all linter counts unchanged and the replacement fixture test file is clean. Do not misreport full lint as green.

- [x] Review code, fixture coverage and documentation.
- [x] Reproduce meaningful checks and confirm no default live-site access.
- [x] Record approval or actionable refinement feedback.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Production change is minimal and adapter-free: an unexported two-method searchPageFetcher interface plus the fetchAndParseSlugs parameter type; *crawler.Fetcher satisfies it structurally and the function body is otherwise untouched (diff-verified), so later-page missing/parse semantics are preserved, not silently fixed or locked in — correctly left to tr-6mk.
- Fixtures assert actual content and call contracts, never nonempty-output: exact slug sequences in page order, full weekly-flag map equality covering both Boolean values (index%7==0), the original found total (205, not the truncated count), the initial page-1 call, and the exact concurrent page count per limit.
- Boundary coverage: realistic 100-per-page sizing with 205 tools across 3 pages; no limit, negative, within-first-page, exact-100, 101-crossing, within-last, and above-found limits each assert slug count, ordering, and fetched page count; blank slugs are skipped with flags and totals intact; a found=0 response asserts zero pages requested and zero results.
- Error paths: initial fetch, initial parse (a genuinely malformed JSON body parsed by the real parser, asserted via errors.As json.SyntaxError), and page-fetch failures all assert stage-prefixed messages, wrapped causes via errors.Is/As, and NO leaked partial results (nil slugs/flags, zero total).
- Context propagation is asserted by identity: both fixture methods fail the test if the caller's context is not the one received.
- Hermetic reproduction: full -race suite passes with NO skip filter under GOPROXY=off, GOSUMDB=off, HTTP/HTTPS/ALL_PROXY=http://127.0.0.1:1 and empty NO_PROXY — the historical -skip TestFetchAndParseSlugs exclusion is now unnecessary. go vet, build and git diff --check pass; tr-6mk exists for the deferred page-defect work.

The retired live-site tests were vacuous by construction (logged and passed when the network was absent); removing them without an opt-in smoke replacement is the right call per the project's vacuous-test lesson.

tr-4ek is unblocked for its separate commit (commands changes + README + one AGENTS.md lesson + tr-4ek, tr-vnt, tr-6mk beans).

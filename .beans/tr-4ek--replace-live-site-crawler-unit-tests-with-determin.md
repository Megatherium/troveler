---
# tr-4ek
title: Replace live-site crawler unit tests with deterministic fixtures
status: completed
type: bug
priority: normal
created_at: 2026-10-03T20:23:09Z
updated_at: 2026-10-04T14:47:48Z
parent: tr-twk
---

commands/update_test.go calls terminaltrove.com in default unit tests. Both tests fail with HTTP 403, and cannot pass reliably offline. Inject a fetcher/transport or serve local fixtures; keep any live-site smoke test separate and explicitly opt-in.

## Acceptance criteria

Default go test ./... needs no live crawler network access. Fixtures cover slug extraction, pagination, limits and fetch/parse failure reporting.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

Slug discovery now accepts a private two-method searchPageFetcher interface, implemented unchanged by the existing crawler.Fetcher. Replaced the two live-site tests and their vacuous log-only assertions with local scripted page fixtures exercising the real JSON parser. No public constructor, HTTP globals, production crawler transport or discovery algorithm changed. No live-site smoke test is retained; any future smoke test must be separate and opt-in.

Cases assert exact ordered slugs and complete true/false weekly metadata, original found count, caller context and initial/batch fetch calls. The catalog spans 205 hits over three realistic pages; limits cover no/negative limit, within first page, exact 100, crossing 101, within last page 201 and above found. Additional cases cover blank slug rejection without consuming the positive limit, empty results/zero page count, initial fetch failure, initial JSON syntax failure and concurrent-page fetch failure. Failures preserve wrapped causes and return no partial results.

README and AGENTS.md describe network-independent unit tests and the private seam. Malformed or missing later pages are currently silently skipped; that pre-existing production behavior is filed separately in tr-6mk rather than changed in this test-isolation fix.

## Validation

Full go test -race ./... -count=1 -timeout 60s passes with no skip filter, including all former live-site tests. It was run with GOPROXY=off, GOSUMDB=off and HTTP/HTTPS/ALL proxy variables set to http://127.0.0.1:1, NO_PROXY empty; no live-site access or module downloads were needed. go vet ./..., build, configuration validation and whitespace checks pass. Final focused race tests also pass after the empty fixture helper adjustment.

Full configured lint still exits 1 with exactly the existing 248 uncapped diagnostics (tr-dvn); counts match for all 16 linters and commands/update_test.go has no configured findings. No new suppression or disabled check was added.

## Review status

Implementation and self-review are ready. Stop uncommitted for independent feedback before landing this fix or starting another.

Independent review is tracked in tr-vnt. Resume with review feedback before committing or beginning the next fix.

## Approval and landing

Independent review tr-vnt approved without refinements. The user confirmed All green, commit and continue. Landing the isolated fixture fix with the review and tr-6mk follow-up Beans.

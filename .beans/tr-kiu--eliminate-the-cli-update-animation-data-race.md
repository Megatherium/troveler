---
# tr-kiu
title: Eliminate the CLI update animation data race
status: completed
type: bug
priority: normal
created_at: 2026-10-03T20:23:08Z
updated_at: 2026-10-04T12:22:19Z
parent: tr-twk
---

commands/update.go protects writes to slugBuffer but renderChaoticStream reads the slice and entries without taking bufferMu. A focused race-detector reproduction reported races between AddSlug and Render. Render from a consistent snapshot or protect all accesses to animation state.

## Acceptance criteria

Run concurrent slug updates and rendering under go test -race. Verify normal rendering and frame advancement remain correct.

- [x] Implement this fix or refactor in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

renderChaoticStream now clones slug entries and captures the frame step together under bufferMu, then releases the mutex before rendering. Every rendering read uses the local snapshot. Existing AddSlug/ticker mutations remain protected by the same mutex, while processed counts remain atomic. Slow text/color formatting does not hold the workers' mutex, and animation calculations remain unchanged.

Added three meaningful tests: four concurrent slug producers and two renderers verify display dimensions, exact final processed count/progress and the bounded 30-entry buffer; rendering alongside the actual runUpdateUI ticker verifies frames and slug ages advance together and the loop stops on cancellation; deterministic rendering verifies exact noise/slug fragments before and after a frame advance plus intermediate progress. Both concurrent tests failed under the race detector before implementation, with reported races against AddSlug and the ticker's age/step writes. All three now pass.

README documents the concurrent update progress display and --log alternative. AGENTS.md records the animation snapshot and locking invariant.

## Validation

- go test -race ./commands -run '^TestUpdateUI' -count=1 -timeout 30s: PASS.
- go test -race ./... -skip '^TestFetchAndParseSlugs' -count=1 -timeout 60s: PASS.
- go vet ./... and go build: PASS.
- golangci-lint run --no-config --enable-only govet,staticcheck,unused,ineffassign --tests=false --timeout 2m ./...: 0 issues.
- git diff --check and beans check --json: PASS.

Known live-site crawler exclusions remain tracked in tr-4ek; the incompatible checked-in linter configuration remains tracked in tr-75b.

## Review status

Implementation and self-review are ready. Stopped for independent user review. This fix remains uncommitted and must not be pushed or followed by another fix until approved.

Independent review is tracked in tr-d0o. Resume with review feedback before committing or starting another bugfix.

## Approved landing

Independent review tr-d0o approved without refinements. The user confirmed all green and authorized the separate commit and push.

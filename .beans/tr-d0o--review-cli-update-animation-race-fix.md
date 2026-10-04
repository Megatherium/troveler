---
# tr-d0o
title: Review CLI update animation race fix
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T08:12:54Z
updated_at: 2026-10-04T12:22:19Z
parent: tr-twk
---

Review tr-kiu before its separate commit. The user requires a pause after every bugfix for independent feedback.

## Scope

- commands/update.go: clone slug entries and capture the frame step together under bufferMu; render entirely from immutable locals after unlocking. Existing atomic counters, buffer bounds, animation formulas and ticker behavior remain intact.
- commands/update_ui_test.go: concurrent slug producers/renderers assert frame dimensions, exact final progress and buffer bounds; concurrent actual-ticker rendering asserts age/step advancement and cancellation; deterministic noise/slug fragment checks confirm normal animation and intermediate progress.
- README.md and one AGENTS.md lesson: concurrent update display, --log alternative and snapshot locking guidance.

## Review focus

Verify every read of mutable animation entries and frame step during rendering uses the same locked snapshot, with no slice backing-array alias retained. Check AddSlug and ticker writers use the same mutex and rendering holds it only for capture. Confirm animation formulas, progress counters, dimensions and lifecycle behavior remain correct. Tests must exercise the public Render path concurrently with real producers and the actual ticker, and assert meaningful output/state.

## Validation

Both concurrent tests reported data races before the fix: renderChaoticStream against AddSlug's slice/entry writes and runUpdateUI's age/step writes. All three new tests now pass under -race. Full race suite with existing live-site exclusions, go vet, build, core production lint (0 issues), whitespace and Beans integrity checks pass. Existing exclusions/config repair remain tracked in tr-4ek and tr-75b.

AGENTS.md also contains the user's uncommitted Beans migration edits; only the new CLI animation lesson belongs to this fix. Other migration changes and unrelated untracked files remain outside scope.

- [x] Review the complete implementation and regression coverage.
- [x] Record approval or actionable feedback for tr-kiu.
- [x] Verify any required refinements before authorizing the separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- Single locked snapshot: renderChaoticStream captures slices.Clone(slugBuffer) and step together under bufferMu, then renders entirely from immutable locals. slices.Clone allocates a fresh backing array, so no alias to the producer-trimmed slice survives; entry values (strings/ints) are copied, making the ticker's in-place age++ on original elements harmless to an in-flight render.
- Writers share the mutex: AddSlug (append + 30-entry trim) and the ticker (age++/step write) both hold bufferMu; rendering holds it only for the copy-and-capture, never during string building.
- Remaining Render reads are safe: processed via atomic load; totalTools/startTime immutable after construction; terminal width comes from getTerminalWidth(), not shared struct state. Full state-access map verified — no unguarded mutable reads remain.
- Formulas and lifecycle preserved: the deterministic animation test asserts exact noise lines and slug fragments for two frame states; I independently recomputed both expectations from the render formulas (pos=(startOffset-age)%width with negative wrap, charIdx rotation, noise shift, the dist<4 boundary column) and they match byte-for-byte. Progress counters, ETA, buffer bound of 30, and cancellation all asserted.
- Tests exercise the real paths: the first race reproduces with 4 AddSlug/IncProcessed workers against 2 goroutines calling the public Render (exact per-frame dimensions via ansi.StringWidth, final 800/800 (100%), bounded buffer); the second runs the actual runUpdateUI ticker goroutine (logOutput=true suppresses printing) against concurrent public renders, asserting step/age advance together and the loop stops within 1s of cancellation. Both previously reported data races under -race and now pass repeatedly (-count=2).

Gates independently reproduced: focused -race tests x2, go vet, go build, git diff --check, full go test -race ./... with the tr-4ek exclusions — all green.

tr-kiu is unblocked for its separate commit (code + beans together).

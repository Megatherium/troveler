---
# tr-0jh
title: Review rejecting superseded install requests
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T06:42:23Z
updated_at: 2026-10-04T07:07:26Z
parent: tr-twk
---

Review tr-t9s before its separate commit. The user requires a pause after each bugfix for independent feedback.

## Scope

- tui/panels/install.go: selection generation advances on Clear/SetTool; normal/mise messages carry their origin; InstallRequest captures immutable command/version data; direct panel actions use it; MatchesSelection validates a nonzero, current generation with commands.
- tui/update_batch.go: global single-tool hotkeys use the same snapshot factory; batch-install dispatch stays unchanged.
- tui/update_handlers.go: both execution handlers drop requests from earlier generations before changing execution/modal state.
- tui/install_request_test.go: 18 regression cases covering both message sources and install modes, different tools, identical command strings, A-to-B-to-A transitions, same-tool refreshes, missing origin, delayed tea.Cmd evaluation and fresh request acceptance.
- tui/selection_state_test.go and tui/update_test.go: earlier tests now generate real versioned requests to keep their existing behavior checks meaningful.
- README.md and one AGENTS.md lesson: stale request handling and Update-thread snapshot guidance.

## Review focus

Verify all production message sources capture origin before background execution; every Clear/SetTool invalidates earlier requests; both execution handlers validate origin before queuing a process; identical IDs/command strings do not bypass invalidation; fresh normal/mise actions retain exact command transformation and execution-state behavior. Confirm no mutable model/panel fields are read by tea.Cmd request closures and no fixture install command is executed by tests.

## Validation

All 18 new cases failed before the fix and now pass. Focused TUI/panel tests, the full race suite with existing live-site exclusions, go vet, build, core production lint (0 issues), git diff --check and beans check pass. Known outstanding gates are tracked in tr-4ek and tr-75b.

AGENTS.md also contains the user's uncommitted Beans migration edits; only the new install-request generation lesson belongs to this fix. Other migration changes and unrelated untracked files are outside scope.

- [x] Review the complete implementation and regression coverage.
- [x] Record approval or actionable feedback for tr-t9s.
- [x] Verify any required refinements before authorizing the separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus:
- All production message sources capture origin before background execution: InstallRequest is the sole emitter of both InstallExecuteMsg and InstallExecuteMiseMsg (grep-verified); global hotkeys and direct panel actions both route through it. Batch install flows through its own batchInstallProgressMsg pipeline and never emits these message types, so leaving batch dispatch unchanged is sound, not an oversight.
- Every invalidation path advances the generation: setSelectedTool always calls Clear (and SetTool when commands load), and both increment selectionGeneration — nil selection, lookup failure, zero instructions, and successful loads all supersede earlier requests. MatchesSelection requires nonzero AND current generation AND HasCommands, which subsumes the previous tr-c39 guard.
- Identical IDs/commands cannot bypass: the counter is monotonic per panel, so same-tool refreshes and A-to-B-to-A round trips invalidate old requests even when tool and command strings match exactly.
- No mutable state in tea.Cmd closures: InstallRequest snapshots command and generation on the Update thread and the closure returns a pre-built immutable message; nothing dereferences the panel or model at evaluation time. All generation reads/writes happen on the Update thread.
- Fresh actions preserved: normal and mise transformations are byte-identical (asserted against install.TransformToMise output), and execution/modal/output state transitions match the previous behavior in the positive routing tests.
- Test safety and quality: 18 cases (2 sources x 2 actions x 4 transitions + 2 origin-less messages); transition cases deliberately evaluate the queued tea.Cmd only after the selection change, proving capture-at-keypress; stale deliveries assert execution/modal/output state untouched; fresh deliveries assert full acceptance; fixture commands are never executed. Earlier tr-c39 tests were upgraded from generation-zero messages to real captured requests, keeping their regression value meaningful under the new guard instead of passing trivially.

Non-blocking observation: a recreated InstallPanel restarts at generation 0, which fails MatchesSelection — the safe direction (all pre-recreation requests rejected); no production path recreates the panel mid-session.

Gates independently reproduced: focused TUI/panels tests, go vet, go build, git diff --check, full go test -race ./... with the tr-4ek exclusions — all green.

tr-t9s is unblocked for its separate commit (code + beans together).

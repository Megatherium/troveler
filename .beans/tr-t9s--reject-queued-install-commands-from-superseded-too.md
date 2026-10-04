---
# tr-t9s
title: Reject queued install commands from superseded tool selections
status: completed
type: bug
priority: high
created_at: 2026-10-04T05:11:04Z
updated_at: 2026-10-04T07:07:26Z
parent: tr-twk
---

InstallExecuteMsg and InstallExecuteMiseMsg contain a command but no originating tool ID or selection generation. Capture an install message for tool A, select tool B with valid commands, then deliver A's queued message: handleInstallExecute/handleInstallExecuteMise see B's selection and HasCommands=true and still accept A's command. tr-c39 disables these messages after selection is cleared or instruction loading fails, but does not identify stale messages once another valid selection has loaded.

## Acceptance criteria

Attach and validate the originating selection identity/generation without reading mutable model state in background commands. Drop queued normal and mise messages after selecting another populated tool, and preserve fresh actions for the current selection.

- [x] Implement this fix in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

InstallPanel now advances a selection generation whenever commands are cleared or loaded. Both normal and mise request messages include the originating generation. InstallRequest snapshots the command and generation on the Update thread into immutable message data, and is used by global hotkeys and direct install-panel actions. Both execution handlers ignore requests that do not match the current populated selection.

A generation check also rejects old requests when tool IDs or command strings happen to match, including returning to the same tool and reloading the same tool's instructions. No mutable panel/model state is read by the request's background tea.Cmd.

Added 18 regression cases: 16 combinations of global/panel actions, normal/mise requests and different-tool/same-command/round-trip/same-tool-refresh transitions, plus two messages without an originating generation. All 18 failed before the fix and now pass. Each transition case delays message generation until after the selection update, verifies stale requests leave modal/execution/output state untouched, and verifies a fresh request queues the correct current command. Fixture installs are never executed.

Updated earlier selection-clear/error tests to deliver real captured requests with valid origins, preserving their regression value, and positive install-routing tests to use the production request factory. Updated README and AGENTS.md guidance.

## Validation

- go test ./tui ./tui/panels -count=1: PASS.
- go test -race ./... -skip '^TestFetchAndParseSlugs' -count=1 -timeout 60s: PASS.
- go vet ./... and go build: PASS.
- golangci-lint run --no-config --enable-only govet,staticcheck,unused,ineffassign --tests=false --timeout 2m ./...: 0 issues.
- git diff --check and beans check --json: PASS.

The two known live-site crawler tests remain excluded (tr-4ek); the incompatible checked-in linter configuration remains tracked in tr-75b.

## Review status

Implementation and self-review are ready. Stopped for user review; this fix is uncommitted and must not be pushed or followed by another fix until approved.

Independent review is tracked in tr-0jh. Resume with review feedback before committing or starting another bugfix.

## Approved landing

Independent review tr-0jh approved the implementation without refinements. The user confirmed all green and authorized its separate commit and push.

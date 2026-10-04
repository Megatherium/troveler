---
# tr-t9s
title: Reject queued install commands from superseded tool selections
status: todo
type: bug
priority: high
created_at: 2026-10-04T05:11:04Z
updated_at: 2026-10-04T05:11:04Z
parent: tr-twk
---

InstallExecuteMsg and InstallExecuteMiseMsg contain a command but no originating tool ID or selection generation. Capture an install message for tool A, select tool B with valid commands, then deliver A's queued message: handleInstallExecute/handleInstallExecuteMise see B's selection and HasCommands=true and still accept A's command. tr-c39 disables these messages after selection is cleared or instruction loading fails, but does not identify stale messages once another valid selection has loaded.

## Acceptance criteria

Attach and validate the originating selection identity/generation without reading mutable model state in background commands. Drop queued normal and mise messages after selecting another populated tool, and preserve fresh actions for the current selection.

- [ ] Implement this fix in isolation.
- [ ] Add meaningful regression coverage and run relevant quality gates.
- [ ] Update documentation for behavior changes.
- [ ] Stop and obtain user review feedback before proceeding.
- [ ] Commit separately after review, including this bean and the ticket reference.

---
# tr-9qe
title: Keep batch-mark counts consistent with nonempty filtered searches
status: todo
type: bug
priority: normal
created_at: 2026-10-04T05:11:04Z
updated_at: 2026-10-04T05:11:04Z
parent: tr-twk
---

ToolsPanel.SetTools preserves the markedTools map across filtered search results. GetMarkedCount counts the entire map, but GetMarkedTools only returns marked tools in the current visible list. Mark tool A, search to a nonempty list containing only B, and Alt+I/Alt+M open batch configuration because the count is positive although the actual install set is empty. tr-c39 clears marks on completely empty results; this separate case needs a coherent policy for nonempty filters.

## Acceptance criteria

Make displayed counts, shortcut dispatch, and the actual batch install set agree when marked tools disappear from a nonempty filtered list. Document the chosen retention behavior. Cover filtering and clearing the filter, preserving visible marks and preventing empty batch configuration.

- [ ] Implement this fix in isolation.
- [ ] Add meaningful regression coverage and run relevant quality gates.
- [ ] Update documentation for behavior changes.
- [ ] Stop and obtain user review feedback before proceeding.
- [ ] Commit separately after review, including this bean and the ticket reference.

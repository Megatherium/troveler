---
# tr-9qe
title: Keep batch-mark counts consistent with nonempty filtered searches
status: completed
type: bug
priority: normal
created_at: 2026-10-04T05:11:04Z
updated_at: 2026-10-04T07:15:17Z
parent: tr-twk
---

ToolsPanel.SetTools preserves the markedTools map across filtered search results. GetMarkedCount counts the entire map, but GetMarkedTools only returns marked tools in the current visible list. Mark tool A, search to a nonempty list containing only B, and Alt+I/Alt+M open batch configuration because the count is positive although the actual install set is empty. tr-c39 clears marks on completely empty results; this separate case needs a coherent policy for nonempty filters.

## Acceptance criteria

Make displayed counts, shortcut dispatch, and the actual batch install set agree when marked tools disappear from a nonempty filtered list. Document the chosen retention behavior. Cover filtering and clearing the filter, preserving visible marks and preventing empty batch configuration.

- [x] Implement this fix in isolation.
- [x] Add meaningful regression coverage and run relevant quality gates.
- [x] Update documentation for behavior changes.
- [x] Stop and obtain user review feedback before proceeding.
- [x] Commit separately after review, including this bean and the ticket reference.

## Summary of Changes

GetMarkedCount now counts marked tools in the current result list, using the same membership rule as GetMarkedTools. Existing status-bar rendering, batch wizard counts, and normal/mise shortcut dispatch therefore agree with the actual batch targets without changing install dispatch.

Nonempty searches retain hidden marks for restoration when the tools reappear. Hidden marks do not count or trigger batch configuration; with no visible marks, a populated selected tool still offers its single install. Completely empty searches continue clearing all marks. README documents the policy and a mark-A-and-B/filter-to-B example; AGENTS.md records the invariant.

Regression coverage includes a panel lifecycle test for visible/hidden counts, mark/unmark, filter restoration and clearing hidden IDs; eight actual Model.Update shortcut cases across normal/mise actions and hidden-only, single-install, mixed-mark and restored-filter scenarios; and an empty-search hidden-mark regression. Six shortcut cases and the panel lifecycle failed before the fix. All ten cases now pass. Batch cases complete the real wizard and assert queued target IDs and progress state, without invoking a tool process.

## Validation

- go test ./tui ./tui/panels -count=1: PASS.
- go test -race ./... -skip '^TestFetchAndParseSlugs' -count=1 -timeout 60s: PASS.
- go vet ./... and go build: PASS.
- golangci-lint run --no-config --enable-only govet,staticcheck,unused,ineffassign --tests=false --timeout 2m ./...: 0 issues.
- git diff --check and beans check --json: PASS.

Known live-site crawler exclusions remain tracked in tr-4ek; the incompatible checked-in linter configuration remains tracked in tr-75b.

## Review status

Implementation and self-review are ready. Stopped for independent user review. This fix remains uncommitted and must not be pushed or followed by another fix until approved.

Independent review is tracked in tr-oax. Resume with review feedback before committing or starting another bugfix.

## Approved landing

Independent review tr-oax approved without refinements. The user confirmed all green and authorized the separate commit and push.

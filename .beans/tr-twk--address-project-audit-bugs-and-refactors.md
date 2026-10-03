---
# tr-twk
title: Address project audit bugs and refactors
status: in-progress
type: epic
priority: high
created_at: 2026-10-03T20:21:54Z
updated_at: 2026-10-03T20:28:50Z
---

Address findings from the project audit, keeping each bugfix and each refactor in a separate commit. Stop after each fix for user review; do not begin the next fix or commit/push unreviewed work. Track implementation, regression coverage, documentation, review, and the eventual commit in each child bean.

## Audit inventory

- tr-fxj: Preserve tool identity and atomically refresh install instructions (bug)
- tr-w1q: Preserve Boolean semantics in installed-status filters (bug)
- tr-3bk: Apply installed-search limits after filtering all candidates (bug)
- tr-c39: Clear stale tool selection and install commands after an empty search (bug)
- tr-lgx: Implement actual search debouncing (bug)
- tr-nif: Discard stale asynchronous search results (bug)
- tr-kiu: Eliminate the CLI update animation data race (bug)
- tr-bob: Honor XDG configuration and data directories (bug)
- tr-b9u: Create the database parent directory on first run (bug)
- tr-z3s: Report missing or unreadable explicitly requested configuration (bug)
- tr-0gz: Apply the documented TUI appearance settings (bug)
- tr-zfh: Correct the README database configuration example (bug)
- tr-75b: Repair the golangci-lint v2 configuration (bug)
- tr-4ek: Replace live-site crawler unit tests with deterministic fixtures (bug)
- tr-nzs: Make integration checks validate actual success (bug)
- tr-519: Consolidate CLI and TUI database update pipelines (task)
- tr-799: Centralize installation planning across CLI and TUI flows (task)

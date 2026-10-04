---
# tr-3bk
title: Apply installed-search limits after filtering all candidates
status: todo
type: bug
priority: high
created_at: 2026-10-03T20:23:08Z
updated_at: 2026-10-03T20:47:18Z
parent: tr-twk
---

db/sqlite_search.go uses a fixed over-fetch factor and cap. Matching installed tools beyond the fetched prefix disappear, and an installed OR branch can exceed the requested limit. Return the first requested number of matches in sort order without relying on a guessed installed ratio.

## Acceptance criteria

Seed more nonmatching tools than the current over-fetch prefix, with matches later in the ordering. Verify positive limits, descending order, OR filters, and documented unlimited behavior.

- [ ] Implement this fix or refactor in isolation.
- [ ] Add meaningful regression coverage and run relevant quality gates.
- [ ] Update documentation for behavior changes.
- [ ] Stop and obtain user review feedback before proceeding.
- [ ] Commit separately after review, including this bean and the ticket reference.

## Scope clarification after tr-w1q

The Boolean-expression fix also removes the OR bypass from final trimming, so OR searches obey the requested output limit. The guessed candidate prefix and capped unlimited behavior remain unfixed; retain this bean for correct candidate exhaustion/pagination and the existing acceptance criteria.

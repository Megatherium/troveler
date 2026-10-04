---
# tr-6mk
title: Report malformed or missing crawler search pages
status: todo
type: bug
created_at: 2026-10-04T14:42:28Z
updated_at: 2026-10-04T14:42:28Z
parent: tr-twk
---

During tr-4ek fixture work, fetchAndParseSlugs in commands/update.go and the duplicated internal/update/service.go search phase were found to silently continue when a fetched page is missing or ParseSearchResponse fails. This can report a successful partial crawl and conceal source failures. Decide and document partial-crawl policy, surface page-numbered failures through both CLI and TUI, and add meaningful malformed-later-page/missing-page coverage. Keep this behavioral fix separate from tr-4ek test isolation and stop for review before its own commit.

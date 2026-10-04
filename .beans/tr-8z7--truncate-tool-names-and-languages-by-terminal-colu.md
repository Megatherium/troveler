---
# tr-8z7
title: Truncate tool names and languages by terminal columns
status: todo
type: bug
priority: low
created_at: 2026-10-04T14:05:24Z
updated_at: 2026-10-04T14:05:24Z
parent: tr-twk
---

While fixing tr-0gz tagline rendering, inspection found that ToolsPanel.renderRow still truncates names and languages using byte length and byte slicing. For example, ten copies of 界 in a name have 30 bytes and are cut at byte 22 by the 25-column name limit, splitting a UTF-8 character; a long CJK language label is similarly cut at byte 7. This also over-truncates multibyte text that fits in terminal columns. Apply terminal-column/grapheme-safe truncation to these two fields, preserve column alignment, and add actual rendering regressions for wide and combining characters. This is separate from appearance configuration and tagline capping in tr-0gz. Run relevant gates, stop for review, and commit separately.

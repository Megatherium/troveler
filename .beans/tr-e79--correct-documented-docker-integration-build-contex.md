---
# tr-e79
title: Correct documented Docker integration build context
status: todo
type: bug
priority: low
created_at: 2026-10-04T14:58:31Z
updated_at: 2026-10-04T14:58:31Z
parent: tr-twk
---

README integration instructions cd into integration and run docker build with context ., but integration/Dockerfile copies root go.mod/go.sum and project source. Build therefore needs the repository root context, e.g. docker build -f integration/Dockerfile -t troveler-test . from root, plus an available integration/troveler.db fixture. Correct and validate the documented commands in a separate reviewed documentation fix; discovered while validating tr-nzs.

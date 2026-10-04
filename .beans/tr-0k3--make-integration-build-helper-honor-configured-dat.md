---
# tr-0k3
title: Make integration build helper honor configured database location
status: todo
type: bug
priority: low
created_at: 2026-10-04T14:58:31Z
updated_at: 2026-10-04T14:58:31Z
parent: tr-twk
---

integration/build_and_test.sh copies only HOME/.local/share/troveler/troveler.db, ignoring the migrated XDG/configured DSN behavior, and first builds an unused CGO-disabled SQLite executable even though Docker rebuilds with CGO enabled. Respect or explicitly select the source database and validate build prerequisites, without touching user databases. Keep this helper correction separate from tr-nzs runner success validation and review it before its own commit.

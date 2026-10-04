---
# tr-15y
title: Close SQLite handles when initialization fails
status: todo
type: bug
priority: normal
created_at: 2026-10-04T13:28:09Z
updated_at: 2026-10-04T13:28:09Z
parent: tr-twk
---

db.New in db/sqlite.go opens a sql.DB but returns from Ping, foreign-key setup, createTables and migration errors without closing it. Failed initialization leaves the connection-opener goroutine running; failures after a successful Ping can also retain an open SQLite connection. This is separate from tr-b9u default-directory preparation. Close the newly opened handle on every initialization failure while preserving successful ownership and wrapped errors. Add meaningful failure-path coverage, run relevant gates, update behavior documentation if needed, stop for independent review, and commit separately with this bean reference.

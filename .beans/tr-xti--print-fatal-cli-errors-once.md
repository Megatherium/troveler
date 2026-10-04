---
# tr-xti
title: Print fatal CLI errors once
status: todo
type: bug
priority: low
created_at: 2026-10-04T13:37:32Z
updated_at: 2026-10-04T13:37:32Z
parent: tr-twk
---

The tr-z3s built-CLI smoke test reproduced duplicate stderr diagnostics for a missing --config file: Cobra prints Error: config load: ... during Execute, then main.go prints the same returned error again with fmt.Fprintln. RootCmd does not set SilenceErrors. Choose a single owner for fatal CLI error reporting while preserving nonzero exit codes and the intended help/usage behavior. Cover an actual failing command with captured stderr and verify exactly one error diagnostic, then stop for review and commit the fix separately. This reporting issue is independent of config.Load correctness and is not changed by tr-z3s.

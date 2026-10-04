---
# tr-68b
title: Review XDG config and data directory resolution
status: completed
type: task
priority: normal
tags:
    - review
created_at: 2026-10-04T12:28:49Z
updated_at: 2026-10-04T13:23:05Z
parent: tr-twk
---

Review tr-bob before its separate commit. The user requires a pause after every bugfix for independent feedback.

## Scope

- config/config.go: resolve absolute XDG_CONFIG_HOME and XDG_DATA_HOME independently before HOME lookup; preserve existing HOME and HOME-unavailable fallbacks and DSN query options.
- config/xdg_test.go: 13 actual-loader cases with competing real config files and exact DSN expectations, including custom roots with spaces, independent/mixed roots, unset/empty/relative values, missing custom config, absent HOME and explicit/config/environment overrides.
- README.md and one AGENTS.md lesson: XDG locations, absolute-value validity, fallback behavior and precedence.

## Review focus

Verify valid XDG locations win and work without HOME; unset/empty/relative values use HOME fallbacks independently. Confirm default config lookup selects the correct actual file and does not fall back to HOME when a valid XDG config is missing. Config dsn and TROVELER_DSN must retain their existing precedence, and explicit config paths must bypass default lookup. Directory creation remains assigned to tr-b9u; do not combine it with this fix.

The primary XDG specification used to verify path validity/fallback semantics is https://specifications.freedesktop.org/basedir/latest/. The existing README db_path sample correction remains tr-zfh.

## Validation

Nine cases failed before implementation; all 13 new cases now pass. Config tests, full race suite with existing live-site exclusions, go vet, build, core production lint (0 issues), whitespace and Beans integrity checks pass. Existing exclusions/config repair remain tracked in tr-4ek and tr-75b.

AGENTS.md also contains the user's uncommitted Beans migration edits; only the new XDG-resolution lesson belongs to this fix. Other migration changes and unrelated untracked files remain outside scope.

- [x] Review the complete implementation and regression coverage.
- [x] Record approval or actionable feedback for tr-bob.
- [x] Verify any required refinements before authorizing the separate commit.


## Summary of Changes

**Verdict: APPROVED — no refinements required.**

Verified against the review focus and the freedesktop basedir spec:
- Valid XDG locations win and work without HOME: absolute XDG_CONFIG_HOME/XDG_DATA_HOME are joined directly; HOME is only consulted in the fallback branch (tested with HOME empty, where os.UserHomeDir errors but is never reached).
- Unset/empty/relative values fall back to HOME independently: filepath.IsAbs rejects "", unset, and relative values in one check per variable; config and data roots resolve independently (mixed absolute-config/relative-data case proves independence).
- No HOME cascade: Load resolves exactly one default path; with a valid XDG config root lacking a troveler config while a HOME config deliberately exists, the HOME file is NOT loaded — defaults apply (dedicated test).
- Precedence preserved and verified: file dsn -> defaultDSN() -> TROVELER_DSN env (env wins over file dsn, tested); explicit config path bypasses default lookup entirely (explicit file read, XDG file ignored).
- HOME-unavailable fallbacks retained: relative "config.toml" and cwd-relative DSN, unchanged from before.
- Scope discipline: no directory-creation logic added (correctly deferred to tr-b9u); DSN query options (cache=shared&mode=rwc) preserved verbatim; README db_path sample untouched (tr-zfh).
- Test quality: 13 cases through the actual Load() with real competing config files in both roots distinguished by tagline_width (11 vs 73), exact full-DSN string comparisons including a custom root containing a space, relative/unset/empty/mixed values, missing custom config, absent HOME, and all override combinations; proper t.Setenv hygiene with TROVELER_DSN neutralized where irrelevant.

Non-blocking note: a relative XDG value with no HOME available falls back to the cwd-relative paths — an untested but safe edge matching the pre-existing fallback style.

Gates independently reproduced: config package verbose (all prior tests still green), go vet, go build, git diff --check, full go test -race ./... with the tr-4ek exclusions — all green.

tr-bob is unblocked for its separate commit (code + beans together).

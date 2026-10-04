# Agent Instructions
## Issue Tracking

This project uses **bd (beads)** for issue tracking.
Run `bd prime` for workflow context (MANDATORY!), or install hooks (`bd hooks install`) for auto-injection.

If there's any contradiction: `bd prime` is right. AGENTS.md is not 100% up to date.

## Landing the Plane (Session Completion)

**When ending a work session** before sayind "done" or "complete", you MUST complete ALL steps below.
Work is NOT complete until `git push` succeeds.
Push is not allowed until the work is REVIEWED

**MANDATORY WORKFLOW:**
State A:
  1. **File issues for remaining work** - Create issues for anything that needs follow-up
  2. **Run quality gates** (if code changed) - Tests, linters, builds
  3. **Run CODE REVIEW & REFINEMENT PROTOCOL** - See `bd prime` for details
-- DO NOT CROSS THE LINE BY TOURSELF --
State B (after SOMEONE ELSE has reviewed it):
  4. **Update issue status** - Close finished work, update in-progress items
  5. **PUSH TO REMOTE** - This is MANDATORY:
    ```bash
    git pull --rebase
    git add (careful with using -A, the user sometimes leaves untracked crap lying around) && git commit ...
    git push
    git status  # MUST show "up to date with origin"
    ```
  6. **Clean up** - Clear stashes, prune remote branches
  7. **Verify** - All changes committed AND pushed
  8. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- Pushing is not allowed until the work is successfully reviewed
- If there's only beads/dolt data that needs pushing: amend it to the last commit unless specified

## Modern tooling

All kinds of modern replacements for standard shell tools are available: rg, fd, sd, choose, hck
The interface is nicer for humans. You pick whatever feels right for you.

## Commit Messages

- **Beads extra**: Add a line like "Affected ticket(s): bb-foo", can be multiple with e.g. review tickets
- **WARNING**: Forgetting the ticket reference line is a commit message format violation. Double-check before committing.

## Lessons learned

- db.New must transfer its opened sql.DB to initializeSQLite immediately. Initialization owns the handle until ping, foreign-key setup, schema creation and migrations all succeed; close it on any failure and preserve the original cause (plus any close error). Successful initialization transfers ownership to SQLiteDB.Close. Test each failure stage with a local connector wrapping real SQLite, assert zero retained connections and an unusable handle, and verify success keeps schema, foreign keys and migrated columns usable until caller cleanup.
- Integration checks must capture command status, stdout and stderr separately. Require exit zero before validating a single search JSON array, expected slugs/languages or exact install rows. Batch dry runs must verify both tools' command plans and completed outcomes plus the summary, rejecting failures/skips even when the CLI exits zero. Non-root checks must exec Troveler through su and validate JSON without a truncating pipeline. Run the Bash/jq runner regression tests as part of go test ./integration and validate the real Alpine image; package installation remains a network-dependent container check. RESULTS_DIR can override /app/results for isolated fixtures.
- Default crawler unit tests must use injected searchPageFetcher fixtures, never crawler.NewFetcher against the live site. Keep the real JSON parser in fixture checks, assert exact ordered slugs and weekly flags, page counts at limit boundaries, original found totals, context forwarding and wrapped fetch/initial-parse failures. Any live-site smoke test must be separate and explicitly opt-in. Later-page missing/parse handling is a separate behavioral concern tracked in tr-6mk.
- Lint configuration must validate with golangci-lint config verify before running golangci-lint run ./.... In v2, linter settings/exclusions belong under linters, formatter settings/exclusions under formatters, and issue limits under issues. Preserve the enabled checks, explicit test exclusions and disabled default exclusion presets; do not disable diagnostics to get a green gate. Configuration validation and a successful lint analysis are separate checks: reported findings still fail the lint gate and require tracked, separately reviewed fixes.
- Database configuration examples must use dsn, not the ignored legacy db_path key. Configured SQLite paths remain literal (no tilde or environment expansion); relative file DSNs use the process working directory, and omitted/empty dsn uses the XDG default. TROVELER_DSN overrides configured/default DSNs, with shell expansion only when performed by the invoking shell. Verify documentation examples by loading the actual TOML sample and checking the database file selected by the CLI.
- TUI appearance settings must be validated through TUIConfig.Resolve and applied to each ToolsPanel from NewModel. Keep palettes copied per panel instead of mutating global UI styles: gradient/custom cycle rows, default uses a single row color, and custom requires #RRGGBB colors. Cap tagline columns by configured and available width, truncate by terminal columns without splitting Unicode, and retain selection/mark backgrounds. Config.Load and programmatic TUI startup must surface invalid settings before starting the terminal.
- Config.Load must distinguish an implicit default path from an explicit request before resolving it. Only os.ErrNotExist for an implicit default is optional; every explicit read failure and every other default read failure must be reported with the selected path and wrapped cause. Read once rather than stat then read, include the path in TOML parse errors, and never return a partial config or let TROVELER_DSN mask config failures.
- Database initialization through WithDB must prepare the loaded default data directory with MkdirAll(0700), preserve existing permissions and report wrapped filesystem errors before opening SQLite. Config loading stays read-only; explicitly configured or overridden DSNs remain caller-managed. Capture the default directory during loading rather than re-reading XDG/HOME at initialization, and skip preparation if the DSN changes afterward.
- Default config and database paths must honor absolute XDG_CONFIG_HOME and XDG_DATA_HOME independently, falling back to HOME/.config and HOME/.local/share for unset, empty or relative values. Valid XDG roots work without HOME. Explicit config paths and configured DSNs retain precedence, with TROVELER_DSN highest; do not additionally load HOME config when a selected XDG config is missing.
- CLI update animation rendering must copy slug entries and the frame step together under bufferMu, then render from that snapshot after unlocking. AddSlug and the ticker mutate shared animation state under the same mutex; processed counts remain atomic.
- Every TUI search command, including startup, must capture a unique request generation, the search-panel input generation and the service before background work. Accept success/error messages only for the current pending request and unchanged input, before any state mutation. Successful retries clear only search-owned errors; fixture completions must obtain real request origins so guards do not make tests vacuous.
- Search debounce timers must snapshot query and generation, return expiry messages, and validate them on the Update thread even when another panel is focused. Input changes, Enter and Escape invalidate older timers and queued triggers before database dispatch; forward Escape as a tea.KeyMsg, not a keybinding object. In-flight result/error ordering is a separate concern.
- Batch mark counts and batch targets must both use marked tools in the current search results. Nonempty searches retain hidden marks so they reappear when the filter is cleared; empty searches clear all marks. Hidden marks must not open batch configuration or block a selected tool's single install.
- Every single-tool install request must snapshot its command and install-panel selection generation on the Update thread. Advance that generation when commands are cleared or reloaded, and reject requests from earlier generations in both normal and mise execution handlers, including switches back to the same tool.

- TUI selection changes must clear stored install instructions and both dependent panels before loading the next tool. Empty results also clear batch marks; instruction lookup failures retain only the new tool's metadata and report the error. Cleared commands must disable both install hotkeys and queued install messages.
- Installed-filter searches must examine ordered candidate batches until the requested number of matches is found or candidates are exhausted. Use the same case-insensitive sort and ID tie-breaker for ordering and continuation, close candidate rows before querying installs, and retain the PATH cache across batches. Direct database searches use nonpositive limits for all matches; the search service applies its own default of 50.
- Installed-status search filters must preserve the full AND/OR/NOT expression. Evaluate SQL predicates for both possible installed states, then select the matching flag after the runtime PATH check; never replace installed leaves with unconditional true or skip filtering under OR.
- Crawl persistence must use `db.SaveToolSnapshot` so tool metadata and install instructions refresh atomically. Existing slugs retain their stored IDs and tags; failed saves must be reported by both CLI and TUI updates.
- Be aware of Go's pass-by semantics especially with closures.
- Don't assume you know what a function does by its name alone. The devil is in the details.
- In Bubble Tea, never mutate application state (like maps or UI models) inside a `tea.Cmd` background goroutine; always return a `tea.Msg` and mutate state safely within the main `Update()` thread.
- Never use defer cancel() on a context passed to Dial if the returned connection will use that context after the function returns - always use a separate dial context with its own timeout.
- teatest strips ANSI color codes, making it fundamentally incapable of testing visual focus states - navigation tests belong in agent_tui_test.go where websocket streaming preserves ANSI codes.
- Delete tests that simulate actions but only verify trivial assertions - vacuous tests like assert.True(t, len(out) > 0) provide false confidence and should be removed entirely.
- WaitFor is for observable state changes, not for timing delays - use time.Sleep for rapid key sequences where intermediate states don't produce detectable output differences.
- When the reviewer says "this is completely untouched," stop and actually look at the exact line they're pointing to
- Adding visible text indicators to UI (like ▶ for focus) is valuable for users even when your testing framework can't leverage them - don't conflate UI improvements with testability.
- A 3.0/10 review score means you fundamentally misunderstood the requirements - don't try to justify partial fixes, just implement exactly what the reviewer specified.
- For harness/process validation, binary matching is alias-based (1 harness can map to multiple executable names, e.g. `kilo` and `kilocode`).
- If the UI isn't updating properly: are the caches being dirtied properly?

## Modern tooling

All kinds of modern replacements for standard shell tools are available: rg, fd, sd, choose, hck
The interface is nicer for humans. You pick whatever feels right for you.

## Commit Messages

- **Conventional Commits**: All commit messages **must** adhere to the Conventional Commits specification.
  - **Format**: `<type>[optional scope]: <description>`
  - **Example**: `feat(harvester): implement reverse-scroll logic for Gemini`
  - **Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`
- **Beads extra**: Add a line like "Affected ticket(s): bb-foo", can be multiple with e.g. review tickets
- **WARNING**: Forgetting the ticket reference line is a commit message format violation. Double-check before committing.

## Documentation

- **New Features**: When implementing new features, **must** update documentation:
  - User-facing features: Update README.md with usage examples
  - Template context changes: Document new fields and legacy compatibility behavior
  - Behavioral changes: Update AGENTS.md to inform agents
  - Always keep both files in sync

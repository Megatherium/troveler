## Issue Tracking

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

Run `bd prime` for workflow context (MANDATORY!), or install hooks (`bd hooks install`) for auto-injection.

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

If there's any contradiction: `bd prime` is right. AGENTS.md is not 100% up to date.

### Quick Start

**Check for ready work:**

```bash
bd ready --json
```

**Create new issues:**

```bash
bd create "Issue title" --description="Detailed context" -t bug|feature|task -p 0-4 --json
bd create "Issue title" --description="What this issue is about" -p 1 --deps discovered-from:bd-123 --json
```

**Claim and update:**

```bash
bd update <id> --claim --json
bd update <id> --priority 1 --json
```

**Complete work:**

```bash
bd close <id> --reason "Completed" --json
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)
- `review` - custom type, follow-up to another ticket requesting code review
- `refinement` - custom type, follow-up to a review ticket if defects need to be remediated

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task atomically**: `bd update <id> --claim --json`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" --description="Details about what was found" -p 1 --deps discovered-from:<parent-id> --json`
5. GOTO Session Close Protocol (if you don't know what this is you have been a bad AI and not read `bd prime`)
6. **Complete**: `bd close <id> --reason "Done" --json`
7. Wait for further instructions

### Auto-Sync

bd automatically syncs via Dolt:

- Each write auto-commits to Dolt history
- Use `bd dolt push` / `bd dolt pull` for remote sync
- No manual export/import needed

### Important Rules

- Use bd for ALL task tracking
- Always use `--json` flag for programmatic use
- Link discovered work with `discovered-from` dependencies
- Check `bd ready` before asking "what should I work on?"
- Do NOT create markdown TODO lists
- Do NOT use external issue trackers
- Do NOT duplicate tracking systems

## Documentation

- **New Features**: When implementing new features, update documentation:
  - User-facing features: update `README.md` with usage examples
  - Behavioral changes: update `AGENTS.md` to inform agents
  - Keep both files in sync

## Landing the Plane (Session Completion)

**When ending a work session**, complete all steps below. Work is **not complete** until `git push` succeeds.

1. **File issues for remaining work** - Create follow-up issues for anything outstanding
2. **Run quality gates** (if code changed):
   - `go test ./...`
   - `golangci-lint run`
3. **Run CODE REVIEW & REFINEMENT PROTOCOL** (below)
4. **Update issue status** - Close finished work, update in-progress items
5. **Push to remote**:
   ```bash
   git pull --rebase
   git push
   git status
   ```
6. **Clean up** - Clear stashes, prune remote branches
7. **Verify** - Confirm all changes are committed and pushed
8. **Hand off** - Provide context for the next session

### Code Review & Refinement Protocol

**1) Initiating review**
- If task is housekeeping/docs/small refactor with very low functional risk, you may skip formal review and state why.
- For functional work: do not close the task yet.
- Create review ticket: `bd create --title="Review: <Task Name>" --type=review --json`
- Link dependency so original task is blocked by review: `bd dep add <original-id> <review-id>`
- An agent should not review their own implementation.

**2) Performing review**
- Assess quality, maintainability, code smells, and patterns.
- Assign a mental score (0.0-10.0).
- If unrelated defects are found, create separate issues immediately.
- If score < 8.5, create refinement ticket with concrete fixes:
  - `bd create --title="Refinement: <Task Name>" --type=refinement --json`

**3) Executing refinement**
- Implement fixes from the refinement ticket.
- Re-review only if changes are substantial or wide-ranging.
- Close refinement and original issue only when quality bar is met.

### Best Practices

- Check `bd ready` at session start
- Move issue state as you work (`in_progress` -> `closed`)
- Create new issues immediately when discovering work
- Use clear titles and correct priority/type
- Run `bd sync` before ending session

## Lessons learned

- Be aware of Go's pass-by semantics especially with closures.
- Don't assume you know what a function does by its name alone. The devil is in the details.
- In Bubble Tea, never mutate application state (like maps or UI models) inside a `tea.Cmd` background goroutine; always return a `tea.Msg` and mutate state safely within the main `Update()` thread.
- Never use defer cancel() on a context passed to Dial if the returned connection will use that context after the function returns - always use a separate dial context with its own timeout.
- teatest strips ANSI color codes, making it fundamentally incapable of testing visual focus states - should this become necessary use `agent-tui` and its live websocket feature
- Delete tests that simulate actions but only verify trivial assertions - vacuous tests like assert.True(t, len(out) > 0) provide false confidence and should be removed entirely.
- WaitFor is for observable state changes, not for timing delays - use time.Sleep for rapid key sequences where intermediate states don't produce detectable output differences.
- When the reviewer says "this is completely untouched," stop and actually look at the exact line they're pointing to
- Adding visible text indicators to UI (like ▶ for focus) is valuable for users even when your testing framework can't leverage them - don't conflate UI improvements with testability.
- A 3.0/10 review score means you fundamentally misunderstood the requirements - don't try to justify partial fixes, just implement exactly what the reviewer specified.

## Modern tooling

All kinds of modern replacements for standard shell tools are available: rg, fd, sd, choose, hck
The interface is nicer for humans. You pick whatever feels right for you.

## File Editing Strategy

- **Use the Right Tool for the Job**: For any non-trivial file modifications, you **must** use the advanced editing tools provided by the MCP server.
  - **Simple Edits**: Use `sed` or `write_file` only for simple, unambiguous, single-line changes or whole-file creation.
  - **Complex Edits**: For multi-line changes, refactoring, or context-aware modifications, use `edit_file` (or equivalent diff-based tool) to minimize regression risks.

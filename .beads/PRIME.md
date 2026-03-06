# Beads Workflow Context

> **Context Recovery**: Run `bd prime` after compaction, context clear, or starting a new session.

# SESSION CLOSE PROTOCOL

Before saying work is complete, run this checklist:

```bash
[ ] 0. RUN CODE REVIEW & REFINEMENT PROTOCOL
[ ] 1. git status
[ ] 2. git add <files>
[ ] 3. git commit -m "..."
[ ] 4. git push
```

Work is not done until changes are pushed.

## Code Review & Refinement Protocol

### 1) Initiating Review
When task implementation is functionally complete:
- Stop and do not close the original ticket yet.
- Create review ticket: `bd create --title="Review: <Task Name>" --type=task --json`
- Link dependency so original work is blocked by review: `bd dep add <original-id> <review-id>`
- The implementer should not review their own work.

### 2) Performing Review (Reviewer)
- Review for quality, maintainability, smells, and patterns.
- Assign a mental score (0.0-10.0).
- Create separate issues for unrelated defects discovered.
- Decision logic:
  - Score < 8.5: create refinement ticket with all required fixes.
    - `bd create --title="Refinement: <Task Name>" --type=task --json`
  - Score >= 9.0: reviewer judgment, pass or do minor cleanup.

### 3) Executing Refinement (Implementer)
- Implement required fixes from the refinement ticket.
- Re-review only if changes are substantial or broad.
- Close refinement and original task once quality criteria are satisfied.

## Core Rules

- Use beads for all task tracking (`bd create`, `bd ready`, `bd close`).
- Do not use markdown TODOs or parallel tracking systems.
- Create/claim issue before writing code.
- Check `bd ready` at session start.
- Capture discovered work immediately with `bd create`.
- Use `--json` for programmatic usage.

## Workflow Pattern

1. Start: `bd ready`
2. Claim: `bd update <id> --status=in_progress --json`
3. Work: implement task
4. Capture discoveries: create linked issues
5. Run review/refinement protocol
6. Complete: `bd close <id> --reason "Done" --json`

## Essential Commands

### Finding Work
- `bd ready --json`
- `bd list --status open --json`
- `bd list --status in_progress --json`
- `bd show <id> --json`

### Creating & Updating
- `bd create "Summary" --description "Context and outcome" --type task --priority 2 --json`
- `bd update <id> --status in_progress --json`
- `bd update <id> --priority 1 --json`
- `bd close <id> --reason "Completed" --json`

### Dependencies & Blocking
- `bd dep add <issue> <depends-on>`
- `bd blocked --json`
- `bd show <id> --json`

### Sync & Health
- `bd search <query> --json`
- `bd status --json`
- `bd doctor --json`

## Common Workflows

### Starting work
```bash
bd ready --json
bd show <id> --json
bd update <id> --status in_progress --json
```

### Completing work
```bash
bd close <id> --reason "Completed" --json
git add <files>
git commit -m "..."
git push
```

### Creating dependent work
```bash
bd create "Implement feature X" --description "..." --type feature --priority 2 --json
bd create "Write tests for X" --description "..." --type task --priority 2 --json
bd dep add <tests-id> <feature-id>
```

# Go Details

## MANDATORY: Use td for Task Management

You must run td usage --new-session at conversation start (or after /clear) to see current work.
Use td usage -q for subsequent reads.

We can use modern Go (1.25+). Use any instead of interface{}, range instead of for where appropriate, Waitgroup where useful

# File Editing Strategy

- **Use the Right Tool for the Job**: For any non-trivial file modifications, you **must** use the advanced editing tools provided by the MCP server.
  - **Simple Edits**: Use `sed` or `write_file` only for simple, unambiguous, single-line changes or whole-file creation.
  - **Complex Edits**: For multi-line changes, refactoring, or context-aware modifications, use `edit_file` (or equivalent diff-based tool) to minimize regression risks.

# Git Workflow & Commit Messages

- **Conventional Commits**: All commit messages **must** adhere to the Conventional Commits specification.
  - **Format**: `<type>[optional scope]: <description>`
  - **Example**: `feat(harvester): implement reverse-scroll logic for Gemini`
  - **Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`.

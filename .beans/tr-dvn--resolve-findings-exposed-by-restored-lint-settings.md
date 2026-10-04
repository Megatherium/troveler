---
# tr-dvn
title: Resolve findings exposed by restored lint settings
status: todo
type: task
created_at: 2026-10-04T14:29:39Z
updated_at: 2026-10-04T14:29:39Z
parent: tr-twk
---

Restoring the intended golangci-lint v2 settings in tr-75b exposes existing diagnostics that were hidden by invalid/ignored configuration. Keep them visible and resolve each distinct bug or refactor in a separate Bean and reviewed commit, as requested by the user. Do not broadly disable linters, add a baseline filter, or return exit status zero while findings remain.

## Reproduction

Run golangci-lint config verify, then golangci-lint run ./.... For a complete inventory without output caps or line deduplication:

golangci-lint run --max-issues-per-linter=0 --max-same-issues=0 --uniq-by-line=false --output.json.path=/tmp/troveler-lint-issues.json ./...

The uncapped report contains 248 diagnostics (some linters overlap on the same source line). Full report captured during configuration repair at /tmp/troveler-lint-review/final-issues.json; reproduce it from source rather than relying on this temporary artifact.

## Diagnostic inventory

- errcheck: 39; files: batch_model.go, config.go, fetcher.go, install.go, install_cli.go, main.go, newest.go, search.go, sqlite_refresh.go, sqlite_search.go, sqlite_tags_query.go, sqlite_tools.go, tag.go, update.go, update_install.go, update_model.go.
- errname: 1; files: update_handlers.go.
- errorlint: 3; files: search_limit_integration_test.go, search_results_test.go, update_handlers.go.
- goconst: 19; files: filter_sql.go, match.go, normalize.go, search.go, service.go, sqlite_search.go, update_handlers.go, update_model.go, virtual.go.
- gocyclo: 13; files: batch_model.go, detect.go, fetcher.go, filter.go, install.go, match.go, parser.go, service.go, sqlite_search.go, tools.go, update.go.
- gofmt: 3; files: checker_test.go, modal_manager.go, update_test.go.
- goheader: 62; files: batch_install.go, batch_model.go, checker.go, colors.go, commands.go, config.go, constants.go, consts.go, detect.go, fetcher.go, filter.go, filter_sql.go, formatter.go, info.go, install.go, install_cli.go, install_display.go, keybindings.go, main.go, match.go, modal_manager.go, model.go, models.go, newest.go, normalize.go, parser.go, search.go, selector.go, service.go, slugwave.go, sqlite.go, sqlite_migrations.go, sqlite_refresh.go, sqlite_schema.go, sqlite_search.go, sqlite_tags_mutate.go, sqlite_tags_query.go, sqlite_tools.go, styles.go, table.go, tag.go, tools.go, transform.go, tui.go, update.go, update_batch.go, update_handlers.go, update_install.go, update_keys.go, update_model.go, view.go, view_layout.go, view_panels.go, virtual.go.
- goimports: 3; files: checker_test.go, modal_manager.go, update_test.go.
- gosec: 6; files: config.go, config_errors_test.go, db_directory_test.go, search_installed_boolean_test.go, search_limit_integration_test.go, update_test.go.
- govet: 8; files: search_results_test.go.
- lll: 14; files: checker.go, info.go, install.go, modal_manager.go, sqlite_search.go, sqlite_tools.go, update_install.go, view.go.
- nlreturn: 51; files: appearance_test.go, config.go, config_errors_test.go, db_directory_test.go, fetcher.go, filter_sql.go, filtered_marks_test.go, install.go, install_request_test.go, modal_manager.go, search_debounce_test.go, search_limit_integration_test.go, search_results_test.go, selection_state_test.go, service.go, sqlite_refresh.go, sqlite_search.go, tools.go, transform.go, tui.go, update.go, update_batch.go, update_handlers.go, update_keys.go, update_model.go, update_test.go.
- nolintlint: 3; files: batch_model.go, parser_test.go.
- revive: 14; files: install.go, modal_manager.go.
- sqlclosecheck: 1; files: sqlite_tools.go.
- staticcheck: 8; files: appearance_test.go, db_directory_test.go, filtered_marks_test.go, modal_manager.go, search_test.go, xdg_test.go.

## Triage guidance

Errcheck flags blank-assigned errors because check-blank is restored. Inspect each call before calling it a bug: SQL rollback/Close and registered flag getters can be deliberate, whereas completion writers, interactive input, subprocess startup and update failures can affect users. Use narrowly explained treatment of deliberate ignored errors; never apply blanket suppression.

Gosec findings include intentional caller-selected configuration paths and test fixtures; assess context and document any narrowly justified exemption. Do not alter test permission scenarios merely to silence a warning.

The SQL close diagnostic is in a batched query loop: defer must not retain rows until the outer function exits with MaxOpenConns(1). If refactoring, isolate a per-batch helper and preserve release-before-next-query behavior.

Search error naming/type checks are internal invariants introduced by tr-nif; inspect whether wrapping can actually occur before changing semantics. Test error assertions should use identity/wrapping checks instead of whole-struct equality. Staticcheck QF findings, goconst, gocyclo, header, formatting, API comments and return spacing are refactors to review separately.

## Acceptance criteria

- [ ] Triage diagnostics into individual bug/refactor Beans before implementing each change.
- [ ] Review and commit each distinct fix/refactor separately.
- [ ] Full configured golangci-lint run ./... exits zero without disabling intended checks.
- [ ] Record final validation and unblock completion of tr-75b.

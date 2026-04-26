# AGENTS.md

## API Change Workflow

When you change API routes or controllers, update the OpenAPI specification file `docs/openapi.yaml` so it reflects all available endpoints.

Include example request bodies for create and update operations when relevant.

Always run `go vet ./...` and `go test ./...` after making changes and report the results.

## App Summary Workflow

Before making code changes, read `APP_SUMMARY.md` to understand the current application context, architecture, modules, database, API conventions, and important project rules.

When your changes affect the application structure, business logic, database schema, API behavior, authentication, file storage, configuration, or run instructions, update `APP_SUMMARY.md` to keep it accurate.

Do not add speculative or planned features to `APP_SUMMARY.md`. Only document what exists in the current codebase.

If something is not implemented or cannot be verified from the source code, write `Not implemented`, `Not found`, or `Unknown` instead of guessing.
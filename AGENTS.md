# AGENTS.md

## API Change Workflow

When you change API routes or controllers, update the OpenAPI specification file `docs/openapi.yaml` so it reflects all available endpoints.

Include example request bodies for create and update operations when relevant.

Always run `go vet ./...` and `go test ./...` after making changes and report the results.

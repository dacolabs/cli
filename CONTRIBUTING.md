# Contributing

Thanks for helping improve the Daco CLI.

## Development

1. Install a recent Go toolchain (1.24+).
2. Clone this repository.
3. Run `make setup` to enable `.githooks` (Conventional Commits `commit-msg` check).
4. Run `go test ./...` and `go build -o daco .` (or `make test` / `make build`).

Commit messages should follow [Conventional Commits](https://www.conventionalcommits.org/) (for example `feat: …`, `fix: …`, `chore: …`).

## Scope

- This repo owns the **CLI UX** (commands, flags, output, local config).
- Catalog HTTP is a **generated Go client** (`internal/catalogapi`) from `openapi/catalog-api.yaml`; do not hand-edit `*.gen.go`.
- Commands are **not** generated from OpenAPI — keep adding them by hand.
- Generated language SDKs live in [`dacolabs/sdk`](https://github.com/dacolabs/sdk).
- The Catalog OpenAPI contract is owned by [`dacolabs/daco`](https://github.com/dacolabs/daco) (`api/catalog-api.yaml`).

## Regenerating the Catalog client

Requires **Python 3** (for `./scripts/sync-catalog-openapi.sh`) and Go 1.24+.

1. Sync the snapshot (strips monorepo-only `x-go-type` imports). From a standalone clone set the product contract path explicitly:

   ```bash
   CATALOG_OPENAPI=/path/to/daco/api/catalog-api.yaml ./scripts/sync-catalog-openapi.sh
   ```

   When this repo is checked out as a `dacolabs/daco` submodule, `../api/catalog-api.yaml` is used by default.

2. Generate: `./scripts/generate-catalog-client.sh` (uses pinned `oapi-codegen@v2.8.0` via `go run`, or set `OAPI_CODEGEN` to that binary)
3. Verify: `./scripts/check-catalog-client-drift.sh` and `go test ./...`

## Pull requests

- Keep changes focused.
- Include tests for new behavior.
- Do not commit secrets or customer tokens.

# Contributing

Thanks for helping improve the Daco CLI.

## Development

1. Install a recent Go toolchain.
2. Clone this repository.
3. Run `go test ./...` and `go build -o daco .`.

## Scope

- This repo owns the **CLI UX** (commands, flags, output, local config).
- Catalog HTTP is a **generated Go client** (`internal/catalogapi`) from `openapi/catalog-api.yaml`; do not hand-edit `*.gen.go`.
- Commands are **not** generated from OpenAPI — keep adding them by hand.
- Generated language SDKs live in [`dacolabs/sdk`](https://github.com/dacolabs/sdk).
- The Catalog OpenAPI contract is owned by [`dacolabs/daco`](https://github.com/dacolabs/daco) (`api/catalog-api.yaml`).

## Regenerating the Catalog client

1. Sync the snapshot (strips monorepo-only `x-go-type` imports): `./scripts/sync-catalog-openapi.sh`
2. Generate: `./scripts/generate-catalog-client.sh`
3. Verify: `./scripts/check-catalog-client-drift.sh` and `go test ./...`

## Pull requests

- Keep changes focused.
- Include tests for new behavior.
- Do not commit secrets or customer tokens.

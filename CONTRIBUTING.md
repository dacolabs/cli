# Contributing

Thanks for helping improve the Daco CLI.

## Development

1. Install a recent Go toolchain.
2. Clone this repository.
3. Run `go test ./...` and `go build -o daco .`.

## Scope

- This repo owns the **CLI UX** (commands, flags, output, local config).
- Generated language SDKs live in [`dacolabs/sdk`](https://github.com/dacolabs/sdk).
- The Catalog OpenAPI contract is owned by [`dacolabs/daco`](https://github.com/dacolabs/daco).

## Pull requests

- Keep changes focused.
- Include tests for new behavior.
- Do not commit secrets or customer tokens.

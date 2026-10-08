# Daco CLI

Official command-line interface for the [Daco](https://dacolabs.com) Catalog API.

This repository is the **CLI product** (`daco`). Language client libraries are generated in [`dacolabs/sdk`](https://github.com/dacolabs/sdk) from Catalog OpenAPI snapshots.

## Status

AuthKit device login works. Catalog commands are still thin (`datasets` lists the first page).

## Build

```bash
go build -o daco .
./daco --help
```

## Authenticate

```bash
export DACO_CLIENT_ID=client_...   # WorkOS AuthKit client ID (same environment as the web app)
export DACO_BASE_URL=https://...   # Catalog API origin

./daco login    # open the printed URL, confirm the code
./daco whoami
./daco datasets
./daco logout
```

`daco login` uses WorkOS AuthKit **device authorization** (public client; no API key in the binary). Tokens are stored at `~/.config/daco/credentials.json` (`0600`), or `DACO_CREDENTIALS_FILE` when set. Access tokens refresh automatically when possible.

## Environment

| Variable | Purpose |
|---|---|
| `DACO_CLIENT_ID` | WorkOS AuthKit client ID |
| `DACO_BASE_URL` | Catalog API origin |
| `DACO_AUTH_API` | AuthKit API host (default `https://api.workos.com`) |
| `DACO_CREDENTIALS_FILE` | Override credentials path |

## Related

- Product / OpenAPI source of truth: [`dacolabs/daco`](https://github.com/dacolabs/daco) (`api/catalog-api.yaml`)
- Generated language clients: [`dacolabs/sdk`](https://github.com/dacolabs/sdk)

## License

Apache License 2.0

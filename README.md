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

The AuthKit **client ID is public** and baked into the CLI for staging (same class of value as a browser OAuth client id). End users do not need to supply it.

```bash
./daco login      # defaults to staging AuthKit + https://main.app.daco.services
./daco whoami
./daco datasets
./daco logout
```

Local Catalog against your machine:

```bash
export DACO_BASE_URL=http://127.0.0.1:8080
./daco login
./daco datasets
```

`daco login` uses WorkOS AuthKit **device authorization** (public client; no API key or client secret in the binary). Tokens are stored at `~/.config/daco/credentials.json` (`0600`), or `DACO_CREDENTIALS_FILE` when set.

## Environment

| Variable | Purpose |
|---|---|
| `DACO_ENV` | `staging` (default) or `production` |
| `DACO_BASE_URL` | Override Catalog API origin (default follows `DACO_ENV`) |
| `DACO_CLIENT_ID` | Override AuthKit client ID (defaults are public and baked in) |
| `DACO_AUTH_API` | AuthKit API host (default `https://api.workos.com`) |
| `DACO_CREDENTIALS_FILE` | Override credentials path |

Production client ID is not checked in yet; use `DACO_ENV=staging` or set `DACO_CLIENT_ID` explicitly until production binding exists.

## Related

- Product / OpenAPI source of truth: [`dacolabs/daco`](https://github.com/dacolabs/daco) (`api/catalog-api.yaml`)
- Generated language clients: [`dacolabs/sdk`](https://github.com/dacolabs/sdk)
- Public staging AuthKit binding: `infra/workos/staging.json` in the product repo

## License

Apache License 2.0

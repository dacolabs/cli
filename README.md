# Daco CLI

Official command-line interface for the [Daco](https://dacolabs.com) Catalog API.

This repository is the **CLI product** (`daco`). Language client libraries are generated in [`dacolabs/sdk`](https://github.com/dacolabs/sdk) from Catalog OpenAPI snapshots.

## Status

AuthKit device login works. Catalog HTTP calls use an **oapi-codegen** client generated from a CLI-safe Catalog OpenAPI snapshot. Commands (`datasets`, etc.) stay hand-written.

## Build

```bash
go build -o daco .
./daco --help
./daco --version
```

## Install

From a [GitHub Release](https://github.com/dacolabs/cli/releases) (preferred once tagged):

```bash
# example: macOS arm64
curl -fsSL -o daco.tar.gz \
  "https://github.com/dacolabs/cli/releases/download/v0.1.0/daco_0.1.0_Darwin_arm64.tar.gz"
tar -xzf daco.tar.gz daco
sudo mv daco /usr/local/bin/
daco --version
```

Or from source at a version tag:

```bash
go install github.com/dacolabs/cli@v0.1.0
```

## Release

Tagged builds publish archives via GoReleaser (`.goreleaser.yaml`, `.github/workflows/release.yml`):

```bash
git checkout main && git pull
git tag v0.1.0
git push origin v0.1.0
```

`daco --version` reports the release version injected at link time (`-X main.version=…`).

## Catalog OpenAPI client

The committed client lives in `internal/catalogapi/` (generated) plus a small auth helper. Refresh from the product contract when Catalog OpenAPI changes:

```bash
# from a daco monorepo checkout with this repo at cli/, or set CATALOG_OPENAPI=
./scripts/sync-catalog-openapi.sh
./scripts/generate-catalog-client.sh
./scripts/check-catalog-client-drift.sh
```

Requires Go 1.24+ and either a PATH `oapi-codegen` at **v2.8.0** or network access for `go run …@v2.8.0`.

## Authenticate

The AuthKit **client ID is public** and baked into the CLI for staging (same class of value as a browser OAuth client id). End users do not need to supply it.

```bash
./daco login      # defaults to staging AuthKit + https://main.app.daco.services
./daco whoami
./daco datasets
./daco apply -f datasets.yaml
./daco logout
```

### Apply Dataset YAML

Declare one or more `kind: Dataset` documents (multi-doc with `---`, or split across files/dirs). Apply creates missing versions, patches title/description/metadata when the contract matches, and errors if the contract changed (bump `version` instead). No prune.

```yaml
kind: Dataset
urn: urn:daco:dataset:orders
version: "1.0.0"
title: Orders
description: Customer orders
metadata: {}
contract:
  schema:
    type: object
  metadata: {}
```

```bash
./daco apply -f datasets.yaml
./daco apply -f ./manifests --dry-run
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

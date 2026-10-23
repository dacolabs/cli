# Daco CLI

Official command-line interface for the [Daco](https://dacolabs.com) Catalog API.

This repository is the **CLI product** (`daco`). Language client libraries are generated in [`dacolabs/sdk`](https://github.com/dacolabs/sdk) from Catalog OpenAPI snapshots.

> OpenDPI / data-product authoring releases through **v0.3.x** remain available as historical GitHub Releases and on the `opendpi-legacy` branch. From **v1.0.0**, `daco` is the Catalog CLI.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install --cask dacolabs/tap/daco
```

### Scoop (Windows)

```powershell
scoop bucket add dacolabs https://github.com/dacolabs/scoop-bucket.git
scoop install daco
```

### Go Install

```bash
go install github.com/dacolabs/cli@v1.0.0
# installs as $(go env GOPATH)/bin/cli (module path ends in /cli)
mv "$(go env GOPATH)/bin/cli" "$(go env GOPATH)/bin/daco"
```

Prefer Homebrew/Scoop or `make build` when you want the `daco` binary name without renaming.

### Manual Download

Download the appropriate binary from the [releases page](https://github.com/dacolabs/cli/releases).

### Build from Source

```bash
git clone https://github.com/dacolabs/cli.git
cd cli
make build
./bin/daco --help
```

## Status

AuthKit device login works. Catalog HTTP calls use an **oapi-codegen** client generated from a CLI-safe Catalog OpenAPI snapshot. Commands (`datasets`, etc.) stay hand-written.

## Authenticate

The AuthKit **client ID is public** and baked into the CLI for staging (same class of value as a browser OAuth client id). End users do not need to supply it.

```bash
./bin/daco login      # defaults to staging AuthKit + https://main.app.daco.services
./bin/daco whoami
./bin/daco datasets
./bin/daco apply -f datasets.yaml
./bin/daco logout
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
./bin/daco apply -f datasets.yaml
./bin/daco apply -f ./manifests --dry-run
```

Local Catalog against your machine:

```bash
export DACO_BASE_URL=http://127.0.0.1:8080
./bin/daco login
./bin/daco datasets
```

`daco login` uses WorkOS AuthKit **device authorization** (public client; no API key or client secret in the binary). Tokens are stored at `~/.config/daco/credentials.json` (`0600`), or `DACO_CREDENTIALS_FILE` when set.

## Environment

| Variable | Purpose |
|---|---|
| `DACO_ENV` | `staging` (default) or `production` (`prod` alias) |
| `DACO_BASE_URL` | Override Catalog API origin (default follows `DACO_ENV`); `http` only for localhost |
| `DACO_CLIENT_ID` | Override AuthKit client ID (defaults are public and baked in) |
| `WORKOS_CLIENT_ID` | Fallback AuthKit client ID when `DACO_CLIENT_ID` and baked-in defaults are empty |
| `DACO_AUTH_API` | AuthKit API host (default `https://api.workos.com`) |
| `DACO_CREDENTIALS_FILE` | Override credentials path |

Production client ID is not checked in yet; use `DACO_ENV=staging` or set `DACO_CLIENT_ID` explicitly until production binding exists.

## Catalog OpenAPI client

The committed client lives in `internal/catalogapi/` (generated) plus a small auth helper. Refresh from the product contract when Catalog OpenAPI changes:

```bash
# from a daco monorepo checkout with this repo as a submodule, or set CATALOG_OPENAPI=
./scripts/sync-catalog-openapi.sh
./scripts/generate-catalog-client.sh
./scripts/check-catalog-client-drift.sh
```

Requires Go 1.24+, **Python 3** (sync script), and network access for `go run …@v2.8.0` (or set `OAPI_CODEGEN` to an oapi-codegen **v2.8.0** binary).

## Release

Catalog releases use the **Release** workflow (`workflow_dispatch`) with GoReleaser, which publishes GitHub Release archives and opens PRs on [`dacolabs/homebrew-tap`](https://github.com/dacolabs/homebrew-tap) and [`dacolabs/scoop-bucket`](https://github.com/dacolabs/scoop-bucket). Versions must be **v1.0.0+** (0.x remains OpenDPI legacy).

```text
Actions → Release → Run workflow → version: 1.0.0
```

After the first Catalog Cask lands in the tap, complete the one-time Homebrew migration in [`dacolabs/homebrew-tap`](https://github.com/dacolabs/homebrew-tap):

1. Add `tap_migrations.json` with `{"daco":"daco"}` so `brew install daco` redirects formula users to the Cask.
2. Remove legacy `Formula/daco.rb`.
3. Document install as `brew install --cask dacolabs/tap/daco`.

`daco --version` reports the release version injected at link time (`-X main.version=…`).

## Related

- Product / OpenAPI source of truth: [`dacolabs/daco`](https://github.com/dacolabs/daco) (`api/catalog-api.yaml`)
- Generated language clients: [`dacolabs/sdk`](https://github.com/dacolabs/sdk)
- Public staging AuthKit binding: `infra/workos/staging.json` in the product repo

## License

Apache License 2.0

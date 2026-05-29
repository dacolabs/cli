# Daco

Swiss army knife for data product projects.

Daco ships as two binaries from a single repository:

- **`daco`** — the command-line tool for authoring, validating, and translating data product definitions
- **`daco-api`** — an HTTP server exposing the same domain over a REST API

Both are released together under one version, so you can always pair a CLI with a compatible server.

## Features

- **OpenDPI-native** — read and write [OpenDPI](https://github.com/opendpi)-compliant data product specifications
- **Schema translation** — convert port schemas to and from many target formats:
  Avro, Protobuf, JSON Schema, Markdown, Python (Pydantic / dataclasses), PySpark,
  Spark SQL, Spark Scala, Scala, Databricks SQL, Databricks PySpark, Databricks Scala,
  Go types, and Databricks DQX YAML
- **Interactive prompts and TUI** for guided authoring
- **HTTP API** with OpenTelemetry-ready logging for programmatic use and integration into platforms
- **Single-binary distribution** for every supported platform

## Installation

### Homebrew (macOS/Linux)

```bash
brew install dacolabs/tap/daco
```

Installs both `daco` and `daco-api`.

### Scoop (Windows)

```powershell
scoop bucket add dacolabs https://github.com/dacolabs/scoop-bucket.git
scoop install daco
```

### Go Install

```bash
go install github.com/dacolabs/daco/cmd/cli@latest   # daco
go install github.com/dacolabs/daco/cmd/api@latest   # daco-api
```

### Manual Download

Each archive on the [releases page](https://github.com/dacolabs/cli/releases)
contains **both** `daco` and `daco-api` for the target platform.

### Build from Source

```bash
git clone https://github.com/dacolabs/cli.git
cd cli
make build
./bin/daco --help
./bin/daco-api    # starts the API server on :8080
```

## Quickstart

### CLI

```bash
# Initialize a new data product
daco init

# Add a port and translate its schema
daco ports add
daco ports translate --target pydantic

# Inspect what you have
daco describe
```

Run `daco --help` for the full command reference.

### API

The API server reads configuration from environment variables and starts on
`:8080` by default:

```bash
daco-api
```

| Variable | Default | Description |
|---|---|---|
| `SERVER_PORT` | `8080` | HTTP listen port |
| `SERVER_READ_TIMEOUT` | `15s` | Request read timeout |
| `SERVER_WRITE_TIMEOUT` | `15s` | Response write timeout |
| `SERVER_IDLE_TIMEOUT` | `60s` | Keep-alive idle timeout |
| `SERVER_SHUTDOWN_TIMEOUT` | `15s` | Graceful shutdown timeout |
| `LOG_FORMAT` | `text` | `text` (human) or `otel` (structured for OpenTelemetry) |

Health check:

```bash
curl http://localhost:8080/health
```

## Repository Layout

```
cmd/cli/           # daco CLI entrypoint
cmd/api/           # daco-api HTTP server entrypoint
internal/cli/      # CLI commands, config, prompts, TUI
internal/api/      # HTTP handlers and middleware
internal/translate/  # Schema translators (one package per target)
internal/version/  # Build-time version info (shared by both binaries)
docs/              # User-facing documentation
landing/           # Marketing site (Next.js)
```

For a deeper tour and a full development guide, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Versioning

`daco` and `daco-api` share a single semantic version. A release tag (e.g. `v0.3.0`)
publishes both binaries at that version simultaneously, so you never need to chase
compatibility between them.

## Documentation

- [Contributing guide](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)

## License

Apache License 2.0 — see [LICENSE](LICENSE).

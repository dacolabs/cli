# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`daco` is a single-binary Go CLI for authoring and managing **data product specifications** in the OpenDPI format. It manages **products** (OpenDPI documents on disk), **connections** (infrastructure endpoints), **schemas** (JSON Schemas), and **ports** (typed data interfaces inside a product), and **translates** schemas into language/runtime targets (PySpark, Avro, Protobuf, Pydantic, Go types, Scala, Spark/Databricks SQL, DQX YAML, Markdown, etc.).

A daco project on disk is a `daco.yaml` (the project registry) that points at per-product `opendpi.yaml`/`opendpi.json` specs and shared `schemas/*.yaml` / `connections/*.yaml` files. Running `daco` with no arguments opens an interactive TUI.

## Common commands

```bash
make build                    # build ./bin/daco with version ldflags
make run-cli ARGS="--help"    # go run with ldflags; pass args via ARGS
make test                     # go test -race -cover ./...
make test-coverage            # writes coverage.out + coverage.html
make lint                     # golangci-lint run ./...
make format                   # gofmt -s -w . && goimports -w .
make release-snapshot         # local goreleaser dry run
make install                  # go install ./cmd/daco to $GOPATH/bin
```

Run a single test:

```bash
go test -race -run TestPortsTranslate ./internal/cli/engine
go test -race -run TestTranslate       ./internal/translate/pyspark
```

CI (`.github/workflows/ci.yml`) runs lint + race tests + cross-platform build on every push/PR. Conventional Commits are enforced for releases (used to generate the changelog via goreleaser).

## Architecture

### Entry point and bootstrap

```
cmd/daco/main.go        — calls app.Run, marshals engine.Errors to JSON on stderr
  └─ cmd/daco/app/app.go — wires TUI runner + registers every top-level command on cli.New()
       └─ internal/cli   — cobra root + ctx helpers
```

`cmd/daco/app/app.Run` is the **only** place where the command tree is assembled. To add a new top-level command, write it in `internal/cli/commands/` and add one `root.AddCommand(commands.X())` line in `app.go`. The TUI is registered via `cli.TUIRunner = tui.Run` — this function-pointer hook exists to break the otherwise-cyclic dependency (the `tui` package imports `cli` for the user/project context helpers, so `cli` cannot import `tui` directly).

### Three layers: cli ↔ engine ↔ settings/opendpi

The internal package layout reflects a strict three-layer separation:

```
internal/cli/commands/   — cobra wrappers. Build an engine.*Input, call engine.X, print Output. No business logic.
internal/cli/tui/        — tview-based interactive driver. Calls the same engine.X functions.
internal/cli/engine/     — all business logic. Each operation is an Input/Output struct + a function.
internal/cli/settings/   — user & project YAML files (the only package that reads/writes ~/.daco and daco.yaml).
internal/opendpi/        — OpenDPI document/schema/connection types + parse/serialize with $ref preservation.
internal/translate/      — schema translation pipeline (see "Translators" below).
internal/version/        — build-time ldflag-injected version info.
```

**Rule of thumb:** if you find yourself adding logic in `commands/` or `tui/`, it belongs in `engine/`. The two drivers must stay thin — both must be able to reach feature parity by calling the same engine functions.

### The engine pattern

Every operation in `internal/cli/engine/` follows the same shape:

```go
type FooInput  struct { Prj *settings.Project; ... }
type FooOutput struct { ... }

func (in FooInput) Valid(ctx context.Context) map[string]string { ... }  // returns problems keyed by field
func Foo(ctx context.Context, in FooInput) (*FooOutput, error) {
    if err := validate(ctx, in); err != nil { return nil, err }
    ...
}
```

`validate` collects the `Valid()` map into a sorted `engine.Errors` (a `[]string` type that implements `error`). `cmd/daco/main.go` does `errors.As(err, &engine.Errors)` and prints a JSON array on stderr — so all validation failures surface as machine-readable output. **Don't replace `engine.Errors` with `errors.Join` or a custom type** — the JSON contract depends on it.

### Settings: user vs. project

`internal/cli/settings/settings.go` defines two files:

- **User** (`~/.daco/settings.yaml`) — global registry mapping project names to absolute directories. Created automatically on first run by `cli.New()`'s `PersistentPreRunE`. Stashed on `cmd.Context()` via `cli.WithUser` / `cli.User`.
- **Project** (`daco.yaml` in the project root) — per-project registry of `products`, `connections`, `schemas` (each a map of name → relative path), plus the project name.

`settings.LoadProject(usr)` is the key resolution function:

1. Look for `daco.yaml` in CWD → use it.
2. Otherwise walk the user registry and pick the **longest** registered project directory that is a prefix of CWD. This is how subcommands work when run from a subdirectory.
3. Otherwise return `ErrProjectNotFound`.

The TUI also has `settings.LoadProjectAt(dir)` for opening a project picked from the registry without `chdir`. Project loading happens lazily inside each command via `loadProject(cmd)` in `commands/products.go` — it is **not** done in a persistent pre-run hook, because `init` runs before any project exists.

### OpenDPI parsing and round-trip

`internal/opendpi/opendpi.go` defines `Document` (Info / Connections / Ports / Components), `Connection` (which may be a `$ref` or inline `type/host/...`), `Port` (with optional `Schema`), and `Schema` (which wraps `jsonschema.Schema`).

Two things to be careful about:

1. **`$ref` preservation.** A `Connection` or `Port.Schema` that came from disk as a `$ref` must round-trip back to disk **as the same `$ref`**, not inlined. The `Connection.Ref` field and `Schema.Ref` are load-bearing; anything that mutates them risks silently inlining external files on save.
2. **Property order.** `Schema.UnmarshalYAML` populates `PropertyOrder` so that emit order matches source order — this matters for deterministic, diff-friendly translator output. `translate.Prepare` consumes it; don't drop it.

`engine.ProjectFormat` and `engine.ProjectLint` walk the project's Schemas, Connections, and Products in sorted name order; format is a pure load/save round-trip, lint also checks that `$ref`s resolve and that ports reference connections that exist on the product.

### Schema translation pipeline

This is the architectural backbone of the translate layer. All translators share one pipeline in `internal/translate/`:

```
JSON Schema  ──►  translate.Prepare  ──►  SchemaData  ──►  text/template  ──►  []byte
                  (uses TypeResolver)     (Defs, Root,
                                           Fields, Constraints)
```

`Prepare` ([internal/translate/prepare.go](internal/translate/prepare.go)) is **target-agnostic**:

- Walks `$defs` in topological order (a def that references another def is emitted after its dependency).
- Preserves original key order via `Schema.PropertyOrder` / `opendpi.TraverseDefs`.
- Extracts inline nested objects as named `TypeDef`s on the fly (named after the field in PascalCase) and appends them to `Defs`. By the time the template runs, every object has a name.
- Maps each property into a `Field` with `Name`, `Type`, `Nullable` (from `schema.Required`), `Description`, and `Constraints` (all JSON Schema validation keywords).

Per-target behavior is concentrated in a `TypeResolver` ([internal/translate/resolver.go](internal/translate/resolver.go)) implemented by each translator package. The resolver controls:

- Primitive type mapping (`PrimitiveType(schemaType, format)` — format wins, so `"string"` + `"date-time"` can become `TimestampType()`).
- Container wrapping (`ArrayType`, `MapType`).
- Naming (`FormatDefName`, `FormatRootName`, `RefType`).
- `EnrichField(*Field)` — the post-processing escape hatch. Mutate `Name` for casing, wrap `Type` for nullability (`Optional[T]`, `*T`), set `Tag` for struct tags / Python defaults. Called once per field, after type resolution, before template execution.

A new translator is therefore: a `Translator` struct with `Translate` + `FileExtension`, a `resolver` implementing `TypeResolver`, and an embedded `*.go.tmpl` file. **You must also register it in `internal/translate/registry/registry.go`'s `Default()` map** — the CLI's `--format` flag accepts whatever keys are in that map. Look at `internal/translate/pyspark/` and `internal/translate/gotypes/` as canonical examples; both consume the same `SchemaData` shape.

### TUI

`internal/cli/tui/` uses `rivo/tview` (mouse-enabled). `tui.Run(ctx)` loads the user, then either drops into the project shell (`pageShell`) if a project resolves from CWD or the projects-picker view (`pageProjects`). All actions ultimately call into `internal/cli/engine` — the TUI is a second driver, not a fork of the CLI logic.

## Conventions

- **Conventional Commits** required for `feat`, `fix`, `docs`, `chore`, `test`, `refactor`, `ci`, `build`, `perf`, `style` — goreleaser uses them to generate the changelog.
- **License header** on every Go file:
  ```go
  // SPDX-License-Identifier: Apache-2.0
  // Copyright 2026 Daco Labs
  ```
- **Package comments** are required by `revive` (see [.golangci.yml](.golangci.yml)). Every package's first file should have `// Package x ...`.
- **Internal-only:** production code lives under `internal/` (private to the module). There is no public API surface.
- Min Go version is **1.26** (see [go.mod](go.mod)).

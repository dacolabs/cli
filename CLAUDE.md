# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`daco` is a Go CLI for authoring and managing **data product specifications** in the OpenDPI format (`opendpi.yaml` / `opendpi.json`). It manages **ports** (typed data interfaces backed by JSON Schemas), **connections** (infrastructure endpoints), and **product** metadata, and **translates** port schemas to language/runtime targets (PySpark, Avro, Protobuf, Pydantic, Go types, Scala, Spark/Databricks SQL, DQX YAML, Markdown, etc.).

A daco project on disk has two files: `daco.yaml` (a thin config pointing at the spec dir) and `opendpi.yaml`/`opendpi.json` under that dir, plus referenced JSON Schemas (`schemas/*.yaml`).

## Common commands

```bash
make build                    # build ./bin/daco with version ldflags
make run ARGS="ports list"    # go run with ldflags, pass args via ARGS
make test                     # go test -v -race ./...
make lint                     # golangci-lint run ./...
make format                   # gofmt -s -w . && goimports -w .
make setup                    # configure git hooks (commit-msg validates Conventional Commits)
make release-snapshot         # local goreleaser dry run
```

Run a single test:

```bash
go test -v -race ./internal/translate/pyspark -run TestTranslate
go test -v -race ./internal/opendpi -run TestParse/yaml_round_trip
```

CI (`.github/workflows/ci.yml`) runs lint + `go test -v -race ./...` + cross-platform build on every push/PR. Conventional Commits are enforced by the `commit-msg` hook.

## Architecture

### Entry point and bootstrap

`cmd/daco/main.go` is intentionally tiny — it calls `cmd/daco/internal.Run`, which **registers all translators** in a `translate.Register` map and hands them to `commands.NewRootCmd`. Adding a new translator means: implement `translate.Translator` in a new `internal/translate/<name>/` package, then add one line to `registerTranslators` in [cmd/daco/internal/app.go](cmd/daco/internal/app.go). The CLI's translator list is the keys of that map; nothing else needs to know.

### Command tree (cobra)

[internal/commands/register.go](internal/commands/register.go) builds the cobra tree: top-level `init` and `describe`, plus `ports`, `connections`, and `product` groups. The three group commands all set `PersistentPreRunE: session.PreRunLoad`, which is the load-the-project hook — `init` and `describe` deliberately do not, because `init` runs before a project exists.

Every subcommand pulls the loaded project via `session.RequireFromCommand(cmd)` rather than re-reading files. Commands that need user input use `internal/prompts/` (Charm `huh` forms); commands run non-interactively when their key flag is provided (e.g. `daco init --name X` skips the form).

### Session / project loading

[internal/session/context.go](internal/session/context.go) defines the loading contract used by every project-aware command:

1. Find `daco.yaml` in CWD (else `ErrNotInitialized`).
2. Decode + `Validate()` the config (`config.CurrentConfigVersion` must match — currently `1`).
3. Resolve the spec dir from `Config.Path` and look for `opendpi.yaml` then `opendpi.json`.
4. Parse via `opendpi.YAML.Parse(f, os.DirFS(specDir))` — the `fs.FS` is required so external `$ref`s like `schemas/user.yaml` can be resolved.
5. Stash the resulting `*session.Context` (config + parsed spec) on `cmd.Context()`.

The sentinel errors (`ErrNotInitialized`, `ErrInvalidConfig`, `ErrSpecNotFound`, `ErrInvalidSpec`) are part of the contract — wrap, don't replace, when adding new failure modes.

### OpenDPI parsing and round-trip

[internal/opendpi/](internal/opendpi/) parses YAML/JSON specs and **resolves all `$ref`s** (both internal `#/...` and external file refs via the supplied `fs.FS`), unifying every schema into `Spec.Schemas`. `Port.SchemaRef` preserves the **original** external ref path so the writer can round-trip the spec back to disk without inlining schemas. Treat that field as load-bearing — anything that adds/edits ports must set it correctly.

### Schema translation pipeline

This is the architectural backbone of the codebase. All translators share one pipeline in [internal/translate/](internal/translate/):

```
JSON Schema  ──►  translate.Prepare  ──►  SchemaData  ──►  text/template  ──►  []byte
                  (uses TypeResolver)     (Defs, Root,
                                           Fields, Constraints)
```

`Prepare` ([internal/translate/prepare.go](internal/translate/prepare.go)) is **target-agnostic**:

- Walks `$defs` in topological order (so a def that references another def is emitted after its dependency).
- Preserves original JSON/YAML key order via `jschema.ExtractKeyOrder` and `Schema.PropertyOrder` — this matters for deterministic, diff-friendly output.
- Extracts inline nested objects as named `TypeDef`s on the fly (named after the field in PascalCase) and appends them to `Defs`. Don't try to emit nested anonymous types in templates — by the time the template runs, every object has a name.
- Maps each property into a `Field` with `Name`, `Type`, `Nullable` (computed from `schema.Required`), `Description`, and `Constraints`.

Per-target behavior is concentrated in a `TypeResolver` ([internal/translate/resolver.go](internal/translate/resolver.go)) implemented by each translator package. The resolver controls:

- Primitive type mapping (`PrimitiveType(schemaType, format)` — format wins, so `"string"` + `"date-time"` can become `TimestampType()`).
- Container wrapping: `ArrayType`, `MapType`.
- Naming: `FormatDefName`, `FormatRootName`, `RefType`.
- `EnrichField(*Field)` — the post-processing escape hatch. Mutate `Name` for casing, wrap `Type` for nullability (`Optional[T]`, `*T`), set `Tag` for struct tags / Python defaults. Called once per field, after type resolution, before template execution.

A new translator is therefore: a `Translator` struct with `Translate` + `FileExtension`, a `resolver` implementing `TypeResolver`, and an embedded `*.tmpl` file. Look at `internal/translate/pyspark/` and `internal/translate/gotypes/` as canonical examples; both consume the same `SchemaData` shape.

**Before writing or non-trivially editing a translator, read [.claude/rules/translators.md](.claude/rules/translators.md)** — it specifies the package layout, the `EnrichField` mutation order, what `Prepare` guarantees so you don't reimplement it, the `RefType`/`FormatDefName` symmetry rule, and the test matrix every translator should cover.

### Constraints

`translate.Constraints` carries the full set of JSON Schema validation keywords (`Enum`, `Pattern`, `Minimum`, `MultipleOf`, `MinItems`, etc.). Whether a translator emits them is target-specific (e.g. dqx-yaml maps them to data-quality checks; gotypes ignores most). Don't drop fields from `Constraints` — translators decide what to use.

## Conventions

- **Conventional Commits required** (`feat`, `fix`, `docs`, `chore`, `test`, `refactor`, `ci`, `build`, `perf`, `style`). The pre-commit hook installed by `make setup` enforces this.
- **License header** on every Go file:
  ```go
  // SPDX-License-Identifier: Apache-2.0
  // Copyright 2026 Daco Labs
  ```
- **Package comments** are required by `revive` (see [.golangci.yml](.golangci.yml)). Every package's first file should have `// Package x ...`.
- **Internal-only**: production code lives under `internal/` (private to this module) and `cmd/daco/internal/` (private to the CLI binary). There is no public API surface.
- Min Go version is **1.23.0** (see [go.mod](go.mod)).
- Generated CLI reference docs live in [docs/cli/](docs/cli/) — keep them in mind if you change command shape.

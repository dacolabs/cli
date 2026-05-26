---
title: "Daco CLI Changelog: v0.2.0"
date: "2026-02-03"
tag: "Changelog"
excerpt: "The first public release of Daco CLI: schema translation to 12+ formats, connection management, and project scaffolding."
cover: "c1"
glyph: "v0.2"
---

v0.2.0 is the first public release of the Daco CLI — a tool for managing data product projects built on OpenDPI.

## What's in the Box

### Schema Translation

The headline feature: translate any OpenDPI port schema to over 12 formats in one command.

```bash
daco ports translate --format pydantic my-port
```

Supported formats in this release:

| Format | Output |
|---|---|
| `avro` | Avro schema JSON |
| `protobuf` | Protocol Buffers `.proto` |
| `pydantic` | Python Pydantic model |
| `python` | Plain Python dataclass |
| `go` | Go struct |
| `scala` | Scala case class |
| `spark-sql` | Spark SQL `CREATE TABLE` |
| `spark-scala` | Spark Scala schema |
| `pyspark` | PySpark `StructType` |
| `databricks-sql` | Databricks SQL |
| `databricks-scala` | Databricks Scala |
| `databricks-pyspark` | Databricks PySpark |

### Connection Management

Add, list, and remove connections to your data sources:

```bash
daco connections add
daco connections list
```

### Project Scaffolding

Initialize a new data product project with a single command:

```bash
daco init
```

## Internal Improvements

This release also refactored schema preparation internals, unified inline and component schema handling, and standardized CLI input/output operations. These changes lay the groundwork for more translation formats in future releases.

## Installing

```bash
# Homebrew
brew install dacolabs/tap/daco

# Scoop
scoop bucket add dacolabs https://github.com/dacolabs/scoop-bucket.git
scoop install daco

# Go
go install github.com/dacolabs/cli/cmd/daco@latest
```

## Contributors

Giuseppe Grieco and Guðmundur Orri Palsson.

---
title: Quick Start
description: Install the Daco CLI and define your first data product in minutes.
order: 1
section: Getting Started
---

## Installation

**macOS / Linux (Homebrew)**

```bash
brew install dacolabs/tap/daco
```

**Windows (Scoop)**

```bash
scoop install daco
```

**Go (any platform)**

```bash
go install github.com/dacolabs/cli/cmd/daco@latest
```

Verify your installation:

```bash
daco --version
```

## Initialize a project

Create a directory for your data product and run `daco init`:

```bash
mkdir customer-analytics && cd customer-analytics
daco init --name "Customer Analytics" --version "1.0.0"
```

This creates two files:

```
customer-analytics/
├── daco.yaml          # Project config
└── spec/
    └── opendpi.yaml   # Your data product spec
```

`daco.yaml` tells the CLI where your spec lives. You commit both files alongside your pipeline code.

## Add a connection

A connection describes where your data lives — a database, message broker, object store, or any infrastructure endpoint.

```bash
daco connections add \
  --name analytics_db \
  --type postgresql \
  --host analytics.db.example.com
```

Run `daco connections add` with no flags to use the interactive form instead.

## Add a port

A port is a named data interface — a table, topic, file path, or API endpoint that your product exposes.

```bash
daco ports add \
  --name daily_metrics \
  --connection analytics_db \
  --location customer_daily_metrics
```

This creates `spec/schemas/daily_metrics.schema.yaml`. Open that file and define the shape of your data using JSON Schema:

```yaml
type: object
properties:
  customer_id:
    type: string
  date:
    type: string
    format: date
  total_orders:
    type: integer
required:
  - customer_id
  - date
  - total_orders
```

## Translate to a target format

Once your schema is defined, generate framework-specific code:

```bash
daco ports translate --all --format pyspark --output models/
```

Daco writes a ready-to-use file to `models/`. Run the same command with a different `--format` to get the same schema as Pydantic, Go structs, Avro, Protobuf, and more.

## Review your product

Get a full overview of what you've defined:

```bash
daco describe
```

## What's next

- [CLI Reference](/docs/cli-reference) — every command and flag documented
- [OpenDPI Specification](/docs/opendpi) — the full spec format
- [Connections](/docs/connections) — connection types, variables, and examples
- [Ports](/docs/ports) — schemas, `$ref`, and location syntax
- [Code Generation](/docs/translation) — all 14 output formats

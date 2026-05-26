---
title: "Getting Started with the Daco CLI"
date: "2026-02-22"
tag: "Guide"
excerpt: "Step-by-step: install the CLI, initialize a data product, and translate your schema to PySpark, dbt, and more."
cover: "c2"
glyph: "$ daco init\n$ daco validate\n$ daco translate\n  --format pyspark"
---

The Daco CLI is the fastest way to start building data products with OpenDPI. This guide covers installation, project initialization, connection setup, port definition, and schema translation.

## Installation

**macOS / Linux (Homebrew)**
```bash
brew install dacolabs/tap/daco
```

**Windows (Scoop)**
```bash
scoop bucket add dacolabs https://github.com/dacolabs/scoop-bucket.git
scoop install daco
```

**Go**
```bash
go install github.com/dacolabs/cli/cmd/daco@latest
```

Verify the installation:
```bash
daco version
```

## Initialize Your First Data Product

Create a new directory and initialize a project:

```bash
mkdir my-data-product
cd my-data-product
daco init
```

This generates an `opendpi.yaml` file with basic metadata:

```yaml
opendpi: "1.0.0"

info:
  title: My Data Product
  version: "0.1.0"
  description: ""
```

## Add a Connection

Connections tell Daco where your data lives. Run:

```bash
daco connections add
```

The CLI walks you through selecting a connection type (PostgreSQL, Kafka, S3, BigQuery) and configuring the required variables. Your `opendpi.yaml` gains a `connections` section:

```yaml
connections:
  analytics_db:
    type: postgresql
    host: analytics.db.example.com
    variables:
      database: analytics
      schema: public
```

## Define a Port

Ports are the datasets your data product exposes. Add one:

```bash
daco ports add
```

This creates a schema file in a `schemas/` directory and registers the port in your `opendpi.yaml`:

```yaml
ports:
  events:
    description: "Raw user events"
    connection: analytics_db
    location: public.events
    schema:
      type: object
      required: [event_id, user_id, timestamp]
      properties:
        event_id:   { type: string }
        user_id:    { type: string }
        event_type: { type: string }
        timestamp:  { type: string, format: date-time }
```

## Translate Your Schema

This is where it gets useful. Translate your port schema to any supported format:

```bash
# PySpark
daco ports translate --format pyspark events

# dbt model YAML
daco ports translate --format dbt events

# Pydantic model
daco ports translate --format pydantic events

# Go struct
daco ports translate --format go events
```

Supported formats include Avro, Protobuf, Pydantic, Python, Go, Scala, Spark SQL, Spark Scala, PySpark, Databricks SQL, Databricks Scala, Databricks PySpark, DQX YAML, and Markdown — over 12 in total.

## What's Next

- Add more ports to model your full data product interface
- Run `daco validate` to check your schema against the OpenDPI specification
- Push to your repository and connect [Daco Studio](/studio) to get a live catalog

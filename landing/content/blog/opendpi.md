---
title: "OpenDPI: A Standard for Data Product Interfaces"
date: "2026-02-03"
tag: "Standard"
excerpt: "An open standard for describing and documenting data product interfaces — and why your team needs one."
cover: "c2"
glyph: "opendpi:\n  ports:\n    metrics:\n      schema:\n        ..."
---

OpenAPI made APIs machine-readable, and that unlocked an entire ecosystem of tooling — documentation generators, mock servers, type-safe clients, API gateways. Data products deserve the same.

OpenDPI is our attempt to build that standard.

## Why Another Standard?

We looked at what existed. Most data product specifications tried to bundle governance, lineage, quality monitoring, access control, and metadata management into one schema. The result was specifications so complex that adoption stalled before teams shipped anything.

We made a different choice: describe only what a data product exposes and how to connect to it. That's it.

The constraint is intentional. A focused spec is easier to implement, easier to build tooling on top of, and easier to adopt incrementally.

## What OpenDPI Describes

An OpenDPI document has three sections:

**Connections** — where the data lives:
```yaml
connections:
  analytics_db:
    type: postgresql
    host: analytics.db.example.com
    variables:
      database: analytics
      schema: public
```

**Ports** — what the data product exposes, with location and schema:
```yaml
ports:
  daily_metrics:
    description: "Daily aggregated customer metrics"
    connections:
      - connection: "#/connections/analytics_db"
        location: customer_daily_metrics
    schema:
      type: object
      properties:
        customer_id:  { type: string }
        date:         { type: string, format: date }
        total_orders: { type: integer }
        revenue:      { type: number }
```

**Info** — metadata about the product:
```yaml
info:
  title: Customer Analytics
  version: "2.1.0"
  description: Aggregated customer behavior metrics
```

## A Full Example

```yaml
opendpi: "1.0.0"

info:
  title: Customer Analytics
  version: "2.1.0"
  description: Aggregated customer behavior metrics

connections:
  analytics_db:
    type: postgresql
    host: analytics.db.example.com
    variables:
      database: analytics
      schema: public

ports:
  daily_metrics:
    description: Daily aggregated customer metrics
    connections:
      - connection: "#/connections/analytics_db"
        location: customer_daily_metrics
    schema:
      type: object
      properties:
        customer_id:  { type: string }
        date:         { type: string, format: date }
        total_orders: { type: integer }
        revenue:      { type: number }
```

## Tooling First

Standards without tools are just documents. OpenDPI is designed to be machine-readable, which enables:

- **Code generation** — translate a port schema to PySpark, Pydantic, Go, Avro, Protobuf, and more
- **Validation** — check data against the contract at pipeline boundaries
- **Catalog discovery** — index data products automatically without manual entry

The Daco CLI implements all of this today.

```bash
brew install dacolabs/tap/daco
daco init
```

## Open and Community-Driven

OpenDPI is MIT-licensed. The specification lives on GitHub and we welcome contributions — whether that's new connection types, additional schema constraints, or tooling integrations.

A standard only works if it's shaped by the people using it. If you're building data products and have opinions on what should be in the spec, open an issue or a PR.

[View the spec on GitHub](https://github.com/dacolabs/daco) · [Try the CLI](/docs)

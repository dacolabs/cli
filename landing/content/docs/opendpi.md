---
title: OpenDPI Specification
description: The open standard for describing data product interfaces — what it is, how it's structured, and why it exists.
order: 2
section: OpenDPI
---

## What is OpenDPI?

OpenDPI (Open Data Product Interface) is an open, vendor-neutral specification for describing how data products expose their interfaces. It is to data products what OpenAPI is to REST APIs: a machine-readable contract that tools, teams, and AI agents can reason about without needing live access to your warehouse.

An OpenDPI document answers three questions:

1. **What data does this product expose?** — ports define named interfaces (tables, topics, files, API endpoints)
2. **Where does the data live?** — connections declare the infrastructure endpoints
3. **What shape is the data?** — schemas describe the structure using JSON Schema

## Document structure

Every OpenDPI document is a YAML or JSON file with four top-level sections:

```yaml
opendpi: "1.0.0"    # Spec version — always "1.0.0" for now

info:               # Product metadata
  title: string     # Required
  version: string   # Required, semantic versioning
  description: string

connections:        # Infrastructure endpoints
  <name>:
    type: string
    host: string
    description: string
    variables: {}

ports:              # Data interfaces
  <name>:
    description: string
    connections:
      - connection: "#/connections/<name>"
        location: string
    schema: {}      # Inline JSON Schema or $ref

components:         # Optional reusable schemas
  schemas:
    <name>: {}
```

## info

The `info` section identifies the product and its current version. The version follows semantic versioning and is managed with `daco product upgrade`.

```yaml
info:
  title: Customer Analytics
  version: "2.1.0"
  description: Aggregated metrics for the customer lifecycle team.
```

## connections

Each connection represents an infrastructure endpoint where data can be found. The `type` is a free-form string — it exists for human readability and downstream tooling, not for validation.

```yaml
connections:
  analytics_db:
    type: postgresql
    host: analytics.db.example.com
    description: Production analytics warehouse
    variables:
      database: analytics
      schema: public

  events_kafka:
    type: kafka
    host: kafka.internal:9092
```

The `variables` map holds any additional configuration your downstream consumers need — database names, schemas, buckets, prefixes. Values are always strings.

## ports

A port is a named data interface your product exposes. It references one or more connections and carries a JSON Schema describing the data's shape.

```yaml
ports:
  daily_metrics:
    description: Daily aggregated customer metrics
    connections:
      - connection: "#/connections/analytics_db"
        location: customer_daily_metrics
    schema:
      type: object
      properties:
        customer_id: { type: string }
        date: { type: string, format: date }
        total_orders: { type: integer }
      required: [customer_id, date, total_orders]
```

**Location syntax** depends on the connection type:

| Connection type | Location format | Example |
|---|---|---|
| postgresql / mysql | `schema.table` or `table` | `public.orders` |
| kafka | topic name | `customer-events` |
| s3 | `bucket/path/` | `analytics/metrics/` |
| http | path | `/api/v1/customers` |

Schemas can also reference a separate file to keep the spec readable:

```yaml
ports:
  users:
    schema:
      $ref: schemas/users.schema.yaml
```

## components

Reusable schema definitions go under `components.schemas`. Reference them from any port with `$ref`:

```yaml
components:
  schemas:
    Address:
      type: object
      properties:
        street: { type: string }
        city: { type: string }
        country: { type: string }

ports:
  customers:
    schema:
      type: object
      properties:
        id: { type: string }
        address:
          $ref: "#/components/schemas/Address"
```

## Schemas

Daco uses **JSON Schema Draft 2020-12** for all schema definitions. The translator understands the full range of keywords:

- Scalar types: `string`, `integer`, `number`, `boolean`, `null`
- Composite: `object` with `properties`, `required`, `additionalProperties`
- Arrays: `array` with `items`
- Composition: `allOf`, `anyOf`, `oneOf`
- Constraints: `enum`, `pattern`, `minimum`, `maximum`, `minLength`, `maxLength`
- Definitions: `$defs` and `$ref`

## Full example

```yaml
opendpi: "1.0.0"

info:
  title: Order Pipeline
  version: "1.3.0"

connections:
  warehouse:
    type: postgresql
    host: dwh.example.com
    variables:
      database: production
      schema: orders

ports:
  raw_orders:
    description: Raw order events from the ingestion layer
    connections:
      - connection: "#/connections/warehouse"
        location: raw_orders
    schema:
      $ref: schemas/raw_orders.schema.yaml

  order_summary:
    description: Aggregated daily order totals
    connections:
      - connection: "#/connections/warehouse"
        location: order_summary
    schema:
      type: object
      properties:
        date: { type: string, format: date }
        total_revenue: { type: number }
        order_count: { type: integer }
      required: [date, total_revenue, order_count]
```

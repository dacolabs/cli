---
title: Ports
description: How to define data interfaces, write schemas, and use $ref to keep your spec organized.
order: 5
section: Guides
---

## What is a port?

A port is a named data interface your product exposes. It could be a database table, a Kafka topic, an S3 path, or an HTTP endpoint. Every port has a name, an optional description, one or more connection bindings, and a schema that describes the shape of the data.

## Adding a port

Run the interactive form:

```bash
daco ports add
```

Or pass everything as flags:

```bash
daco ports add \
  --name orders \
  --description "Live order records" \
  --connection warehouse \
  --location public.orders
```

This creates two things:

1. An entry in `spec/opendpi.yaml`
2. A blank schema file at `spec/schemas/orders.schema.yaml`

## Writing a schema

Open the generated schema file and define your port's data shape using JSON Schema:

```yaml
# spec/schemas/orders.schema.yaml
type: object
properties:
  order_id:
    type: string
  customer_id:
    type: string
  placed_at:
    type: string
    format: date-time
  total_cents:
    type: integer
  status:
    type: string
    enum: [pending, confirmed, shipped, cancelled]
required:
  - order_id
  - customer_id
  - placed_at
  - total_cents
  - status
```

The schema file is referenced from the spec via `$ref`:

```yaml
ports:
  orders:
    description: Live order records
    connections:
      - connection: "#/connections/warehouse"
        location: public.orders
    schema:
      $ref: schemas/orders.schema.yaml
```

You can also write the schema inline directly in `opendpi.yaml` — useful for simple ports with only a few fields.

## Supported JSON Schema keywords

Daco's translator understands the full range of JSON Schema Draft 2020-12:

**Types:** `string`, `integer`, `number`, `boolean`, `null`

**Objects:**

```yaml
type: object
properties:
  name: { type: string }
required: [name]
additionalProperties: false
```

**Arrays:**

```yaml
type: array
items:
  type: string
```

**Composition:**

```yaml
allOf:
  - $ref: "#/components/schemas/BaseEvent"
  - type: object
    properties:
      payload: { type: string }
```

**Constraints:**

```yaml
type: string
minLength: 1
maxLength: 255
pattern: "^[a-z0-9_]+$"
```

```yaml
type: integer
minimum: 0
maximum: 1000000
```

**Enums:**

```yaml
type: string
enum: [active, inactive, archived]
```

## Using $ref and $defs

For schemas with repeated sub-structures, use `$defs` to define them once and `$ref` to reference them:

```yaml
type: object
properties:
  billing_address:
    $ref: "#/$defs/Address"
  shipping_address:
    $ref: "#/$defs/Address"
$defs:
  Address:
    type: object
    properties:
      street: { type: string }
      city: { type: string }
      country: { type: string }
    required: [street, city, country]
```

You can also reference schemas from `components.schemas` in `opendpi.yaml`:

```yaml
schema:
  $ref: "#/components/schemas/Order"
```

## Location syntax

The `location` inside a connection binding points to the specific resource within that infrastructure endpoint:

```yaml
connections:
  - connection: "#/connections/warehouse"
    location: public.orders        # schema.table for SQL databases
```

```yaml
connections:
  - connection: "#/connections/kafka_prod"
    location: order-events         # topic name for Kafka
```

```yaml
connections:
  - connection: "#/connections/s3_data"
    location: analytics/orders/    # bucket prefix for S3
```

```yaml
connections:
  - connection: "#/connections/orders_api"
    location: /api/v2/orders       # path for HTTP
```

## Viewing and managing ports

```bash
daco ports list
```

```bash
daco ports describe orders
```

Displays the port's connections, location bindings, and a preview of the schema.

## Removing a port

```bash
daco ports remove orders
```

Use `--force` to skip the confirmation prompt. The schema file at `spec/schemas/orders.schema.yaml` is not deleted automatically — remove it manually if you no longer need it.

## Nested objects and arrays

Schemas can be arbitrarily nested. The translator handles all depth levels:

```yaml
type: object
properties:
  id: { type: string }
  line_items:
    type: array
    items:
      type: object
      properties:
        sku: { type: string }
        quantity: { type: integer }
        unit_price_cents: { type: integer }
      required: [sku, quantity, unit_price_cents]
required: [id, line_items]
```

This translates cleanly to nested PySpark `StructType`, Pydantic models, Go structs, Protobuf messages, and all other supported formats.

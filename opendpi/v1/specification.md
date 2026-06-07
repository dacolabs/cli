# OpenDPI v1 Specification

This document is the complete reference for the OpenDPI v1 specification.

## Document Structure

An OpenDPI document has the following structure:

```yaml
opendpi: "1.0.0"        # Required - spec version
info: { ... }           # Required - product metadata
connections: { ... }    # Required - infrastructure connections
ports: { ... }          # Required - data interfaces
components: { ... }     # Optional - reusable definitions
```

## Schema

The JSON Schema for validation:
- **Bundled**: `schema/1.0.0/opendpi.json`
- **$id**: `https://opendpi.org/schema/1.0.0`

---

## Root Fields

### opendpi

| Property | Value |
|----------|-------|
| Type | `string` |
| Required | Yes |
| Allowed Values | `"1.0.0"` |

The version of the OpenDPI specification this document conforms to.

```yaml
opendpi: "1.0.0"
```

---

## Info Object

Metadata about the data product.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | Yes | The name of the data product |
| `version` | string | Yes | The version of the data product |
| `description` | string | No | A brief description |

### Example

```yaml
info:
  title: Customer Analytics
  version: "2.1.0"
  description: Aggregated customer behavior metrics
```

---

## Connections Object

A registry of infrastructure connections. This is a map where keys are connection identifiers and values are Connection Objects.

```yaml
connections:
  <connection_id>: <Connection Object>
  <connection_id>: <Connection Object>
```

### Connection Object

A connection map value is either an **inline** definition or a **reference**
to an external connection file (`$ref`). The two forms are mutually exclusive.

#### Inline form

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Type identifier (user-defined) |
| `host` | string | Yes | Addressable host, broker list, or base URL |
| `description` | string | No | Human-readable description |
| `variables` | object | No | Type-specific key-value configuration |

#### Ref form

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `$ref` | string | Yes | Relative or absolute file path to an external connection definition (YAML or JSON). The referenced file must itself match the inline form. |

Example showing both forms side by side:

```yaml
connections:
  warehouse:
    type: postgresql
    host: warehouse.db.example.com:5432

  shared_events:
    $ref: ../connections/events.yaml
```

The `$ref` form is useful when multiple data products share the same
infrastructure: each product references a single source-of-truth connection
file rather than duplicating its definition.

#### Type

The `type` field is a user-defined string that identifies the type of infrastructure. OpenDPI does not enforce a specific list - use whatever identifier makes sense for your tooling.

Common examples:
- `postgresql`, `mysql`, `mongodb` - Databases
- `kafka`, `rabbitmq`, `pulsar` - Message brokers
- `s3`, `gcs`, `azure-blob` - Object storage
- `http`, `grpc` - APIs
- Custom types for internal systems

#### Variables

The `variables` field holds type-specific configuration as key-value pairs. Values can be strings, numbers, or booleans.

### Example

```yaml
connections:
  analytics_db:
    type: postgresql
    host: analytics.db.example.com:5432
    description: Primary analytics database
    variables:
      database: analytics
      schema: public
      sslmode: require

  event_stream:
    type: kafka
    host: kafka-1.example.com:9092,kafka-2.example.com:9092
    variables:
      security.protocol: SASL_SSL
```

---

## Ports Object

A registry of data output interfaces. This is a map where keys are port identifiers and values are Port Objects.

```yaml
ports:
  <port_id>: <Port Object>
  <port_id>: <Port Object>
```

### Port Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | No | What data this port provides |
| `connections` | array | Yes | Array of connection-location pairs |
| `schema` | Schema or Reference | Yes | Shape of the data |

#### Connections Array

An array of connection-location pairs where this port's data can be accessed. Each element contains:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `connection` | string | Yes | Reference to a connection (e.g., `#/connections/my_db`) |
| `location` | string | Yes | Specific pointer within the connection |

```yaml
connections:
  - connection: "#/connections/<connection_id>"
    location: <location_string>
```

#### Location

The `location` field identifies where the data is within the connection. The format depends on the type:

| Type | Location Example |
|----------|------------------|
| postgresql | `schema.table_name` or `table_name` |
| kafka | `topic-name` |
| s3 | `bucket/path/to/data/` |
| http | `/api/v1/resource` |

#### Schema

The schema can be:
1. **Inline JSON Schema** - Define the schema directly
2. **Reference** - Point to a schema in `components/schemas`

```yaml
# Inline
schema:
  type: object
  properties:
    id: { type: integer }

# Reference
schema:
  $ref: "#/components/schemas/User"
```

### Example

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
        customer_id:
          type: string
        date:
          type: string
          format: date
        total_orders:
          type: integer

  user_events:
    description: Real-time user activity events
    connections:
      - connection: "#/connections/event_stream"
        location: user-events-v1
    schema:
      $ref: "#/components/schemas/UserEvent"

  # Port available at multiple locations
  user_profiles:
    description: User profiles available via database and API
    connections:
      - connection: "#/connections/analytics_db"
        location: users
      - connection: "#/connections/api"
        location: /v1/users
    schema:
      $ref: "#/components/schemas/User"
```

---

## Components (Optional)

Reusable definitions that can be referenced throughout the document.

### Schemas

A registry of reusable JSON Schema definitions.

```yaml
components:
  schemas:
    <schema_id>: <JSON Schema>
```

Reference schemas using `$ref`:

```yaml
schema:
  $ref: "#/components/schemas/<schema_id>"
```

### Example

```yaml
components:
  schemas:
    User:
      type: object
      required:
        - id
        - email
      properties:
        id:
          type: integer
        email:
          type: string
          format: email
        name:
          type: string

    UserEvent:
      type: object
      properties:
        user_id:
          type: integer
        event_type:
          type: string
        timestamp:
          type: string
          format: date-time
```

---

## JSON Schema Reference

OpenDPI uses [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12/json-schema-core.html) for schema definitions.

Common schema keywords:

| Keyword | Description |
|---------|-------------|
| `type` | Data type (object, array, string, number, integer, boolean, null) |
| `properties` | Object properties |
| `required` | Required property names |
| `items` | Array item schema |
| `format` | Semantic format (date, date-time, email, uri, etc.) |
| `enum` | Allowed values |
| `$ref` | Reference to another schema |

---

## Full Example

```yaml
opendpi: "1.0.0"

info:
  title: E-Commerce Analytics
  version: "3.0.0"
  description: Analytics data products for the e-commerce platform

connections:
  warehouse:
    type: postgresql
    host: warehouse.db.example.com:5432
    variables:
      database: analytics
      schema: public

  streams:
    type: kafka
    host: kafka.example.com:9092

  datalake:
    type: s3
    host: s3.amazonaws.com
    variables:
      bucket: analytics-data

ports:
  order_summary:
    description: Daily order summaries by region
    connections:
      - connection: "#/connections/warehouse"
        location: order_daily_summary
      - connection: "#/connections/datalake"
        location: orders/daily_summary/
    schema:
      $ref: "#/components/schemas/OrderSummary"

  order_events:
    description: Real-time order events
    connections:
      - connection: "#/connections/streams"
        location: orders-v2
    schema:
      $ref: "#/components/schemas/OrderEvent"

components:
  schemas:
    OrderSummary:
      type: object
      properties:
        date:
          type: string
          format: date
        region:
          type: string
        total_orders:
          type: integer
        total_revenue:
          type: number

    OrderEvent:
      type: object
      properties:
        order_id:
          type: string
        event_type:
          type: string
          enum: [created, updated, completed, cancelled]
        timestamp:
          type: string
          format: date-time
```

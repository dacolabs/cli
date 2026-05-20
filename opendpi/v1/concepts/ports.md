# Ports

Ports are the data output interfaces exposed by your data product. Each port represents a specific dataset that consumers can access. A port can be available at multiple connection-location pairs, allowing the same data to be accessed through different infrastructure endpoints.

## Structure

```yaml
ports:
  <port_id>:
    description: <string>    # Optional - what this port provides
    connections:             # Required - array of connection-location pairs
      - connection: "#/connections/<connection_id>"  # Required - reference to a connection
        location: <string>   # Required - where within the connection
    schema: <object>         # Required - shape of the data
```

## Port Object

### description

A human-readable description of what data this port provides. While optional, it's highly recommended for discoverability.

```yaml
ports:
  daily_metrics:
    description: Daily aggregated customer metrics, updated at 2am UTC
```

### connections

An array of connection-location pairs where this port's data can be accessed. Each element contains:

- **connection**: A string reference to a connection defined in the `connections` section using JSON Pointer syntax
- **location**: The specific pointer within that connection where the data resides

```yaml
connections:
  - connection: "#/connections/analytics_db"
    location: public.metrics
```

The reference must match a key in the connections object:

```yaml
connections:
  analytics_db:        # <-- Referenced here
    type: postgresql
    host: db.example.com

ports:
  metrics:
    connections:
      - connection: "#/connections/analytics_db"  # <-- Points to above
        location: public.metrics
```

### location

The specific pointer within the connection where the data resides. The format depends on the type:

| Type | Location Examples |
|----------|-------------------|
| postgresql | `table_name`, `schema.table_name` |
| kafka | `topic-name` |
| s3 | `path/to/data/`, `prefix/*.parquet` |
| http | `/api/v1/users`, `/graphql` |
| mongodb | `collection_name`, `database.collection` |

```yaml
ports:
  users:
    connections:
      - connection: "#/connections/db"
        location: public.users

  events:
    connections:
      - connection: "#/connections/kafka"
        location: user-events-v2

  reports:
    connections:
      - connection: "#/connections/s3"
        location: reports/daily/
```

### schema

The shape of the data at this port. Can be either:

1. **Inline JSON Schema** - Define the schema directly
2. **Reference** - Point to a schema in `components/schemas`

**Inline schema:**

```yaml
ports:
  users:
    connections:
      - connection: "#/connections/db"
        location: users
    schema:
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
        created_at:
          type: string
          format: date-time
```

**Referenced schema:**

```yaml
ports:
  users:
    connections:
      - connection: "#/connections/db"
        location: users
    schema:
      $ref: "#/components/schemas/User"

components:
  schemas:
    User:
      type: object
      properties:
        id: { type: integer }
        email: { type: string, format: email }
```

## Multiple Ports

A data product typically exposes multiple ports:

```yaml
ports:
  # Raw data
  raw_events:
    description: Raw user activity events
    connections:
      - connection: "#/connections/kafka"
        location: raw-user-events
    schema: { ... }

  # Processed data
  daily_summary:
    description: Daily aggregated metrics
    connections:
      - connection: "#/connections/warehouse"
        location: daily_summary
    schema: { ... }

  # API access
  user_profile:
    description: User profile data via REST API
    connections:
      - connection: "#/connections/api"
        location: /users/{id}
    schema: { ... }
```

## Multiple Connections per Port

A single port can be available at multiple connection-location pairs. This is useful when the same data is replicated across different infrastructure or available through different types:

```yaml
connections:
  warehouse:
    type: postgresql
    host: warehouse.example.com
  datalake:
    type: s3
    host: s3.amazonaws.com
    variables:
      bucket: analytics-data

ports:
  users:
    description: User data available in both warehouse and data lake
    connections:
      - connection: "#/connections/warehouse"
        location: public.users
      - connection: "#/connections/datalake"
        location: users/parquet/
    schema: { ... }
```

## Ports Sharing Connections

Multiple ports can reference the same connection when they represent different datasets in the same infrastructure:

```yaml
connections:
  warehouse:
    type: postgresql
    host: warehouse.example.com

ports:
  users:
    connections:
      - connection: "#/connections/warehouse"
        location: public.users
    schema: { ... }

  orders:
    connections:
      - connection: "#/connections/warehouse"
        location: public.orders
    schema: { ... }

  products:
    connections:
      - connection: "#/connections/warehouse"
        location: public.products
    schema: { ... }
```

## Schema Design

### Keep Schemas Focused

Each port should have a well-defined schema. Avoid catch-all schemas:

```yaml
# Good - specific schema
schema:
  type: object
  properties:
    user_id: { type: string }
    event_type: { type: string }
    timestamp: { type: string, format: date-time }

# Avoid - too generic
schema:
  type: object
  additionalProperties: true
```

### Use References for Shared Schemas

When multiple ports share the same schema or embed common objects, use `$ref`:

```yaml
ports:
  created_events:
    connections:
      - connection: "#/connections/kafka"
        location: user-created-events
    schema:
      $ref: "#/components/schemas/UserEvent"

  updated_events:
    connections:
      - connection: "#/connections/kafka"
        location: user-updated-events
    schema:
      $ref: "#/components/schemas/UserEvent"

components:
  schemas:
    UserEvent:
      type: object
      properties:
        user_id: { type: string }
        event_type: { type: string }
        timestamp: { type: string, format: date-time }
```

### Document Formats

Use JSON Schema's `format` keyword for semantic types:

```yaml
schema:
  type: object
  properties:
    id:
      type: string
      format: uuid
    email:
      type: string
      format: email
    created_at:
      type: string
      format: date-time
    website:
      type: string
      format: uri
```

## Examples

### Database Table

```yaml
ports:
  customer_orders:
    description: Customer order history with line items
    connections:
      - connection: "#/connections/orders_db"
        location: orders.customer_orders
    schema:
      type: object
      required:
        - order_id
        - customer_id
        - created_at
      properties:
        order_id:
          type: string
          format: uuid
        customer_id:
          type: string
        status:
          type: string
          enum: [pending, confirmed, shipped, delivered, cancelled]
        total_amount:
          type: number
        created_at:
          type: string
          format: date-time
```

### Kafka Topic

```yaml
ports:
  user_activity:
    description: Real-time user activity events
    connections:
      - connection: "#/connections/event_bus"
        location: user-activity-v2
    schema:
      type: object
      properties:
        event_id:
          type: string
          format: uuid
        user_id:
          type: string
        action:
          type: string
        properties:
          type: object
          additionalProperties: true
        timestamp:
          type: string
          format: date-time
```

### S3 Data Files

```yaml
ports:
  daily_reports:
    description: Daily sales reports in Parquet format
    connections:
      - connection: "#/connections/data_lake"
        location: reports/daily/sales/
    schema:
      type: object
      properties:
        date:
          type: string
          format: date
        region:
          type: string
        total_sales:
          type: number
        order_count:
          type: integer
```

### Multi-Location Port

```yaml
ports:
  user_profiles:
    description: User profiles available via API and data lake
    connections:
      - connection: "#/connections/api"
        location: /v1/users
      - connection: "#/connections/data_lake"
        location: users/profiles/
    schema:
      type: object
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
          format: email
        name:
          type: string
        created_at:
          type: string
          format: date-time
```

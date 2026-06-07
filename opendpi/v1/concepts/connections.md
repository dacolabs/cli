# Connections

Connections define where your data infrastructure lives. They form a registry that ports can reference.

## Structure

A connection can be defined inline:

```yaml
connections:
  <connection_id>:
    type: <string>          # Required - type identifier
    host: <string>          # Required - addressable endpoint
    description: <string>   # Optional - human-readable description
    variables: <object>     # Optional - type-specific config
```

…or as a reference to an external connection file:

```yaml
connections:
  <connection_id>:
    $ref: <path>            # Required - relative or absolute file path
```

The `$ref` form is mutually exclusive with the inline fields. See
[Sharing connections across products](#sharing-connections-across-products)
for when to reach for it.

## Connection Object

### type

The `type` field identifies the type of infrastructure. This is a **user-defined string** - OpenDPI does not enforce a specific list.

Use whatever identifier makes sense for your organization and tooling. The type helps consumers understand how to connect and what to expect.

**Common patterns:**

| Category | Examples |
|----------|----------|
| Databases | `postgresql`, `mysql`, `mongodb`, `redis` |
| Message Brokers | `kafka`, `rabbitmq`, `pulsar`, `nats` |
| Object Storage | `s3`, `gcs`, `azure-blob`, `minio` |
| APIs | `http`, `grpc`, `graphql` |
| Data Warehouses | `bigquery`, `redshift`, `databricks` |
| Custom | `internal-api`, `legacy-system`, `custom-type` |

**Example:**

```yaml
connections:
  analytics:
    type: postgresql
    host: analytics.db.example.com

  events:
    type: kafka
    host: kafka.example.com:9092

  files:
    type: s3
    host: s3.us-east-1.amazonaws.com
```

### host

The addressable endpoint for the connection. Format depends on the type:

| Type | Host Format |
|----------|-------------|
| Databases | `hostname:port` or `hostname` |
| Kafka | `broker1:port,broker2:port` (comma-separated) |
| S3 | `s3.<region>.amazonaws.com` or `bucket.s3.<region>.amazonaws.com` |
| HTTP | `https://api.example.com` |

### variables

Type-specific configuration as key-value pairs. Values can be strings, numbers, or booleans.

**Why variables?**

Different types need different configuration. Instead of defining a rigid structure for every possible type, `variables` provides flexibility.

**Common variable patterns:**

```yaml
# PostgreSQL
variables:
  database: analytics
  schema: public
  sslmode: require

# Kafka
variables:
  security.protocol: SASL_SSL
  sasl.mechanism: PLAIN

# S3
variables:
  bucket: my-data-bucket
  region: us-east-1
  prefix: data/

# HTTP API
variables:
  basePath: /api/v2
  timeout: 30
```

## Multiple Connections

A data product can have multiple connections when data spans different systems:

```yaml
connections:
  primary_db:
    type: postgresql
    host: primary.db.example.com
    variables:
      database: production

  replica_db:
    type: postgresql
    host: replica.db.example.com
    variables:
      database: production

  cache:
    type: redis
    host: cache.example.com:6379

  events:
    type: kafka
    host: kafka-1.example.com:9092,kafka-2.example.com:9092
```

## Referencing Connections

Ports reference connections using a string with JSON Pointer syntax:

```yaml
connections:
  my_db:
    type: postgresql
    host: db.example.com

ports:
  users:
    connections:
      - connection: "#/connections/my_db"
        location: users
    schema: { ... }
```

The connection reference must point to a connection defined in the same document.

## Sharing connections across products

When several data products read from the same infrastructure, repeating the
connection definition in every document drifts over time. Pull the connection
into its own file and reference it with `$ref`:

```
my-project/
├── connections/
│   └── warehouse.yaml          # single source of truth
└── products/
    ├── orders/opendpi.yaml     # references ../../connections/warehouse.yaml
    └── customers/opendpi.yaml  # references ../../connections/warehouse.yaml
```

The standalone connection file follows the same shape as an inline connection:

```yaml
# connections/warehouse.yaml
type: postgresql
host: warehouse.db.example.com:5432
variables:
  database: analytics
  schema: public
```

And each product references it from its own `connections:` map:

```yaml
# products/orders/opendpi.yaml
connections:
  warehouse:
    $ref: ../../connections/warehouse.yaml
```

Paths are resolved relative to the directory of the file containing the `$ref`.

## Design Considerations

### Connection vs Port

- **Connection**: The infrastructure endpoint (database, broker, storage)
- **Port**: The specific data within that infrastructure (table, topic, path)

One connection can serve multiple ports:

```yaml
connections:
  warehouse:
    type: postgresql
    host: warehouse.example.com

ports:
  users:
    connections:
      - connection: "#/connections/warehouse"
        location: users
    schema: { ... }

  orders:
    connections:
      - connection: "#/connections/warehouse"
        location: orders
    schema: { ... }

  products:
    connections:
      - connection: "#/connections/warehouse"
        location: products
    schema: { ... }
```

### Security Considerations

OpenDPI documents describe **interfaces**, not credentials. Do not include:

- Passwords or secrets
- API keys
- Connection strings with embedded credentials

Authentication details should be managed separately through your infrastructure's secret management.

## Examples

### Database Connection

```yaml
connections:
  analytics_db:
    type: postgresql
    host: analytics.db.example.com:5432
    description: Primary analytics database (read replicas available)
    variables:
      database: analytics
      schema: public
      sslmode: verify-full
```

### Kafka Connection

```yaml
connections:
  event_bus:
    type: kafka
    host: kafka-1.example.com:9092,kafka-2.example.com:9092,kafka-3.example.com:9092
    description: Production Kafka cluster
    variables:
      security.protocol: SASL_SSL
      sasl.mechanism: SCRAM-SHA-256
```

### S3 Connection

```yaml
connections:
  data_lake:
    type: s3
    host: s3.us-east-1.amazonaws.com
    description: Raw data lake storage
    variables:
      bucket: company-data-lake
      region: us-east-1
```

### HTTP API Connection

```yaml
connections:
  internal_api:
    type: http
    host: https://api.internal.example.com
    description: Internal REST API
    variables:
      basePath: /v2
      timeout: 30
```

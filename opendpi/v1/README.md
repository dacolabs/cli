# OpenDPI v1

This is the documentation for OpenDPI version 1.x.

## Getting Started

### What is OpenDPI?

OpenDPI (Open Data Product Interface) is a specification for describing data product interfaces. It allows you to define:

- **What data** your product exposes (ports)
- **Where the data lives** (connections)
- **What the data looks like** (schemas)

### Your First OpenDPI Document

Create a file named `data-product.yaml`:

```yaml
opendpi: "1.0.0"

info:
  title: My First Data Product
  version: "1.0.0"
  contact:
    name: Your Name
    email: you@example.com

connections:
  my_database:
    type: postgresql
    host: localhost:5432
    variables:
      database: mydb

ports:
  users:
    description: User account data
    connections:
      - connection: "#/connections/my_database"
        location: users
    schema:
      type: object
      properties:
        id:
          type: integer
        email:
          type: string
          format: email
        created_at:
          type: string
          format: date-time
```

### Breaking It Down

#### 1. Version Declaration

```yaml
opendpi: "1.0.0"
```

Every OpenDPI document starts with the spec version.

#### 2. Info Section

```yaml
info:
  title: My First Data Product
  version: "1.0.0"
  contact:
    name: Your Name
```

Metadata about your data product. The `version` here is your product's version, not the spec version.

#### 3. Connections

```yaml
connections:
  my_database:
    type: postgresql
    host: localhost:5432
```

A registry of infrastructure connections. Each connection has:
- A unique name (key)
- A `type` identifier
- A `host` address
- Optional `variables` for type-specific config

#### 4. Ports

```yaml
ports:
  users:
    connections:
      - connection: "#/connections/my_database"
        location: users
    schema: { ... }
```

Ports are the data interfaces. Each port has a `connections` array that specifies:
- Which connection to use
- Where to find the data within that connection (`location`)
- What the data looks like (`schema`)

## Documentation

| Document | Description |
|----------|-------------|
| [Specification](./specification.md) | Complete reference for all fields |
| [Concepts](./concepts/) | Deep dives into key concepts |
| [Examples](./examples/) | Real-world examples in JSON and YAML |

### Concepts

- [Overview](./concepts/overview.md) - What is OpenDPI and why use it
- [Connections](./concepts/connections.md) - Understanding connections and types
- [Ports](./concepts/ports.md) - Defining data interfaces
- [Schemas](./concepts/schemas.md) - Working with JSON Schema

## Validating Your Document

Use any JSON Schema validator with the OpenDPI schema:

```bash
# Schema URL
https://opendpi.org/schema/1.0.0

# Or use the local schema file
schema/1.0.0/opendpi.json
```

## Next Steps

1. Read the [Specification](./specification.md) for complete field reference
2. Explore [Examples](./examples/) for common patterns
3. Learn about [Connections](./concepts/connections.md) and types

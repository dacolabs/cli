# OpenDPI

**The open standard for defining Data Product Interfaces.**

OpenDPI is to Data Products what OpenAPI is to REST APIs.

## The Problem

Data products are everywhere, but how do you describe them?

- **Inconsistent documentation** - Every team describes their data differently
- **Discovery is hard** - No standard way to find what data a product exposes
- **Integration friction** - Connecting to data products requires tribal knowledge
- **Tooling fragmentation** - Each platform has its own metadata format

## The Solution

OpenDPI provides a simple, vendor-neutral specification to:

- **Describe** what data your product exposes
- **Document** how to connect and access that data
- **Define** the exact shape of your data with JSON Schema
- **Enable** tooling, catalogs, and automation

## Quick Example

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
        customer_id:
          type: string
        date:
          type: string
          format: date
        total_orders:
          type: integer
        revenue:
          type: number
```

## Key Concepts

| Concept | Description |
|---------|-------------|
| **Connections** | Where your data lives - any type, any infrastructure |
| **Ports** | What data you expose - the interface to your data product |
| **Components** | Reusable schema definitions for DRY documents |

## Features

- **Type agnostic** - Works with any data infrastructure (databases, streams, object storage, APIs)
- **JSON Schema based** - Familiar, well-tooled, widely supported
- **Reference support** - Use `$ref` to share schemas across ports and connections across products
- **Lightweight** - Minimal required fields, easy to get started
- **Extensible** - Custom variables per connection for any type

## Documentation

| Version | Status | Docs |
|---------|--------|------|
| [v1](./v1/) | Current | [Getting Started](./v1/README.md) · [Specification](./v1/specification.md) · [Schema](./schema/1.0.0/) |

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on proposing changes to the specification.

## License

[Apache 2.0](../LICENSE)

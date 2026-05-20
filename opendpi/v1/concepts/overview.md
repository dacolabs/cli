# Overview

## What is a Data Product?

A data product is a self-contained unit of data that serves a specific purpose. Unlike raw data dumps or ad-hoc queries, a data product is:

- **Discoverable** - Others can find and understand it
- **Addressable** - It has a stable location and interface
- **Trustworthy** - It has clear ownership and quality guarantees
- **Self-describing** - Its structure and meaning are documented

Data products can be:
- A table in a database
- A topic in a message broker
- Files in object storage
- An API endpoint
- Or any combination of the above

## What is a Data Product Interface?

The **interface** is how consumers interact with your data product. It answers:

- What data is available?
- Where can I find it?
- What does it look like?
- Who owns it?

OpenDPI provides a standard way to describe this interface.

## Why OpenDPI?

### The Problem

Without a standard, every team describes their data differently:

```
Team A: "The customer data is in Postgres, table users, ask John for access"
Team B: "Check the wiki for the schema, but it might be outdated"
Team C: "Here's a Confluence page with screenshots from 2021"
```

This leads to:
- Hours spent finding and understanding data
- Tribal knowledge that doesn't scale
- Integration errors from misunderstood schemas
- Stale documentation

### The Solution

OpenDPI provides a machine-readable, standardized format:

```yaml
opendpi: "1.0.0"
info:
  title: Customer Data
  version: "1.0.0"
connections:
  db:
    type: postgresql
    host: prod.db.example.com
ports:
  users:
    connections:
      - connection: "#/connections/db"
        location: users
    schema:
      type: object
      properties:
        id: { type: integer }
        email: { type: string, format: email }
```

Now anyone (or any tool) can:
- Discover what data exists
- Understand how to connect
- Know the exact schema

## Comparison to Other Standards

| Standard | Purpose | Relationship to OpenDPI |
|----------|---------|------------------------|
| **OpenAPI** | REST API interfaces | OpenDPI is similar but for data products |
| **AsyncAPI** | Event-driven APIs | OpenDPI borrows the "connections" pattern |
| **JSON Schema** | Data structure validation | OpenDPI uses JSON Schema for schemas |
| **Data Contracts** | SLAs and quality | Complementary - could reference OpenDPI |

OpenDPI focuses specifically on **what data is exposed and how to access it**. It doesn't cover:
- Data quality SLAs
- Access control policies
- Lineage or provenance
- Processing pipelines

These concerns are complementary and can reference an OpenDPI document.

## Core Concepts

### Connections

A **connection** describes where data infrastructure lives:

```yaml
connections:
  my_database:
    type: postgresql
    host: db.example.com
```

See [Connections](./connections.md) for details.

### Ports

A **port** is a data interface - a specific output from your data product:

```yaml
ports:
  users:
    connections:
      - connection: "#/connections/my_database"
        location: users_table
    schema: { ... }
```

See [Ports](./ports.md) for details.

### Schemas

**Schemas** define the shape of your data using JSON Schema:

```yaml
schema:
  type: object
  properties:
    id: { type: integer }
    name: { type: string }
```

See [Schemas](./schemas.md) for details.

## When to Use OpenDPI

Use OpenDPI when you want to:

- **Document data products** in a standard, machine-readable way
- **Enable discovery** through data catalogs
- **Validate data** against a contract
- **Generate code** from data definitions
- **Automate integration** between systems

## Next Steps

- [Getting Started](../README.md) - Create your first OpenDPI document
- [Specification](../specification.md) - Complete field reference
- [Examples](../examples/) - Real-world examples

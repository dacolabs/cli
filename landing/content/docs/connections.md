---
title: Connections
description: How to define, manage, and reference infrastructure connections in your data product spec.
order: 4
section: Guides
---

## What is a connection?

A connection describes a single infrastructure endpoint where data in your product can be found. It has a name, a type, and a host — and optionally a set of key-value variables for any additional configuration consumers might need.

Connections are declared once in the spec and referenced by ports. This separation means you can update an endpoint (say, a database host) in one place without touching every port that uses it.

## Defining a connection

Add connections interactively:

```bash
daco connections add
```

Or pass all values as flags:

```bash
daco connections add \
  --name analytics_db \
  --type postgresql \
  --host analytics.db.example.com \
  --description "Production analytics warehouse"
```

In your `opendpi.yaml`, this produces:

```yaml
connections:
  analytics_db:
    type: postgresql
    host: analytics.db.example.com
    description: Production analytics warehouse
```

## Connection types

The `type` field is a free-form string. There is no enumeration — use whatever label clearly communicates the technology to your team and downstream tools.

Common values:

| Type | Typical host format |
|---|---|
| `postgresql` | `hostname` or `hostname:port` |
| `mysql` | `hostname` or `hostname:port` |
| `kafka` | `broker:port` |
| `s3` | `s3.amazonaws.com` or region endpoint |
| `http` | base URL (e.g., `api.example.com`) |
| `bigquery` | GCP project ID |
| `snowflake` | account identifier |
| `databricks` | workspace hostname |

## Variables

Use `variables` for any extra configuration that consumers need to connect — database names, schema names, bucket names, prefixes, etc.

```yaml
connections:
  warehouse:
    type: postgresql
    host: dwh.example.com
    variables:
      database: production
      schema: public
      port: "5432"
```

Variables are always string values. They appear in `daco connections describe` output and are available to downstream tooling that reads the spec.

## Referencing a connection from a port

When you add a port and bind it to a connection, the spec stores a JSON Reference:

```yaml
ports:
  daily_metrics:
    connections:
      - connection: "#/connections/analytics_db"
        location: customer_daily_metrics
```

The `connection` value is always `"#/connections/<name>"`. The CLI generates this automatically — you should not need to write it by hand.

## Listing and inspecting connections

```bash
daco connections list
```

Shows all connections, their types, hosts, and how many ports reference each one.

```bash
daco connections describe analytics_db
```

Shows the full connection detail, including variables and the list of ports that use it.

## Removing a connection

```bash
daco connections remove analytics_db
```

Connections that are referenced by one or more ports cannot be removed. Remove or rebind those ports first, then delete the connection.

Use `--force` to skip the confirmation prompt:

```bash
daco connections remove unused_conn --force
```

## Multiple connections per port

A port can reference more than one connection — for example, if the same data is mirrored to two stores, or if the port spans a primary source and a replica:

```yaml
ports:
  events:
    connections:
      - connection: "#/connections/kafka_prod"
        location: customer-events
      - connection: "#/connections/s3_archive"
        location: events/customer/
```

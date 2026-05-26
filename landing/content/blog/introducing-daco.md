---
title: "Welcome to Daco"
date: "2026-01-30"
tag: "Announcement"
excerpt: "A new standard for data teams to align on definitions and ship data products faster, without rewriting their stack."
cover: "c3"
glyph: "daco"
---

Managing data products across teams is rarely a technical bottleneck. It's a synchronization bottleneck.

Data teams run into the same three problems, regardless of stack:

**Definition inconsistency.** The same concept — "active user," "revenue," "conversion" — is defined differently in the analytics warehouse, the ML feature store, and the reporting dashboard. Pipelines built on different definitions produce different numbers. Arguments about which number is "right" replace building.

**Manual translation.** Every time a schema changes, someone rewrites the PySpark StructType, updates the dbt YAML, regenerates the Pydantic model. This work is tedious, error-prone, and it compounds with every new downstream consumer.

**Brittleness at change.** When the source system changes — a renamed column, a new table, a migrated database — everything downstream breaks. Finding all the affected pipelines requires archaeology.

## What We Built

Daco is a development platform that treats data definitions as code.

You define your data products once, in OpenDPI format, alongside your pipelines. The Daco CLI uses those definitions to generate schemas for any framework — dbt, PySpark, Pydantic, Go, Avro, Protobuf. When a definition changes, you regenerate. Nothing drifts.

```bash
# Define once
daco init
daco ports add

# Generate for any target
daco ports translate --format pyspark my-port
daco ports translate --format pydantic my-port
daco ports translate --format dbt my-port
```

## Built on an Open Standard

Daco is built on OpenDPI — an open, vendor-neutral specification for data product interfaces. Think of it as OpenAPI for data products: a machine-readable format that describes what a data product exposes and how to connect to it.

OpenDPI is MIT-licensed. Your definitions aren't locked to any platform.

## An Invitation

We're building this in the open because a standard only works if it's shaped by the people using it. If you're a data engineer, data platform lead, or someone who has fought these problems before, we want to hear from you.

The best way to start: install the CLI, define a data product you're currently managing manually, and tell us where it breaks.

```bash
brew install dacolabs/tap/daco
```

We're at [daco@dacolabs.com](mailto:daco@dacolabs.com) and on [Discord](https://discord.gg/dacolabs).

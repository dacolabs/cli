---
title: Code Generation
description: Translate your OpenDPI schemas to PySpark, Pydantic, Go, Avro, Protobuf, SQL, and more.
order: 6
section: Guides
---

## Overview

`daco ports translate` reads the schema for one or more ports and writes ready-to-use code in your chosen format. You define your data shape once in JSON Schema; Daco handles every translation.

```bash
daco ports translate --all --format pyspark --output models/
```

Run without flags for the interactive form, which walks you through port selection, format, and output directory.

## All available formats

| Format | Language / Tool | File extension |
|---|---|---|
| `pyspark` | PySpark `StructType` | `.py` |
| `databricks-pyspark` | Databricks PySpark | `.py` |
| `pydantic` | Pydantic v2 model | `.py` |
| `python` | Python dataclass | `.py` |
| `gotypes` | Go struct | `.go` |
| `scala` | Scala case class | `.scala` |
| `spark-scala` | Spark Scala `StructType` | `.scala` |
| `databricks-scala` | Databricks Scala | `.scala` |
| `avro` | Avro schema | `.avsc` |
| `protobuf` | Protocol Buffers | `.proto` |
| `spark-sql` | Spark SQL `CREATE TABLE` | `.sql` |
| `databricks-sql` | Databricks SQL `CREATE TABLE` | `.sql` |
| `dqx-yaml` | DQX quality checks | `.yaml` |
| `markdown` | Readable schema table | `.md` |

## Examples

### PySpark

Given this schema for an `orders` port:

```yaml
type: object
properties:
  order_id: { type: string }
  customer_id: { type: string }
  total_cents: { type: integer }
  placed_at: { type: string, format: date-time }
required: [order_id, customer_id, total_cents, placed_at]
```

Running:

```bash
daco ports translate --name orders --format pyspark --output models/
```

Produces `models/orders.py`:

```python
from pyspark.sql.types import StructType, StructField, StringType, LongType, TimestampType

orders_schema = StructType([
    StructField("order_id", StringType(), False),
    StructField("customer_id", StringType(), False),
    StructField("total_cents", LongType(), False),
    StructField("placed_at", TimestampType(), False),
])
```

### Pydantic

```bash
daco ports translate --name orders --format pydantic --output models/
```

Produces `models/orders.py`:

```python
from pydantic import BaseModel
from datetime import datetime

class Orders(BaseModel):
    order_id: str
    customer_id: str
    total_cents: int
    placed_at: datetime
```

### Go

```bash
daco ports translate --name orders --format gotypes --output models/
```

Produces `models/orders.go`:

```go
package models

type Orders struct {
    OrderID    string `json:"order_id"`
    CustomerID string `json:"customer_id"`
    TotalCents int64  `json:"total_cents"`
    PlacedAt   string `json:"placed_at"`
}
```

### Protobuf

```bash
daco ports translate --name orders --format protobuf --output proto/
```

Produces `proto/orders.proto`:

```protobuf
syntax = "proto3";

package proto;

message Orders {
  string order_id = 1;
  string customer_id = 2;
  int64 total_cents = 3;
  string placed_at = 4;
}
```

### SQL (Spark / Databricks)

```bash
daco ports translate --name orders --format spark-sql --output ddl/
```

Produces `ddl/orders.sql`:

```sql
CREATE TABLE orders (
  order_id STRING NOT NULL,
  customer_id STRING NOT NULL,
  total_cents BIGINT NOT NULL,
  placed_at TIMESTAMP NOT NULL
);
```

## Translating all ports at once

```bash
daco ports translate --all --format pydantic --output models/
```

One file per port, all written to `models/`. File names match port names.

## Output directory

The `--output` flag sets the destination directory. For Go and Protobuf, it also becomes the package name in the generated code. The directory is created automatically if it does not exist.

```bash
daco ports translate --all --format gotypes --output internal/schema
```

## Committing generated code

Generated files are ordinary source files — commit them to your repository alongside your pipelines. When the spec changes, re-run translate and commit the diff. The changes are always minimal and reviewable.

## Type mapping reference

The translator maps JSON Schema types to the appropriate types in each target language. For formats not listed here, the mapping follows the same pattern: the most specific available type for the declared JSON Schema type and format.

| JSON Schema | PySpark | Pydantic | Go | Protobuf |
|---|---|---|---|---|
| `string` | `StringType` | `str` | `string` | `string` |
| `string` + `format: date` | `DateType` | `date` | `string` | `string` |
| `string` + `format: date-time` | `TimestampType` | `datetime` | `string` | `string` |
| `integer` | `LongType` | `int` | `int64` | `int64` |
| `number` | `DoubleType` | `float` | `float64` | `double` |
| `boolean` | `BooleanType` | `bool` | `bool` | `bool` |
| `array` | `ArrayType` | `list[...]` | `[]T` | `repeated` |
| `object` | `StructType` | nested model | nested struct | nested message |

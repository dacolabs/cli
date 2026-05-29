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
| `databricks-pyspark` | Databricks PySpark with column comments | `.py` |
| `pydantic` | Pydantic v2 model (runtime validation) | `.py` |
| `python` | Python dataclass | `.py` |
| `gotypes` | Go struct | `.go` |
| `scala` | Scala case class | `.scala` |
| `spark-scala` | Spark Scala `StructType` | `.scala` |
| `databricks-scala` | Databricks Scala with column comments | `.scala` |
| `avro` | Avro schema | `.avsc` |
| `protobuf` | Protocol Buffers (proto3) | `.proto` |
| `spark-sql` | Spark SQL `CREATE TABLE` | `.sql` |
| `databricks-sql` | Databricks SQL `CREATE TABLE … USING DELTA` with `CHECK` constraints | `.sql` |
| `dqx-yaml` | DQX row-level quality checks | `.yaml` |
| `markdown` | Readable schema table | `.md` |

## Constraint-driven translation

Daco doesn't just translate types — it translates JSON Schema's constraint vocabulary into each target's most specific representation. The same property, with different constraints, produces different output.

Take this single field:

```yaml
price:
  type: number
  multipleOf: 0.01
  maximum: 99999.99
```

| Target            | Output                                                         |
|-------------------|----------------------------------------------------------------|
| `spark-sql`       | `price DECIMAL(7, 2)`                                          |
| `databricks-sql`  | `price DECIMAL(7, 2) NOT NULL`                                 |
| `pyspark`         | `T.StructField("price", T.DecimalType(7, 2), nullable=False)`  |
| `spark-scala`     | `StructField("price", DecimalType(7, 2), nullable = false)`    |
| `scala`           | `price: BigDecimal`                                            |
| `pydantic`        | `price: float = Field(ge=0, le=99999.99, multiple_of=0.01)`    |
| `python`          | `price: float`                                                 |
| `gotypes`         | `Price float64`                                                |
| `protobuf`        | `double price`                                                 |
| `avro`            | `{"type": "bytes", "logicalType": "decimal", "precision": 7, "scale": 2}` |
| `markdown`        | `number (maximum: 99999.99, multipleOf: 0.01, narrows to: decimal(7,2))` |
| `dqx-yaml`        | range check + multiple_of check                                |

`enum`, `const`, integer bounds, `maxLength`, `pattern`, `format`, and `additionalProperties` all follow the same shape.

### Integer narrowing

| JSON Schema range          | spark-sql / databricks-sql | pyspark / spark-scala | scala / gotypes (signed) | gotypes (unsigned) | protobuf       |
|----------------------------|----------------------------|-----------------------|--------------------------|--------------------|----------------|
| `-128 ≤ x ≤ 127`           | `TINYINT`                  | `ByteType`            | `Byte` / `int8`          | `uint8`            | `int32` / `uint32` |
| `-32768 ≤ x ≤ 32767`       | `SMALLINT`                 | `ShortType`           | `Short` / `int16`        | `uint16`           | `int32` / `uint32` |
| 32-bit range               | `INT`                      | `IntegerType`         | `Int` / `int32`          | `uint32`           | `int32` / `uint32` |
| Unbounded / wider          | `BIGINT`                   | `LongType`            | `Long` / `int64`         | `uint64`           | `int64`        |

Unsigned variants kick in when `minimum >= 0`.

### Decimal narrowing

When `multipleOf` is a decimal fraction (`0.01`, `0.001`, …), targets with a fixed-point type promote `number` to it. Scale comes from the digits after the decimal point in `multipleOf`; precision comes from the bounds (or a sensible default).

- `spark-sql` / `databricks-sql`: `DECIMAL(precision, scale)`
- `pyspark` / `spark-scala` / `databricks-*`: `DecimalType(precision, scale)`
- `scala`: `BigDecimal`
- `avro`: `{"type": "bytes", "logicalType": "decimal", "precision": …, "scale": …}`
- `pydantic`: `Field(multiple_of=…, ge=…, le=…)`

Targets without a native decimal type (`python`, `gotypes`, `protobuf`) keep `float64` / `double`.

### String constraints

| Constraint     | Targets that consume it                                                            |
|----------------|------------------------------------------------------------------------------------|
| `maxLength`    | `spark-sql` & `databricks-sql` → `VARCHAR(n)`; `pydantic` → `Field(max_length=n)`  |
| `minLength`    | `databricks-sql` → `CHECK (length(col) >= n)`; `pydantic` → `Field(min_length=n)`  |
| `pattern`      | `databricks-sql` → `CHECK (col RLIKE …)`; `pydantic` → `Field(pattern=…)`; `dqx-yaml` → `regex_match` |
| `format`       | every target maps `date`, `date-time`, `uuid` to its native type. `markdown` and `dqx-yaml` understand the full format vocabulary (email, ipv4, ipv6, etc.) |

### Enum & const

`enum` and `const` get the richest treatment:

- `pydantic` / `python`: `Literal["A", "B", "C"]` typing (mypy/pyright/pydantic all enforce it).
- `avro`: a native `{"type": "enum", "name": …, "symbols": […]}` record.
- `protobuf`: a top-level `enum Name { NAME_UNSPECIFIED = 0; … }` declaration.
- `databricks-sql`: `CONSTRAINT col_enum CHECK (col IN (…))` / `CONSTRAINT col_const CHECK (col = …)`.
- `dqx-yaml`: `is_in_list` / `is_equal_to`.
- `markdown`: rendered inline in the field description.

### Annotations

| Annotation     | Targets that consume it                                                            |
|----------------|------------------------------------------------------------------------------------|
| `description`  | `databricks-pyspark` (metadata comment), `databricks-scala` (`.withComment`), `databricks-sql` (`COMMENT '…'`), `pydantic` (`Field(description=…)`), `markdown` (rendered) |
| `required`     | every target — required fields become non-nullable / non-optional                  |

### Maps & arrays

- `additionalProperties` without `properties` becomes a map type in every target with a native map (`MAP<K, V>`, `dict[K, V]`, `map[K]V`, `MapType`, `map<K, V>`).
- `items` becomes the target's native array (`ARRAY<T>`, `list[T]`, `[]T`, `ArrayType`, `repeated`).
- `minItems` / `maxItems` are enforced by `pydantic` (`Field(min_length=, max_length=)`) and `dqx-yaml` (`size()` checks).

## Examples

### PySpark — integer narrowing + decimal

```yaml
type: object
properties:
  order_id:    { type: string }
  customer_id: { type: string }
  age:         { type: integer, minimum: 0, maximum: 150 }
  price:       { type: number, multipleOf: 0.01, maximum: 99999.99 }
required: [order_id, customer_id]
```

```python
import pyspark.sql.types as T

orders_schema = T.StructType([
    T.StructField("order_id",    T.StringType(),      nullable=False),
    T.StructField("customer_id", T.StringType(),      nullable=False),
    T.StructField("age",         T.ShortType(),       nullable=True),
    T.StructField("price",       T.DecimalType(7, 2), nullable=True),
])
```

### Pydantic — runtime validation

```yaml
type: object
required: [id, status, email]
properties:
  id:     { type: string, format: uuid }
  status: { type: string, enum: [ACTIVE, INACTIVE, PENDING] }
  email:  { type: string, minLength: 3, maxLength: 254, pattern: "^.+@.+$" }
```

```python
from __future__ import annotations
from typing import Literal
from pydantic import BaseModel, Field

class UsersSchema(BaseModel):
    id: str
    status: Literal["ACTIVE", "INACTIVE", "PENDING"]
    email: str = Field(min_length=3, max_length=254, pattern="^.+@.+$")
```

### Databricks SQL — Delta CHECK constraints

```yaml
type: object
required: [id, status, schema_version, email]
properties:
  id:             { type: string, format: uuid }
  status:         { type: string, enum: [ACTIVE, INACTIVE, PENDING] }
  schema_version: { type: string, const: v1 }
  email:          { type: string, minLength: 3, maxLength: 254, pattern: "^.+@.+$" }
```

```sql
CREATE TABLE orders_schema (
    id             STRING NOT NULL,
    status         STRING NOT NULL,
    schema_version STRING NOT NULL,
    email          VARCHAR(254) NOT NULL,
    CONSTRAINT status_enum         CHECK (`status` IN ('ACTIVE', 'INACTIVE', 'PENDING')),
    CONSTRAINT schema_version_const CHECK (`schema_version` = 'v1'),
    CONSTRAINT email_pattern       CHECK (`email` RLIKE '^.+@.+$'),
    CONSTRAINT email_minlen        CHECK (length(`email`) >= 3)
)
USING DELTA;
```

### Protobuf — top-level enum

```yaml
type: object
properties:
  name:   { type: string }
  status: { type: string, enum: [ACTIVE, INACTIVE, PENDING] }
```

```protobuf
syntax = "proto3";
package proto;

enum Status {
  STATUS_UNSPECIFIED = 0;
  ACTIVE = 1;
  INACTIVE = 2;
  PENDING = 3;
}

message UsersSchema {
  string name = 1;
  Status status = 2;
}
```

### Avro — native enum + decimal logical type

```yaml
type: object
required: [status, price]
properties:
  status: { type: string, enum: [ACTIVE, INACTIVE, PENDING] }
  price:  { type: number, multipleOf: 0.01, maximum: 99999.99 }
```

```json
{
  "type": "record",
  "name": "UsersSchema",
  "namespace": "schemas",
  "fields": [
    { "name": "status", "type": { "type": "enum", "name": "Status", "symbols": ["ACTIVE", "INACTIVE", "PENDING"] } },
    { "name": "price",  "type": { "type": "bytes", "logicalType": "decimal", "precision": 7, "scale": 2 } }
  ]
}
```

### DQX — quality checks

```yaml
type: object
required: [id, status]
properties:
  id:     { type: string, format: uuid }
  status: { type: string, enum: [active, inactive] }
  age:    { type: integer, minimum: 0, maximum: 150 }
```

```yaml
- criticality: error
  check: { function: is_not_null,    arguments: { column: id } }
- criticality: error
  check: { function: regex_match,    arguments: { column: id, regex: "^[0-9a-fA-F]{8}-…$" } }
- criticality: error
  check: { function: is_not_null,    arguments: { column: status } }
- criticality: error
  check: { function: is_in_list,     arguments: { column: status, allowed: [active, inactive] } }
- criticality: error
  check: { function: sql_expression, arguments: { expression: "`age` IS NULL OR (`age` >= 0 AND `age` <= 150)", msg: "..." } }
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

The translator maps JSON Schema types to the appropriate types in each target language. For formats not listed here, the mapping follows the same pattern: the most specific available type for the declared JSON Schema type and format. See each translator's README (`internal/translate/<format>/README.md`) for the full feature matrix.

| JSON Schema | PySpark | Pydantic | Go | Protobuf | Avro |
|---|---|---|---|---|---|
| `string` | `StringType` | `str` | `string` | `string` | `string` |
| `string` + `format: date` | `DateType` | `datetime.date` | `time.Time` | `string` | `int` (logical `date`) |
| `string` + `format: date-time` | `TimestampType` | `datetime.datetime` | `time.Time` | `string` | `long` (logical `timestamp-millis`) |
| `string` + `enum: [...]` | `StringType` | `Literal[...]` | `string` | top-level `enum` | `enum` |
| `string` + `const: v` | `StringType` | `Literal[v]` | `string` | `string` | `string` |
| `string` + `maxLength: n` | `StringType` | `Field(max_length=n)` | `string` | `string` | `string` |
| `integer` (no bounds) | `LongType` | `int` | `int64` | `int64` | `long` |
| `integer` (byte range) | `ByteType` | `int` | `int8` / `uint8` | `int32` / `uint32` | `int` |
| `integer` (int32 range) | `IntegerType` | `int` | `int32` / `uint32` | `int32` / `uint32` | `int` |
| `number` | `DoubleType` | `float` | `float64` | `double` | `double` |
| `number` + `multipleOf: 0.01` | `DecimalType(p, 2)` | `Field(multiple_of=0.01)` | `float64` | `double` | logical `decimal` |
| `boolean` | `BooleanType` | `bool` | `bool` | `bool` | `boolean` |
| `array` | `ArrayType` | `list[...]` | `[]T` | `repeated` | `array` |
| `object` (with `properties`) | `StructType` | nested model | nested struct | nested message | nested `record` |
| `object` (`additionalProperties` only) | `MapType` | `dict[K, V]` | `map[K]V` | `map<K, V>` | `map` |

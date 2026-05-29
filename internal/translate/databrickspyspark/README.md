# Databricks PySpark

Translates JSON Schema to Databricks PySpark `StructType` definitions with column comments (`.py`).

Identical to the [`pyspark`](../pyspark/README.md) translator plus column-level metadata: every `description` becomes a `metadata={"comment": ...}` entry that Databricks surfaces in catalogs and notebooks.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "properties": {
    "name":   { "type": "string",  "description": "Full name" },
    "age":    { "type": "integer", "minimum": 0,    "maximum": 150 },
    "status": { "type": "integer", "minimum": -128, "maximum": 127 },
    "price":  { "type": "number",  "multipleOf": 0.01, "minimum": 0, "maximum": 99999.99, "description": "Order total in EUR" }
  }
}
```

**Output** (Databricks PySpark):

```python
import pyspark.sql.types as T

orders_schema = T.StructType([
    T.StructField("name",   T.StringType(),      nullable=True, metadata={"comment": "Full name"}),
    T.StructField("age",    T.ShortType(),       nullable=True),
    T.StructField("status", T.ByteType(),        nullable=True),
    T.StructField("price",  T.DecimalType(7, 2), nullable=True, metadata={"comment": "Order total in EUR"}),
])
```

Integer narrowing, decimal narrowing, and `additionalProperties → MapType` behave exactly like [`pyspark`](../pyspark/README.md).

## Supported JSON Schema Features

### Type Keywords
- [x] type
- [ ] enum
- [ ] const

### Type Values
- [x] string
- [x] integer
- [x] number
- [x] boolean
- [x] array
- [x] object
- [ ] null

### Schema Composition
- [ ] allOf
- [ ] anyOf
- [ ] oneOf
- [ ] not

### Object Keywords
- [x] properties
- [x] required
- [x] additionalProperties
- [ ] patternProperties
- [ ] propertyNames
- [ ] minProperties / maxProperties
- [ ] unevaluatedProperties
- [ ] dependentRequired

### Array Keywords
- [x] items
- [ ] prefixItems
- [ ] contains
- [ ] minItems / maxItems
- [ ] uniqueItems
- [ ] unevaluatedItems
- [ ] maxContains / minContains

### Numeric Validation
- [x] minimum / maximum
- [ ] exclusiveMinimum / exclusiveMaximum
- [x] multipleOf

### String Validation
- [ ] minLength / maxLength
- [ ] pattern

### References & Definitions
- [x] $ref
- [x] $defs
- [ ] $id
- [ ] $anchor
- [ ] $dynamicRef / $dynamicAnchor

### String Formats
- [x] date
- [x] date-time
- [ ] time
- [ ] duration
- [x] uuid
- [ ] uri / uri-reference / uri-template
- [ ] iri / iri-reference
- [ ] email / idn-email
- [ ] hostname / idn-hostname
- [ ] ipv4 / ipv6
- [ ] json-pointer / relative-json-pointer
- [ ] regex

### Annotations
- [x] description
- [ ] title
- [ ] default
- [ ] deprecated
- [ ] readOnly / writeOnly
- [ ] examples

### Conditional
- [ ] if / then / else
- [ ] dependentSchemas

### Content
- [ ] contentEncoding
- [ ] contentMediaType
- [ ] contentSchema

# PySpark

Translates JSON Schema to PySpark `StructType` definitions (`.py`).

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "properties": {
    "name":     { "type": "string" },
    "age":      { "type": "integer", "minimum": 0,    "maximum": 150 },
    "status":   { "type": "integer", "minimum": -128, "maximum": 127 },
    "price":    { "type": "number",  "multipleOf": 0.01, "minimum": 0, "maximum": 99999.99 },
    "tags":     { "type": "array", "items": { "type": "string" } },
    "metadata": { "type": "object", "additionalProperties": { "type": "string" } }
  }
}
```

**Output** (PySpark):

```python
import pyspark.sql.types as T

orders_schema = T.StructType([
    T.StructField("name",     T.StringType(),                  nullable=True),
    T.StructField("age",      T.ShortType(),                   nullable=True),
    T.StructField("status",   T.ByteType(),                    nullable=True),
    T.StructField("price",    T.DecimalType(7, 2),             nullable=True),
    T.StructField("tags",     T.ArrayType(T.StringType()),     nullable=True),
    T.StructField("metadata", T.MapType(T.StringType(), T.StringType()), nullable=True),
])
```

`minimum`/`maximum` narrow integers to `ByteType` / `ShortType` / `IntegerType` / `LongType`. `multipleOf` paired with bounds promotes numbers to `DecimalType(precision, scale)`. `additionalProperties` without `properties` becomes `MapType`.

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
- [ ] description
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

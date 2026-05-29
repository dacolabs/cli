# Spark Scala

Translates JSON Schema to Spark Scala `StructType` definitions (`.scala`).

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "properties": {
    "name":  { "type": "string" },
    "age":   { "type": "integer", "minimum": 0,    "maximum": 150 },
    "short": { "type": "integer", "minimum": -32768, "maximum": 32767 },
    "price": { "type": "number",  "multipleOf": 0.01, "maximum": 99999.99 }
  }
}
```

**Output** (Spark Scala):

```scala
import org.apache.spark.sql.types._

object Users extends Serializable {
  lazy val users_schema: StructType = StructType(Array(
    StructField("name",  StringType,         nullable = true),
    StructField("age",   ByteType,           nullable = true),
    StructField("short", ShortType,          nullable = true),
    StructField("price", DecimalType(7, 2),  nullable = true)
  ))
}
```

Integer bounds narrow to `ByteType` / `ShortType` / `IntegerType` / `LongType`. `multipleOf` paired with bounds promotes numbers to `DecimalType(precision, scale)`.

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

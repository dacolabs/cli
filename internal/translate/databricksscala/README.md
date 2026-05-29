# Databricks Scala

Translates JSON Schema to Databricks Scala `StructType` definitions with column comments (`.scala`).

Identical to [`spark-scala`](../sparkscala/README.md) plus `withComment(...)` on every field carrying a `description`.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "properties": {
    "name":  { "type": "string",  "description": "Full name" },
    "age":   { "type": "integer", "minimum": 0,    "maximum": 150 },
    "price": { "type": "number",  "multipleOf": 0.01, "maximum": 99999.99, "description": "Order total" }
  }
}
```

**Output** (Databricks Scala):

```scala
import org.apache.spark.sql.types._

object Users extends Serializable {
  val users_schema = StructType(Seq(
    StructField("name",  StringType,        nullable = true).withComment("Full name"),
    StructField("age",   ByteType,          nullable = true),
    StructField("price", DecimalType(7, 2), nullable = true).withComment("Order total")
  ))
}
```

Integer narrowing and decimal narrowing behave the same as [`spark-scala`](../sparkscala/README.md).

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

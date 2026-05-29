# Scala

Translates JSON Schema to Scala case class definitions (`.scala`). Integer bounds narrow to `Byte` / `Short` / `Int` / `Long`; `multipleOf` on numbers promotes to `BigDecimal`.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "properties": {
    "name":  { "type": "string" },
    "byte":  { "type": "integer", "minimum": 0,    "maximum": 127 },
    "short": { "type": "integer", "minimum": -32768, "maximum": 32767 },
    "price": { "type": "number",  "multipleOf": 0.01 }
  }
}
```

**Output** (Scala):

```scala
case class UsersSchema(
  name:  Option[String]     = None,
  byte:  Option[Byte]       = None,
  short: Option[Short]      = None,
  price: Option[BigDecimal] = None
)
```

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

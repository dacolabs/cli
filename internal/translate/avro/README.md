# Avro

Translates JSON Schema to Apache Avro schema documents (`.avsc`). Integer bounds narrow `long → int`. `multipleOf` on a number promotes to the `decimal` logical type. String-only `enum` constraints become native Avro `enum` records.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "required": ["id", "status", "count", "price"],
  "properties": {
    "id":     { "type": "string", "format": "uuid" },
    "status": { "type": "string", "enum": ["ACTIVE", "INACTIVE", "PENDING"] },
    "count":  { "type": "integer", "minimum": 0, "maximum": 2147483647 },
    "price":  { "type": "number",  "multipleOf": 0.01, "maximum": 99999.99 }
  }
}
```

**Output** (Avro `.avsc`):

```json
{
  "type": "record",
  "name": "OrdersSchema",
  "namespace": "schemas",
  "fields": [
    { "name": "id",     "type": { "type": "string", "logicalType": "uuid" } },
    { "name": "status", "type": { "type": "enum", "name": "Status", "symbols": ["ACTIVE", "INACTIVE", "PENDING"] } },
    { "name": "count",  "type": "int" },
    { "name": "price",  "type": { "type": "bytes", "logicalType": "decimal", "precision": 7, "scale": 2 } }
  ]
}
```

Enum names default to the PascalCased field name. Enums with non-string values or symbols that violate Avro's identifier rules (`[A-Za-z_][A-Za-z0-9_]*`) fall back to `string`.

## Constraint handling

Follows the narrow + validate principle (see [../CONSTRAINTS.md](../CONSTRAINTS.md)). `long` narrows to `int` when bounds fit signed 32-bit; a fractional `multipleOf` narrows to the `decimal` logical type; string enums with valid symbols become a native Avro `enum`. The field description maps to Avro `doc`, and every remaining constraint is preserved losslessly as field custom attributes (Avro permits and ignores unknown field attributes).

## Supported JSON Schema Features

### Type Keywords
- [x] type
- [x] enum
- [ ] const

### Type Values
- [x] string
- [x] integer
- [x] number
- [x] boolean
- [x] array
- [x] object
- [x] null

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

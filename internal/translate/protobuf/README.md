# Protocol Buffers

Translates JSON Schema to Protocol Buffers (proto3) message definitions (`.proto`). Integer bounds narrow `int64 → int32`/`uint32`/`uint64`. String-only `enum` constraints become top-level proto enums.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "properties": {
    "name":   { "type": "string" },
    "status": { "type": "string",  "enum": ["ACTIVE", "INACTIVE", "PENDING"] },
    "count":  { "type": "integer", "minimum": 0, "maximum": 2147483647 },
    "id":     { "type": "integer", "minimum": -2147483648, "maximum": 2147483647 }
  }
}
```

**Output** (proto3):

```protobuf
syntax = "proto3";

package proto;

enum Status {
  STATUS_UNSPECIFIED = 0;
  ACTIVE = 1;
  INACTIVE = 2;
  PENDING = 3;
}

message OrdersSchema {
  string name = 1;
  Status status = 2;
  uint32 count = 3;
  int32 id = 4;
}
```

Each enum field becomes a top-level `enum` declaration (named after the field, PascalCased), with the conventional `*_UNSPECIFIED = 0;` zero value. Identical `(name, symbols)` pairs are deduplicated across fields. Enums with non-string values or symbols that violate proto identifier rules fall back to `string`. Protobuf has no native decimal type; `multipleOf` doesn't narrow `double`. Protobuf has no 8/16-bit scalars, so `Int8`/`Int16` bounds widen to `int32` (or `uint32` when non-negative).

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
- [ ] multipleOf

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
- [ ] date
- [ ] date-time
- [ ] time
- [ ] duration
- [ ] uuid
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

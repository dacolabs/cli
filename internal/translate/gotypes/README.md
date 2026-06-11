# Go Types

Translates JSON Schema to Go struct definitions (`.go`). Integer bounds narrow to the smallest fitting type; non-negative bounds promote to the matching unsigned type.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "properties": {
    "name":     { "type": "string" },
    "byte":     { "type": "integer", "minimum": 0,      "maximum": 127 },
    "short":    { "type": "integer", "minimum": -32768, "maximum": 32767 },
    "id":       { "type": "integer", "minimum": 0,      "maximum": 2147483647 },
    "metadata": { "type": "object", "additionalProperties": { "type": "string" } }
  }
}
```

**Output** (Go):

```go
package schemas

type UsersSchema struct {
    Name     *string           `json:"name,omitempty"`
    Byte     *uint8            `json:"byte,omitempty"`
    Short    *int16            `json:"short,omitempty"`
    ID       *uint32           `json:"id,omitempty"`
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

Integer narrowing: bounds in `[-128, 127]` → `int8`, `[-32768, 32767]` → `int16`, `[INT32_MIN, INT32_MAX]` → `int32`, otherwise `int64`. When `minimum >= 0`, the result is the matching unsigned type (`uint8` / `uint16` / `uint32` / `uint64`). Go has no native decimal type — `multipleOf` doesn't narrow `float64`.

## Constraint handling

Follows the narrow + validate principle (see [../CONSTRAINTS.md](../CONSTRAINTS.md)). `int64` narrows to sized signed/unsigned ints (`int8`…`uint64`) from bounds and sign. Remaining constraints are emitted as a [go-playground/validator](https://github.com/go-playground/validator) `validate:"..."` struct tag alongside the `json` tag: `minimum`/`maximum` and `minLength`/`maxLength` → `min`/`max`, exclusive bounds → `gt`/`lt`, `enum` → `oneof`, `const` → `eq`. `pattern` and `multipleOf` have no validator builtin and are omitted (the generated code assumes the validator library downstream).

## Supported JSON Schema Features

### Type Keywords
- [x] type
- [x] enum
- [x] const

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
- [x] exclusiveMinimum / exclusiveMaximum
- [ ] multipleOf

### String Validation
- [x] minLength / maxLength
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

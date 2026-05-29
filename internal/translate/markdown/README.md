# Markdown

Translates JSON Schema to human-readable Markdown documentation (`.md`). Every constraint appears as a parenthetical in the Description column, and constraint combinations that would narrow to a specific physical type also surface a `narrows to:` hint.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "description": "User information",
  "required": ["name", "status", "byte"],
  "properties": {
    "name":   { "type": "string",  "description": "Full name", "minLength": 1, "maxLength": 64 },
    "status": { "type": "string",  "enum": ["ACTIVE", "INACTIVE"] },
    "byte":   { "type": "integer", "minimum": 0, "maximum": 127 },
    "price":  { "type": "number",  "multipleOf": 0.01, "maximum": 99999.99 }
  }
}
```

**Output** (Markdown):

```markdown
# Users

User information

## Fields

| Field    | Type    | Required | Description |
|----------|---------|----------|-------------|
| `name`   | string  | Yes      | Full name (minLength: 1, maxLength: 64) |
| `status` | string  | Yes      | (enum: `ACTIVE`, `INACTIVE`) |
| `byte`   | integer | Yes      | (minimum: 0, maximum: 127, narrows to: int8) |
| `price`  | number  | No       | (maximum: 99999.99, multipleOf: 0.01, narrows to: decimal(7,2)) |
```

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
- [x] minItems / maxItems
- [ ] uniqueItems
- [ ] unevaluatedItems
- [ ] maxContains / minContains

### Numeric Validation
- [x] minimum / maximum
- [x] exclusiveMinimum / exclusiveMaximum
- [x] multipleOf

### String Validation
- [x] minLength / maxLength
- [x] pattern

### References & Definitions
- [x] $ref
- [x] $defs
- [ ] $id
- [ ] $anchor
- [ ] $dynamicRef / $dynamicAnchor

### String Formats
- [x] date
- [x] date-time
- [x] time
- [x] duration
- [x] uuid
- [x] uri / uri-reference / uri-template
- [x] iri / iri-reference
- [x] email / idn-email
- [x] hostname / idn-hostname
- [x] ipv4 / ipv6
- [x] json-pointer / relative-json-pointer
- [x] regex

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

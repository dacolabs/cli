# Pydantic

Translates JSON Schema to Pydantic v2 `BaseModel` definitions (`.py`). Constraints flow through as runtime-enforced `Field(...)` arguments and `Literal[...]` types — bad data raises at model construction.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "required": ["id", "status", "schema_version", "price", "email"],
  "properties": {
    "id":             { "type": "string", "format": "uuid" },
    "status":         { "type": "string", "enum": ["ACTIVE", "INACTIVE", "PENDING"] },
    "schema_version": { "type": "string", "const": "v1" },
    "price":          { "type": "number", "multipleOf": 0.01, "minimum": 0, "maximum": 99999.99 },
    "email":          { "type": "string", "minLength": 3, "maxLength": 254, "pattern": "^.+@.+$" },
    "tags":           { "type": "array",  "items": { "type": "string" }, "minItems": 1, "maxItems": 10 },
    "notes":          { "type": "string", "description": "Optional notes" }
  }
}
```

**Output** (Pydantic):

```python
from __future__ import annotations

from typing import Optional, Literal
from pydantic import BaseModel, Field

class OrdersSchema(BaseModel):
    id: str
    status: Literal["ACTIVE", "INACTIVE", "PENDING"]
    schema_version: Literal["v1"]
    price: float = Field(ge=0, le=99999.99, multiple_of=0.01)
    email: str = Field(min_length=3, max_length=254, pattern="^.+@.+$")
    tags: list[str] = Field(min_length=1, max_length=10)
    notes: Optional[str] = Field(default=None, description="Optional notes")
```

Constraint translation map:

| JSON Schema             | Pydantic                           |
|-------------------------|------------------------------------|
| `enum`                  | `Literal[a, b, c]`                 |
| `const`                 | `Literal[x]`                       |
| `minimum` / `maximum`   | `Field(ge=..., le=...)`            |
| `exclusiveMin/Max`      | `Field(gt=..., lt=...)`            |
| `multipleOf`            | `Field(multiple_of=...)`           |
| `minLength` / `maxLength` | `Field(min_length=, max_length=)` |
| `pattern`               | `Field(pattern=...)`               |
| `minItems` / `maxItems` | `Field(min_length=, max_length=)`  |
| `description`           | `Field(description=...)`           |

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

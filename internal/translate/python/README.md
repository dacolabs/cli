# Python

Translates JSON Schema to plain Python `dataclasses` (`.py`). Plain dataclasses don't validate at runtime — for runtime checks, use [`pydantic`](../pydantic/README.md). The dataclass output still benefits from `Literal[...]` typing for `enum` and `const`, which static type checkers (`mypy`, `pyright`) enforce.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "required": ["id", "status"],
  "properties": {
    "id":             { "type": "string", "format": "uuid" },
    "status":         { "type": "string", "enum": ["ACTIVE", "INACTIVE", "PENDING"] },
    "schema_version": { "type": "string", "const": "v1" },
    "tags":           { "type": "array",  "items": { "type": "string" } }
  }
}
```

**Output** (Python):

```python
from __future__ import annotations

import dataclasses
from typing import Optional, Literal

@dataclasses.dataclass
class OrdersSchema:
    id: str
    status: Literal["ACTIVE", "INACTIVE", "PENDING"]
    schema_version: Optional[Literal["v1"]] = None
    tags: Optional[list[str]] = None
```

Numeric ranges, lengths, patterns, and other constraints aren't enforceable in plain dataclass output; switch to `pydantic` if you need them.

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
- [ ] minimum / maximum
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

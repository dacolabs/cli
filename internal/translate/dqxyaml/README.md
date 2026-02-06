# DQX YAML

Translates JSON Schema constraints to [DQX YAML quality checks](https://databrickslabs.github.io/dqx/docs/guide/quality_checks_definition/#yaml-format-declarative-approach) (.yaml). The translator requires `databricks-labs-dqx==0.8.0`.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "required": ["id", "status"],
  "properties": {
    "id": { "type": "string", "format": "uuid" },
    "status": { "type": "string", "enum": ["active", "inactive"] },
    "age": { "type": "integer", "minimum": 0, "maximum": 150 }
  }
}
```

**Output** (DQX YAML):

```yaml
- criticality: error
  check:
    function: is_not_null
    arguments:
      column: id
- criticality: error
  check:
    function: regex_match
    arguments:
      column: id
      regex: ^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$
- criticality: error
  check:
    function: is_not_null
    arguments:
      column: status
- criticality: error
  check:
    function: is_in_list
    arguments:
      allowed:
        - active
        - inactive
      column: status
- criticality: error
  check:
    function: is_in_range
    arguments:
      column: age
      max_limit: 150
      min_limit: 0
```

## Supported JSON Schema Features

### Type Keywords
- [ ] type
- [x] enum
- [x] const

### Type Values
- [ ] string
- [ ] integer
- [ ] number
- [ ] boolean
- [ ] array
- [ ] object
- [ ] null

### Schema Composition
- [ ] allOf
- [ ] anyOf
- [ ] oneOf
- [ ] not

### Object Keywords
- [x] properties
- [x] required
- [ ] additionalProperties
- [ ] patternProperties
- [ ] propertyNames
- [ ] minProperties / maxProperties
- [ ] unevaluatedProperties
- [ ] dependentRequired

### Array Keywords
- [ ] items
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
- [x] email / idn-email
- [ ] hostname / idn-hostname
- [x] ipv4 / ipv6
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

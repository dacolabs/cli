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
    function: sql_expression
    arguments:
      expression: "`age` IS NULL OR (`age` >= 0 AND `age` <= 150)"
      msg: age must be between 0 and 150
```

DQX is the only translator whose output is _purely_ validation. Where other targets translate JSON Schema's type system to a target type system, DQX translates JSON Schema's _constraints_ to PySpark-evaluated row-level checks. Use it alongside any of the other formats to catch bad rows at load time.

Constraint translation map:

| JSON Schema                  | DQX check                    |
|------------------------------|------------------------------|
| `required`                   | `is_not_null`                |
| `enum`                       | `is_in_list`                 |
| `const`                      | `is_equal_to`                |
| `minimum` + `maximum`        | `is_in_range`                |
| `minimum` only               | `is_not_less_than`           |
| `maximum` only               | `is_not_greater_than`        |
| `exclusiveMinimum`/`Maximum` | `sql_expression` (`> / <`)   |
| `minLength` / `maxLength`    | `sql_expression` (`length()`)|
| `pattern`                    | `regex_match`                |
| `minItems` / `maxItems`      | `sql_expression` (`size()`)  |
| `format: date`               | `is_valid_date`              |
| `format: date-time`          | `is_valid_timestamp`         |
| `format: ipv4` / `ipv6`      | `is_valid_ipv4_address` / `is_valid_ipv6_address` |
| `format: uuid` / `email`     | `regex_match`                |

For nullable fields, every check is wrapped as `col IS NULL OR (…)` so optional columns don't fail when absent.

## Constraint handling

dqx-yaml is a pure validation target: it emits a DQX data-quality check for every constraint — `enum` (is_in_list), `const` (is_equal_to), `minimum`/`maximum` (is_in_range / bounds), exclusive bounds and `multipleOf` (sql_expression), `minLength`/`maxLength` and `minItems`/`maxItems` (sql_expression), `pattern` (regex_match), and the string formats. See [../CONSTRAINTS.md](../CONSTRAINTS.md).

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

# Spark SQL

Translates JSON Schema to Spark DDL `CREATE TABLE` statements (`.sql`). Integer bounds narrow to `TINYINT` / `SMALLINT` / `INT` / `BIGINT`; `multipleOf` promotes numbers to `DECIMAL(p, s)`; `maxLength` promotes strings to `VARCHAR(n)`.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "required": ["id", "code", "price"],
  "properties": {
    "id":     { "type": "string",  "format": "uuid" },
    "code":   { "type": "string",  "maxLength": 32 },
    "byte":   { "type": "integer", "minimum": 0,    "maximum": 127 },
    "short":  { "type": "integer", "minimum": -32768, "maximum": 32767 },
    "price":  { "type": "number",  "multipleOf": 0.01, "maximum": 99999.99 }
  }
}
```

**Output** (Spark SQL):

```sql
CREATE TABLE users_schema (
    id    STRING NOT NULL,
    code  VARCHAR(32) NOT NULL,
    byte  TINYINT,
    short SMALLINT,
    price DECIMAL(7, 2) NOT NULL
);
```

For Delta-Lake CHECK constraints on `enum`/`const`/`pattern`/`minLength`, use [`databricks-sql`](../databrickssql/README.md).

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

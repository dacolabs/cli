# Databricks SQL

Translates JSON Schema to Databricks SQL `CREATE TABLE … USING DELTA` statements (`.sql`). Adds everything `spark-sql` does plus column-level `COMMENT` and table-level Delta `CONSTRAINT … CHECK (…)` clauses derived from `enum`, `const`, `pattern`, and `minLength`.

## Example

**Input** (JSON Schema):

```json
{
  "type": "object",
  "required": ["id", "status", "schema_version", "code", "price", "email"],
  "properties": {
    "id":             { "type": "string", "format": "uuid",       "description": "Order ID" },
    "status":         { "type": "string", "enum": ["ACTIVE","INACTIVE","PENDING"] },
    "schema_version": { "type": "string", "const": "v1" },
    "code":           { "type": "string", "maxLength": 32 },
    "email":          { "type": "string", "minLength": 3, "maxLength": 254, "pattern": "^.+@.+$" },
    "price":          { "type": "number", "multipleOf": 0.01, "maximum": 99999.99 }
  }
}
```

**Output** (Databricks SQL):

```sql
CREATE TABLE orders_schema (
    id             STRING NOT NULL COMMENT 'Order ID',
    status         STRING NOT NULL,
    schema_version STRING NOT NULL,
    code           VARCHAR(32) NOT NULL,
    email          VARCHAR(254) NOT NULL,
    price          DECIMAL(7, 2) NOT NULL,
    CONSTRAINT status_enum            CHECK (`status` IN ('ACTIVE', 'INACTIVE', 'PENDING')),
    CONSTRAINT schema_version_const   CHECK (`schema_version` = 'v1'),
    CONSTRAINT email_pattern          CHECK (`email` RLIKE '^.+@.+$'),
    CONSTRAINT email_minlen           CHECK (length(`email`) >= 3)
)
USING DELTA;
```

CHECK clauses are appended after the column list. Delta enforces them on insert/update.

## Constraint handling

Follows the narrow + validate principle (see [../CONSTRAINTS.md](../CONSTRAINTS.md)). `BIGINT` narrows to `TINYINT`/`SMALLINT`/`INT` from bounds, `DOUBLE` to `DECIMAL(p,s)` from fractional `multipleOf`, `STRING` to `VARCHAR(n)` from `maxLength`. Because narrowing only approximates a bound, the exact constraints are additionally enforced via Delta `CHECK` constraints: `enum`/`const` (IN / =), `pattern` (RLIKE), `minLength` (length), `minimum`/`maximum` (BETWEEN), `exclusiveMinimum`/`exclusiveMaximum` (>, <), and `multipleOf` (% = 0). Descriptions become column `COMMENT`s.

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

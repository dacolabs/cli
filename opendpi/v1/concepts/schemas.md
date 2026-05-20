# Schemas

Schemas define the shape of your data using JSON Schema. Every port must have a schema that describes what data consumers can expect.

## JSON Schema

OpenDPI uses [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12/json-schema-core.html) for schema definitions. JSON Schema is a widely-adopted standard with extensive tooling support.

## Basic Types

### Primitive Types

```yaml
# String
type: string

# Number (includes decimals)
type: number

# Integer (whole numbers only)
type: integer

# Boolean
type: boolean

# Null
type: null
```

### Objects

```yaml
type: object
properties:
  name:
    type: string
  age:
    type: integer
required:
  - name
```

### Arrays

```yaml
type: array
items:
  type: string

# Array of objects
type: array
items:
  type: object
  properties:
    id: { type: integer }
    name: { type: string }
```

## Common Patterns

### Required Fields

Mark fields as required using the `required` keyword:

```yaml
type: object
required:
  - id
  - email
properties:
  id:
    type: integer
  email:
    type: string
  name:
    type: string  # Optional
```

### Formats

Use `format` for semantic validation:

```yaml
properties:
  email:
    type: string
    format: email
  created_at:
    type: string
    format: date-time
  website:
    type: string
    format: uri
  uuid:
    type: string
    format: uuid
  date:
    type: string
    format: date
```

Common formats:
- `date-time` - RFC 3339 timestamp
- `date` - Full date (YYYY-MM-DD)
- `time` - Time (HH:MM:SS)
- `email` - Email address
- `uri` - URI
- `uuid` - UUID

### Enums

Restrict values to a specific set:

```yaml
status:
  type: string
  enum:
    - pending
    - active
    - completed
    - cancelled
```

### Nullable Fields

Allow null values:

```yaml
# Option 1: Type array
deleted_at:
  type:
    - string
    - "null"
  format: date-time

# Option 2: Using oneOf
deleted_at:
  oneOf:
    - type: string
      format: date-time
    - type: "null"
```

### Additional Properties

Control extra properties on objects:

```yaml
# Only allow defined properties
type: object
properties:
  id: { type: integer }
additionalProperties: false

# Allow any extra properties
type: object
properties:
  id: { type: integer }
additionalProperties: true

# Extra properties must be strings
type: object
properties:
  id: { type: integer }
additionalProperties:
  type: string
```

## Inline vs Referenced Schemas

### Inline Schema

Define the schema directly in the port:

```yaml
ports:
  users:
    connections:
      - connection: "#/connections/db"
        location: users
    schema:
      type: object
      properties:
        id: { type: integer }
        email: { type: string, format: email }
```

**Use when:**
- Schema is simple
- Schema is unique to this port
- Document is small

### Referenced Schema

Define in `components/schemas` and reference with `$ref`:

```yaml
ports:
  users:
    connections:
      - connection: "#/connections/db"
        location: users
    schema:
      $ref: "#/components/schemas/User"

components:
  schemas:
    User:
      type: object
      properties:
        id: { type: integer }
        email: { type: string, format: email }
```

**Use when:**
- Schema is complex
- Schema is shared across ports
- You want better organization
- Schema has nested objects that could be reused

## Reusable Components

### Shared Schemas

Define once, use everywhere:

```yaml
ports:
  user_created:
    schema:
      $ref: "#/components/schemas/UserEvent"

  user_updated:
    schema:
      $ref: "#/components/schemas/UserEvent"

  user_deleted:
    schema:
      $ref: "#/components/schemas/UserEvent"

components:
  schemas:
    UserEvent:
      type: object
      properties:
        user_id: { type: string }
        timestamp: { type: string, format: date-time }
        event_type: { type: string }
```

### Composing Schemas

Reference schemas within other schemas:

```yaml
components:
  schemas:
    Address:
      type: object
      properties:
        street: { type: string }
        city: { type: string }
        country: { type: string }
        postal_code: { type: string }

    User:
      type: object
      properties:
        id: { type: integer }
        name: { type: string }
        billing_address:
          $ref: "#/components/schemas/Address"
        shipping_address:
          $ref: "#/components/schemas/Address"
```

### allOf for Extension

Combine schemas using `allOf`:

```yaml
components:
  schemas:
    BaseEvent:
      type: object
      properties:
        event_id: { type: string, format: uuid }
        timestamp: { type: string, format: date-time }
        source: { type: string }

    UserEvent:
      allOf:
        - $ref: "#/components/schemas/BaseEvent"
        - type: object
          properties:
            user_id: { type: string }
            action: { type: string }
```

## Best Practices

### Be Specific

```yaml
# Good - specific types and constraints
properties:
  age:
    type: integer
    minimum: 0
    maximum: 150
  email:
    type: string
    format: email
  status:
    type: string
    enum: [active, inactive]

# Avoid - too generic
properties:
  age: {}
  email: { type: string }
  status: { type: string }
```

### Document with Descriptions

```yaml
properties:
  retention_days:
    type: integer
    description: Number of days to retain data before archival
    minimum: 1
    maximum: 365
```

### Use Consistent Naming

```yaml
# Consistent: snake_case
properties:
  user_id: { type: string }
  created_at: { type: string, format: date-time }
  order_total: { type: number }

# Avoid mixing styles
properties:
  userId: { type: string }
  created_at: { type: string }
  OrderTotal: { type: number }
```

## Complete Example

```yaml
components:
  schemas:
    Order:
      type: object
      description: A customer order with line items
      required:
        - order_id
        - customer_id
        - items
        - created_at
      properties:
        order_id:
          type: string
          format: uuid
          description: Unique order identifier
        customer_id:
          type: string
          description: Reference to customer
        status:
          type: string
          enum:
            - pending
            - confirmed
            - shipped
            - delivered
            - cancelled
          default: pending
        items:
          type: array
          minItems: 1
          items:
            $ref: "#/components/schemas/OrderItem"
        total_amount:
          type: number
          minimum: 0
          description: Total order amount in cents
        currency:
          type: string
          pattern: "^[A-Z]{3}$"
          default: USD
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    OrderItem:
      type: object
      required:
        - product_id
        - quantity
        - unit_price
      properties:
        product_id:
          type: string
        quantity:
          type: integer
          minimum: 1
        unit_price:
          type: number
          minimum: 0
```

## Further Reading

- [JSON Schema Specification](https://json-schema.org/specification.html)
- [Understanding JSON Schema](https://json-schema.org/understanding-json-schema/)
- [JSON Schema Validators](https://json-schema.org/implementations.html#validators)

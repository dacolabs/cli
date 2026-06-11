# Constraint handling: the narrow + validate principle

Every translator handles JSON Schema validation constraints (`minimum`, `maximum`,
`multipleOf`, `enum`, `pattern`, `minLength`, …) with one shared principle:

> **Narrow** the physical type as tightly as the target allows, **and** preserve every
> remaining constraint through that target's best validation or metadata channel —
> losslessly.

Narrowing is an optimisation; the constraint set is the source of truth and must never be
dropped. Because a narrowed type only *approximates* a constraint (a Spark `TINYINT` is
wider than `0..150`, a Go `int8` doesn't encode `multipleOf`), the preserve/validate pass
is what makes output lossless. It must not assume the narrowed type already "covers" a
constraint.

## Two passes per field

1. **Narrow** — in `EnrichField`, pick the tightest target type the constraints justify,
   using the shared helpers so the policy stays identical across targets:
   - `NarrowInteger` / `NarrowNumber` ([numeric.go](numeric.go)) — integer width and
     fixed-point decimal shape.
   - `MaxStringLength` — `VARCHAR(n)` / sized strings.
   - `IsNonNegative` — unsigned integer selection.
   - `EnumStringSymbols` ([enum.go](enum.go)) — native enum promotion.
2. **Preserve / validate** — render *all* present constraints through the target's
   channel using the shared renderers in [constraints.go](constraints.go):
   - `ActiveConstraints(c)` — the canonical, ordered `[]KV` every preserve pass iterates.
   - `ConstraintsJSON` / `ConstraintEntriesJSON` — JSON object form (Avro props, Spark
     Scala `Metadata.fromJson`).
   - `ConstraintsPyDict` / `ConstraintEntriesPy` — Python dict form (PySpark `metadata=`).
   - `ConstraintsText` — human-readable form (comment channels).

`EnrichField` is the only place that reads `f.Constraints` for narrowing. The preserve
pass may run in `EnrichField` (writing `f.Tag`) or in `translator.go` (extra template data
via `data.Extra`), matching each translator's existing structure.

## Per-target channel (the contract)

| Translator | Narrow | Validate / preserve channel |
|---|---|---|
| pyspark, spark-scala | type narrowing | `StructField(..., metadata=<constraints>)` |
| databricks-pyspark, databricks-scala | type narrowing | comment **and** constraints in one metadata object |
| scala | type narrowing | (case classes; constraints preserved via field types only) |
| databricks-sql | type narrowing | `CHECK` constraints (enforceable) + `COMMENT` |
| spark-sql | type narrowing | constraints as an inline `/* … */` comment (engine has no `CHECK`) |
| avro | type narrowing | field custom properties + `doc` |
| protobuf | type narrowing + formats→Timestamp | `[(buf.validate.field)…]` options + leading comment |
| pydantic | richer types (uuid/Decimal) | `Field(ge=, le=, pattern=, …)` |
| python | richer types | `__post_init__` assertions |
| gotypes | type narrowing | `validate:"…"` struct tag (go-playground) alongside `json` |
| dqx-yaml | n/a (pure checks) | DQX checks for every constraint, incl. `multipleOf` |
| markdown | narrowing hint | renders every constraint as documentation |

## Cross-target caveat

The same constraint legitimately produces different artefacts per target (a byte type
here, a `ge=0` rule there, a comment elsewhere). That divergence is intentional — each
target does its best — but it means a constraint a user writes as a *business rule* can
become a *storage decision* in some targets. Treat "constraints may change the physical
type and are preserved via the target's channel" as a documented contract, not a surprise.

## Known per-target limitations

- **protobuf**: `multipleOf` has no protovalidate rule — it survives only in the leading
  comment. `pattern`/`minLength`/`maxLength` map to `(buf.validate.field).string`.
- **gotypes**: `pattern` and `multipleOf` have no go-playground builtin and are omitted
  from the `validate` tag (other constraints map to `min`/`max`/`gt`/`lt`/`oneof`/`eq`).
- **non-enforcing targets** (Spark `StructType`, plain Spark SQL, Avro): constraints are
  preserved as metadata/comments — machine-readable but not engine-enforced.
- **exclusiveMinimum/Maximum** never drive narrowing (only inclusive bounds do); they are
  preserved/validated like any other constraint.

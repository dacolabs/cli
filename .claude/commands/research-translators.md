---
description: Research schema-binding ecosystems for translator gaps and propose GitHub issues
argument-hint: [optional focus area, e.g. "pyspark" or "iceberg types"]
---

# /research-translators

Survey upstream schema-binding ecosystems for changes that the translators under [internal/translate/](../../internal/translate/) don't yet cover, then propose GitHub issues for the highest-value gaps. Issue creation is gated on user approval — the command never opens issues unattended.

## Workflow

1. **Index current translators**
   - List packages under [internal/translate/](../../internal/translate/) and read each `README.md` (the user-facing capability matrix per [.claude/rules/translators.md](../../.claude/rules/translators.md)).
   - For each translator, capture: target ecosystem, supported JSON Schema features, and any explicit "not supported" notes.
   - Read [internal/translate/prepare.go](../../internal/translate/prepare.go) and [internal/translate/resolver.go](../../internal/translate/resolver.go) so gap analysis is framed in terms of `TypeResolver` methods (`PrimitiveType`, `ArrayType`, `MapType`, `RefType`, `FormatDefName`, `EnrichField`) and `SchemaData`/`Constraints`, not invented vocabulary.

2. **Research upstream changes** (use `WebSearch` then `WebFetch` for primary sources only — release notes, spec changelogs, official docs)
   - Existing translator targets: Avro spec, Protobuf, PySpark / Databricks type system, Pydantic, Scala/Spark SQL, DQX, Go `time`/`encoding/json` evolutions.
   - JSON Schema itself: latest draft activity (2020-12 → next), new keywords, format registrations.
   - Adjacent ecosystems worth a translator: Apache Iceberg type system, Arrow, Delta Lake schema, Polars.
   - Focus area from arguments: $ARGUMENTS
   - Skip secondary sources (blog roundups, tutorials). Cite the primary URL for every claim.

3. **Check for duplicates**
   - Run `gh issue list --state all --search "translator"` and scan titles/bodies before proposing anything new. Note duplicates in the report rather than re-filing.

4. **Gap analysis**
   - For each candidate gap, write: target translator (or "new translator: X"), the specific schema feature or upstream change, and the concrete `TypeResolver` change it implies (e.g. "extend `PrimitiveType` to map `format: duration` to `IntervalType`" or "add `EnrichField` integer narrowing for `exclusiveMaximum`").
   - Rank by: (a) appears in ≥2 independent primary sources or in the target's official release notes, (b) implementation fits the existing pipeline without changing `Prepare` or `SchemaData` shape, (c) no existing open issue. Drop anything that doesn't clear (a).

5. **Propose, don't execute**
   - Print a numbered list of proposed issues to the user. For each: title, 3-5 line body preview, source links, suggested `TypeResolver`/`EnrichField` mapping.
   - If zero gaps clear the bar, say so and stop. Do not manufacture issues to hit a quota.
   - Ask the user which (if any) to file. Wait for explicit approval per issue or "file all".

6. **File approved issues**
   - Only after approval, run `gh issue create` for each. Title format: `translators: support {feature} in {target}`. Body must include:
     - **Problem**: what JSON Schema input is currently mis-translated or rejected.
     - **Proposed change**: which translator package, which `TypeResolver` method or `EnrichField` step, what `Constraints` field (if any) drives it.
     - **References**: primary source URLs.
     - **Test cases**: concrete schema → expected output snippets, matching the test shape in [internal/translate/pyspark/translator_test.go](../../internal/translate/pyspark/translator_test.go).
   - Report the created issue numbers and URLs back to the user.

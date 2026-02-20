# Scenario Structure and Metadata Checklist

## Purpose

Define mandatory structure and metadata conformance checks for every `test-cases/TC-*.md` scenario file.

## Normative Sources

- `docs/test-case-template.md`
- `docs/test-case-specification.md`

## Compliance Rules

1. Each scenario must define exactly one startup script and exactly one startup command.
2. Each scenario must define exactly one `Functional Behavior Reference` metadata field.
3. `Functional Behavior Reference` must be a Markdown reference to one subsection under `docs/product-documentation.md#4-functional-behavior`.
4. Metadata section may include only `Case ID`, `Functional Behavior Reference`, `Startup Script`, and `Startup Command`.
5. Headings and order must match the template exactly:
   - `## 1. Metadata`
   - `## 2. Scenario`
   - `## 3. Preconditions`
   - `## 4. Test Steps`
   - `## 5. Assertions`
   - `## 6. Final Result`
   - `## 7. Cleanup`
6. Required table columns from the template must remain present in Metadata, Test Steps, and Assertions tables.
7. Assertions table must include `Functional Behavior Reference` column.
8. Every assertion row must use exactly one `Functional Behavior Reference`, and all assertion references must match scenario metadata reference.

## Deterministic Fail Triggers

Checklist result is `FAIL` when at least one of the following is true:

- a scenario has zero startup script bindings,
- a scenario has more than one startup script binding,
- a scenario has zero startup commands,
- a scenario has more than one startup commands,
- a scenario has zero `Functional Behavior Reference` metadata fields,
- a scenario has more than one `Functional Behavior Reference` metadata fields,
- `Functional Behavior Reference` is not a Markdown reference to one subsection under `docs/product-documentation.md#4-functional-behavior`,
- any required heading is missing,
- required headings are out of order,
- a required metadata field is missing,
- a required table column is missing,
- assertion rows use mixed Functional Behavior references,
- assertion rows use Functional Behavior reference different from metadata.

## Baseline Audit (`TC-001`)

| Check ID | Check | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- |
| C1 | Exactly one startup script is declared. | `PASS` | Metadata has one `Startup Script` row with `scripts/start-direct-launch.sh`. |
| C2 | Exactly one startup command is declared. | `PASS` | Metadata has one `Startup Command` row with `bash scripts/start-direct-launch.sh`. |
| C3 | Exactly one Functional Behavior reference is declared. | `FAIL` | Metadata has no `Functional Behavior Reference` row. |
| C4 | Metadata field set matches allowed set exactly. | `FAIL` | Metadata rows are missing `Functional Behavior Reference`. |
| C5 | Required headings exist in required order. | `PASS` | `TC-001` includes sections `1` through `7` in template order. |
| C6 | Required table columns are present. | `FAIL` | Assertions table is missing `Functional Behavior Reference` column. |
| C7 | Assertion Functional Behavior references match metadata reference. | `FAIL` | Cannot validate equality because scenario metadata reference is missing. |

## Baseline Checklist Result

- Result: `FAIL`
- Reason: Startup binding and section order are correct, but Functional Behavior metadata and assertion-reference ownership fields are missing.

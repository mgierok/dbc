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

## Full-Suite Audit (`TC-001` to `TC-008`)

| Check ID | Check | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- |
| C1 | Exactly one startup script is declared. | `PASS` | `8/8` scenarios contain exactly one metadata `Startup Script` row. |
| C2 | Exactly one startup command is declared. | `PASS` | `8/8` scenarios contain exactly one metadata `Startup Command` row. |
| C3 | Exactly one Functional Behavior reference is declared. | `PASS` | `8/8` scenarios contain exactly one metadata `Functional Behavior Reference` row. |
| C4 | Metadata field set matches allowed set exactly. | `PASS` | Each scenario metadata table contains `6` table rows (header + separator + 4 allowed data rows only). |
| C5 | Required headings exist in required order. | `PASS` | Required headings `## 1` through `## 7` are present in all active scenarios (`8/8` per heading). |
| C6 | Required table columns are present. | `PASS` | Assertions table header and test-step header are present in all active scenarios (`8/8`). |
| C7 | Assertion Functional Behavior references match metadata reference. | `PASS` | All scenarios pass one-area ownership/purity checks with no metadata/assertion reference mismatch. |

## Full-Suite Checklist Result

- Result: `PASS`
- Reason: Active scenario set satisfies startup binding, one-reference ownership, template structure, and assertion-reference equality rules.

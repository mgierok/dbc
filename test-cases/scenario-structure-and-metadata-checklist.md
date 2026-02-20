# Scenario Structure and Metadata Checklist

## Purpose

Define mandatory structure and metadata conformance checks for every `test-cases/TC-*.md` scenario file.

## Normative Sources

- `docs/test-case-template.md`
- `docs/test-case-specification.md`

## Compliance Rules

1. Each scenario must define exactly one startup script and exactly one startup command.
2. Metadata section may include only `Case ID`, `Startup Script`, and `Startup Command`.
3. Headings and order must match the template exactly:
   - `## 1. Metadata`
   - `## 2. Scenario`
   - `## 3. Preconditions`
   - `## 4. Test Steps`
   - `## 5. Assertions`
   - `## 6. Final Result`
   - `## 7. Cleanup`
4. Required table columns from the template must remain present in Metadata, Test Steps, and Assertions tables.

## Deterministic Fail Triggers

Checklist result is `FAIL` when at least one of the following is true:

- a scenario has zero startup script bindings,
- a scenario has more than one startup script binding,
- a scenario has zero startup commands,
- a scenario has more than one startup commands,
- any required heading is missing,
- required headings are out of order,
- a required metadata field is missing,
- a required table column is missing.

## Baseline Audit (`TC-001`)

| Check ID | Check | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- |
| C1 | Exactly one startup script is declared. | `PASS` | Metadata has one `Startup Script` row with `scripts/start-direct-launch.sh`. |
| C2 | Exactly one startup command is declared. | `PASS` | Metadata has one `Startup Command` row with `bash scripts/start-direct-launch.sh`. |
| C3 | Metadata field set matches allowed set exactly. | `PASS` | Metadata rows are only `Case ID`, `Startup Script`, `Startup Command`. |
| C4 | Required headings exist in required order. | `PASS` | `TC-001` includes sections `1` through `7` in template order. |
| C5 | Required table columns are present. | `PASS` | Metadata/Test Steps/Assertions tables match template column sets. |

## Baseline Checklist Result

- Result: `PASS`
- Reason: `TC-001` satisfies startup binding, metadata, heading order, and required table structure checks.

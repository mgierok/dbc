# Test Fixture and Test Case Standard

## Purpose

This document defines fixture assets and mandatory contracts for behavior-oriented manual regression cases.
It applies to startup and runtime behavior verification scenarios in `test-cases/`.

Every test case must be a separate Markdown file and must follow `docs/test-case-template.md`.

## Fixture Database

- Fixture source file: `scripts/test.db`
- Scope: local/manual startup and runtime behavior verification.
- Domain modeled by fixture: `customers`, `categories`, `products`, `orders`, `order_items`.

## Startup Scripts Catalog

Run all commands from repository root.

| Script | Run | Use When |
| --- | --- | --- |
| [`scripts/start-direct-launch.sh`](../scripts/start-direct-launch.sh) | `bash scripts/start-direct-launch.sh` | Scenario must start in runtime immediately (`-d`) with no selector step. |
| [`scripts/start-selector-from-config.sh`](../scripts/start-selector-from-config.sh) | `bash scripts/start-selector-from-config.sh` | Scenario must start from selector with a valid config entry. |
| [`scripts/start-without-database.sh`](../scripts/start-without-database.sh) | `bash scripts/start-without-database.sh` | Scenario must start in mandatory first-entry setup (no configured databases). |
| [`scripts/start-informational.sh`](../scripts/start-informational.sh) | `bash scripts/start-informational.sh <help\|version>` | Scenario must validate startup informational behavior for `--help` or `--version`. |

### Output and Cleanup Rules

- Every startup script prints `TMP_ROOT=...`.
- `scripts/start-without-database.sh` additionally prints `TMP_DB=...`.
- Mandatory cleanup after each execution:
  - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`

## Test Case File Contract (Mandatory)

### 1. File Placement and Naming

- Default location: `test-cases/`.
- Each case is one Markdown file (`.md`).
- Filename must start with `TC-<NNN>` and follow: `TC-<NNN>-<behavior>-<expected-result>.md`.
- Do not use generic names such as `test1.md`, `scenario.md`, `case.md`.

### 2. Required Startup Script Binding

- Each test case must reference exactly one startup script from the catalog above.
- Metadata must include both script path and exact run command.
- If two startup contexts are needed, split into separate test cases.

### 3. Functional Behavior Ownership Contract

- Each scenario must declare exactly one `Functional Behavior Reference` in metadata.
- The value must be a Markdown reference targeting one subsection under:
  - `docs/product-documentation.md#4-functional-behavior`
- Assertions must be area-pure:
  - every assertion row must include one `Functional Behavior Reference`,
  - every assertion reference must be identical to the scenario metadata reference.
- Product documentation is the source of truth for available Functional Behavior subsections.
- This specification and template must not define an independent local allowlist of areas.

### 4. Minimal Required Metadata

Only fields below are allowed in `## 1. Metadata`:

- `Case ID`
- `Functional Behavior Reference`
- `Startup Script`
- `Startup Command`

### 5. Required Scenario Contract

Every test case must define:

- subject under test (what behavior is being tested),
- expected result (single observable behavior contract),
- explicit step-to-assertion mapping.

Each step row must contain:

- one user action,
- one expected UI/system outcome,
- one linked assertion ID.

Assertions table must contain:

- assertion ID,
- functional behavior reference,
- pass criteria,
- result (`PASS`/`FAIL`),
- evidence.

### 6. Expand-First Evidence Contract

- Before creating a new `TC-*` file for scoped behavior additions, execution evidence must show that expansion/refactor of existing scenarios was evaluated first.
- Release-readiness evidence must classify each coverage addition as:
  - `Expanded Existing TC`, or
  - `New TC`.
- Every `New TC` classification must include explicit rationale for why expansion was not viable.

### 7. Suite Governance Artifact Contract

The following files are mandatory and must remain synchronized:

- `test-cases/suite-coverage-matrix.md`
- `test-cases/scenario-structure-and-metadata-checklist.md`
- `test-cases/deterministic-result-audit-checklist.md`
- `test-cases/full-suite-release-readiness-audit.md`

Coverage matrix contract:

- Must map Functional Behavior reference -> scenario IDs -> assertion IDs.
- For each Product Documentation Functional Behavior subsection in active coverage scope, both `Scenario IDs` and `Assertion IDs` must be non-empty.
- Missing scenario mapping or missing assertion mapping is an audit `FAIL`.

Cross-artifact mismatch contract:

- Any mismatch between template fields, specification fields, and checklist rules is an audit `FAIL`.
- Any startup command used by scenarios but missing from startup scripts catalog is an audit `FAIL`.

### 8. Deterministic Result Rule

- Allowed assertion and final-result values are only `PASS` or `FAIL`.
- Final `PASS` is valid only when all assertions are `PASS`.
- Any unmet precondition, blocked execution, or failed expectation must produce final `FAIL` with reason.
- No third state (`SKIPPED`, `UNKNOWN`, `PARTIAL`) is allowed.

### 9. Strict Structure Rule

- Section order/headings from `docs/test-case-template.md` are mandatory.
- Required fields/columns in template tables cannot be removed.
- Additional notes are allowed only in the dedicated `Notes` field.
- Full consistency between this document and the template is mandatory.

## Canonical Template

- Template file: `docs/test-case-template.md`
- All new test cases must be created by copying this file and filling placeholders.

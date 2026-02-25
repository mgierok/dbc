# Test Case Authoring and Modification Standard

## Purpose

This document defines mandatory contracts for creating and modifying behavior-oriented manual regression scenarios in `test-cases/`.

Execution and result-reporting rules are defined in `docs/test-case-execution-reporting-specification.md`.

Every test case must be a separate Markdown file and must follow `docs/test-case/template.md`.

## 1. Test Case File Contract (Mandatory)

### 1.1 File Placement and Naming

- Default location: `test-cases/`.
- Each case is one Markdown file (`.md`).
- Filename must start with `TC-<NNN>` and follow: `TC-<NNN>-<behavior>-<expected-result>.md`.
- Do not use generic names such as `test1.md`, `scenario.md`, `case.md`.

### 1.2 Required Startup Script Binding

- Each test case must reference exactly one startup script from the catalog in `docs/test-case-execution-reporting-specification.md`.
- Metadata must include both script path and exact run command.
- If two startup contexts are needed, split into separate test cases.

### 1.3 Functional Behavior Ownership Contract

- Each scenario must declare exactly one `Functional Behavior Reference` in metadata.
- The value must be a Markdown reference targeting one subsection under:
  - `docs/product-documentation.md#4-functional-behavior`
- Assertions must be area-pure:
  - every assertion row must include one `Functional Behavior Reference`,
  - every assertion reference must be identical to the scenario metadata reference.
- Product documentation is the source of truth for available Functional Behavior subsections.
- This specification and template must not define an independent local allowlist of areas.

### 1.4 Minimal Required Metadata

Only fields below are allowed in `## 1. Metadata`:

- `Case ID`
- `Functional Behavior Reference`
- `Startup Script`
- `Startup Command`

### 1.5 Required Scenario Contract

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

### 1.6 Expand-First Evidence Contract

- Before creating a new `TC-*` file for scoped behavior additions, execution evidence must show that expansion/refactor of existing scenarios was evaluated first.
- Release-readiness evidence must classify each coverage addition as:
  - `Expanded Existing TC`, or
  - `New TC`.
- Every `New TC` classification must include explicit rationale for why expansion was not viable.

### 1.7 Suite Governance Artifact Contract

The following files are mandatory and must remain synchronized:

- `test-cases/suite-coverage-matrix.md`

Coverage matrix contract:

- Must map Functional Behavior reference -> scenario IDs -> assertion IDs.
- `test-cases/suite-coverage-matrix.md` stores only factual mapping data (no local instructions, rules, or conclusions).
- For each Product Documentation Functional Behavior subsection in active coverage scope, both `Scenario IDs` and `Assertion IDs` must be non-empty.
- Every mapped reference must be a Markdown link to one subsection under `docs/product-documentation.md#4-functional-behavior`.
- Every listed scenario ID must resolve to an existing `test-cases/TC-*.md` file.
- Missing scenario mapping or missing assertion mapping is an audit `FAIL`.
- Invalid reference links or unresolved scenario IDs are an audit `FAIL`.

Cross-artifact mismatch contract:

- Any mismatch between template fields and this specification is an audit `FAIL`.
- Any startup command used by scenarios but missing from startup scripts catalog in `docs/test-case-execution-reporting-specification.md` is an audit `FAIL`.

Governance maintenance workflow:

- For every new or modified `TC-*` scenario, update `test-cases/suite-coverage-matrix.md`.
- For every new or modified `TC-*` scenario, execute structure/metadata and determinism governance checks.
- Governance evidence must stay concrete and include scenario IDs, assertion IDs, and exact file paths.
- Suite-level governance status is `PASS` only when coverage mapping and per-scenario governance checks are all `PASS`.
- Governance checks are required only when creating or modifying `TC-*` scenarios.
- Governance checks must not be persisted as `test-cases/*audit*.md` snapshots.

### 1.8 Strict Structure Rule

- Section order/headings from `docs/test-case/template.md` are mandatory.
- Required fields/columns in template tables cannot be removed.
- Additional notes are allowed only in the dedicated `Notes` field.
- Full consistency between this document and the template is mandatory.
- Structure/metadata conformance checks must verify:
  - exactly one startup script binding and exactly one startup command,
  - Section 1.3 `Functional Behavior Ownership Contract`.

Structure/metadata conformance `FAIL` triggers:

- a scenario has zero startup script bindings,
- a scenario has more than one startup script binding,
- a scenario has zero startup commands,
- a scenario has more than one startup commands,
- any required heading from `docs/test-case/template.md` is missing,
- required headings from `docs/test-case/template.md` are out of order,
- a required metadata field or required table column from `docs/test-case/template.md` is missing,
- any violation of Section 1.3 `Functional Behavior Ownership Contract`.

## 2. Canonical Templates

- Template file: `docs/test-case/template.md`
- All new test cases must be created by copying this file and filling placeholders.
- Suite coverage matrix template file: `docs/test-case/suite-coverage-matrix-template.md`
- `test-cases/suite-coverage-matrix.md` must follow `docs/test-case/suite-coverage-matrix-template.md`.

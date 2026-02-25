# Test Case Execution and Result Reporting Standard

## Purpose

This document defines fixture assets and mandatory contracts for executing behavior-oriented manual regression scenarios and reporting their results.
It applies to startup and runtime behavior verification scenarios in `test-cases/`.

Authoring and modification rules for test cases are defined in `docs/test-case-authoring-specification.md`.

## 1. Fixture Database

- Fixture source file: `scripts/test.db`
- Scope: local/manual startup and runtime behavior verification.
- Domain modeled by fixture: `customers`, `categories`, `products`, `orders`, `order_items`.

## 2. Startup Scripts Catalog

Run all commands from repository root.

| Script | Run | Use When |
| --- | --- | --- |
| [`scripts/start-direct-launch.sh`](../scripts/start-direct-launch.sh) | `bash scripts/start-direct-launch.sh` | Scenario must start in runtime immediately (`-d`) with no selector step. |
| [`scripts/start-selector-from-config.sh`](../scripts/start-selector-from-config.sh) | `bash scripts/start-selector-from-config.sh` | Scenario must start from selector with a valid config entry. |
| [`scripts/start-without-database.sh`](../scripts/start-without-database.sh) | `bash scripts/start-without-database.sh` | Scenario must start in mandatory first-entry setup (no configured databases). |
| [`scripts/start-informational.sh`](../scripts/start-informational.sh) | `bash scripts/start-informational.sh <help\|version>` | Scenario must validate startup informational behavior for `--help` or `--version`. |

### 2.1 Output and Cleanup Rules

- Every startup script prints `TMP_ROOT=...`.
- `scripts/start-without-database.sh` additionally prints `TMP_DB=...`.
- Mandatory cleanup after each execution:
  - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`

### 2.2 Subagent Isolation Contract

- Each `TC-*` scenario must be executed by a separate independent subagent instance.
- Subagent context must be isolated per scenario:
  - include only scenario-local execution inputs (the target `TC-*` file and its startup binding),
  - do not include cross-scenario notes, summaries, decisions, or outcomes.
- Full-suite execution must be orchestrated as one subagent per scenario, then aggregated only at final reporting level.

## 3. Execution Result Output Contract

- Result outputs must be displayed immediately after each single test-case execution and after each full-suite execution.
- Display output must include only one of:
  - `SINGLE` output for one executed test case, or
  - `SUITE` output for full-suite execution.
- Execution output must not include governance-check sections.
- Result outputs are display-only; do not create or maintain persistent release-readiness result files in `test-cases/`.
- Single-case and suite output format must follow `docs/test-case/execution-output-template.md`.
- Output values remain binary:
  - assertion/test/suite results: `PASS` or `FAIL` only.

## 4. Deterministic Result Rule

- Allowed assertion and final-result values are only `PASS` or `FAIL`.
- Final `PASS` is valid only when all assertions are `PASS`.
- Any unmet precondition, blocked execution, or failed expectation must produce final `FAIL` with reason.
- No third state (`SKIPPED`, `UNKNOWN`, `PARTIAL`) is allowed.
- `Functional Behavior Reference` cardinality/purity rules are defined in `docs/test-case-authoring-specification.md#13-functional-behavior-ownership-contract` and are normative for deterministic checks.

Violation Count contract:

- Each execution output (single or suite) must include explicit integer `Violation Count`.
- Any `Violation Count > 0` forces result `FAIL`.
- `Violation Count = 0` is required for result `PASS`.

Deterministic `FAIL` triggers:

- assertion result includes a value other than `PASS` or `FAIL`,
- final `Test Result` includes a value other than `PASS` or `FAIL`,
- final `PASS` is declared while at least one assertion is not `PASS`,
- final `FAIL` omits failure reason/context,
- ambiguous language prevents binary resolution,
- any violation of `docs/test-case-authoring-specification.md#13-functional-behavior-ownership-contract`,
- `Violation Count` is missing or not numeric.

## 5. Canonical Template

- Execution output template file: `docs/test-case/execution-output-template.md`

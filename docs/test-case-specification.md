# Test Fixture and Test Case Standard

## Purpose

This document defines the fixture assets and the mandatory format for behavior-oriented test cases.
It applies to tests covering:

- user journeys,
- user interaction with the TUI,
- critical runtime paths.

The target use of these cases is regression testing for startup and runtime behaviors.

Every test case must be a separate Markdown file and must follow the template in `docs/test-case-template.md`.

## Fixture Database

- Fixture source file: `scripts/test.db`.
- Scope: local/manual startup and runtime behavior verification.
- Domain modeled by fixture: `customers`, `categories`, `products`, `orders`, `order_items`.

### Table Summary

- `customers`
  - `id` (PK), `email` (UNIQUE, NOT NULL), `full_name` (NOT NULL, CHECK), `phone` (NULLable), `loyalty_points` (DEFAULT `0`), `is_active` (DEFAULT `1`, CHECK), `created_at` (DEFAULT timestamp).
- `categories`
  - `id` (PK), `name` (UNIQUE, NOT NULL), `description` (NOT NULL, DEFAULT `''`, CHECK length).
- `products`
  - `id` (PK), `sku` (UNIQUE, NOT NULL), `name` (NOT NULL), `category_id` (FK), `price_cents` (CHECK), `notes` (NULLable), `weight_kg` (`REAL`), `metadata` (`BLOB`), `is_discontinued` (DEFAULT `0`, CHECK).
- `orders`
  - `id` (PK), `customer_id` (FK), `status` (NOT NULL, enum-like CHECK), `placed_at` (NULLable), `note` (NOT NULL, DEFAULT `''`), `total_cents` (DEFAULT `0`, CHECK).
- `order_items`
  - `id` (PK), `order_id` (FK with `ON DELETE CASCADE`), `product_id` (FK), `quantity` (CHECK), `unit_price_cents` (CHECK), `discount_percent` (`REAL`), UNIQUE(`order_id`, `product_id`).

### Relationship Summary

- `products.category_id` -> `categories.id`
- `orders.customer_id` -> `customers.id`
- `order_items.order_id` -> `orders.id` (`ON DELETE CASCADE`)
- `order_items.product_id` -> `products.id`

## Startup Scripts Catalog

Run all commands from repository root.

| Script | Run | Use When |
| --- | --- | --- |
| [`scripts/start-direct-launch.sh`](../scripts/start-direct-launch.sh) | `bash scripts/start-direct-launch.sh` | Scenario must start in runtime immediately (`-d`) with no selector step. |
| [`scripts/start-selector-from-config.sh`](../scripts/start-selector-from-config.sh) | `bash scripts/start-selector-from-config.sh` | Scenario must start from selector with a valid config entry. |
| [`scripts/start-without-database.sh`](../scripts/start-without-database.sh) | `bash scripts/start-without-database.sh` | Scenario must start in mandatory first-entry setup (no configured databases). |

### Output and Cleanup Rules

- Every startup script prints `TMP_ROOT=...`.
- `scripts/start-without-database.sh` additionally prints `TMP_DB=...`.
- Mandatory cleanup after each test execution:
  - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`

## Test Case File Contract (Mandatory)

### 1. File Placement and Naming

- Default location: `test-cases/`.
- Each case is one file in Markdown format (`.md`).
- Filename must be descriptive and scenario-specific.
- Filename must start with `TC-<NNN>`.
- Use pattern: `TC-<NNN>-<behavior>-<expected-result>.md`.
- Hint to get next number quickly (without opening/scanning test content):
  ```bash
  LAST_NNN="$(rg --files test-cases 2>/dev/null | rg -o 'TC-[0-9]{3,}' | sed 's/TC-//' | sort -n | tail -1)"
  printf 'TC-%03d\n' "$((10#${LAST_NNN:-0}+1))"
  ```
- Do not use generic names such as `test1.md`, `scenario.md`, `case.md`.

### 2. Required Startup Script Binding

- Each test case must reference exactly one startup script from the catalog above.
- Test case must include:
  - script path,
  - exact run command.
- If two startup contexts are needed, split into two separate test cases.

### 3. Minimal Required Metadata

Only the fields below are allowed in `Metadata` section:

- `Case ID`
- `Startup Script`
- `Startup Command`

Do not add extra metadata fields (for example priority, owner, tags, dates) unless explicitly requested.

### 4. Required Scenario Contract

Every test case must define all elements below explicitly:

- subject under test (what behavior is being tested),
- expected result (single, observable behavior contract),
- pass/fail criteria mapped to explicit assertions.
- Each case must contribute to regression coverage of at least one area from this set:
  - user journey,
  - TUI behavior,
  - critical path.
- The full `test-cases/` suite must collectively cover all three areas above.
- A single test case may cover one, two, or all three areas.
- Prefer expanded scenarios with multiple context-relevant assertions instead of single-assertion scenarios split into separate files.
- Repeating an assertion across scenarios is allowed only when that assertion is necessary in the context of the scenario under test.
- Repeated assertions must not be used as a reason to create a new scenario.

Each step must contain:

- one user action,
- expected UI/system outcome,
- linked assertion ID.

### 5. Deterministic Result Rule

- Allowed final results are only `PASS` or `FAIL`.
- `PASS` is valid only when all assertions are marked `PASS`.
- Any unmet precondition, blocked execution, or failed expectation must be reported as `FAIL` with a reason.
- No third state (`SKIPPED`, `UNKNOWN`, `PARTIAL`) is allowed.

### 6. Strict Structure Rule

- The section order and headings from `docs/test-case-template.md` are mandatory.
- Required fields in template tables cannot be removed.
- Additional notes are allowed only in dedicated `Notes` fields.
- `docs/test-case-template.md` is an integral, normative part of this specification.
- Full consistency between this document and the template is mandatory.
- If any inconsistency is found, this specification is the parent document and has priority; the template must be updated to match it in the same change set.

## Canonical Template

- Template file: `docs/test-case-template.md`
- All new test cases must be created by copying this file and filling placeholders.

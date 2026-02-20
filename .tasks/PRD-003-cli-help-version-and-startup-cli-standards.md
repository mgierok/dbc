# Overview

This PRD defines a startup CLI standardization release for `dbc`, focused on discoverability (`--help`/`-h`), version introspection (`--version`/`-v`), and deterministic argument error behavior.

We believe that adding standards-aligned help/version commands and clear usage-error semantics for terminal-first DBC users will reduce startup friction and improve automation reliability.
We will know this is true when startup CLI compliance reaches target criteria during release validation.

# Metadata

- Status: DONE

# Problem Statement

Current startup CLI behavior supports direct database launch but does not provide a formal help command or version command. This creates avoidable friction for users who need self-service usage guidance and for scripts that require deterministic version output and exit-code behavior.

# Current State (As-Is)

- Startup supports direct-launch arguments (`-d` / `--database`).
- Unsupported or malformed startup arguments are rejected.
- Dedicated `--help`/`-h` behavior is not defined as a product contract.
- Dedicated `--version`/`-v` behavior is not defined as a product contract.
- Usage-error exit-code behavior is not explicitly standardized to CLI standards.
- CLI standards exist in `docs/cli-parameter-and-output-standards.md` but startup CLI surface is not fully aligned to that standard contract.

# Target State After Release (To-Be)

- Startup CLI exposes `--help` with short alias `-h`.
- Startup CLI exposes `--version` with short alias `-v`.
- `--help`/`-h` output is deterministic and includes short description, usage, option aliases, and practical examples.
- `--version`/`-v` prints a single-token value to stdout:
  - short commit hash when available,
  - `dev` when hash metadata is unavailable.
- Informational flags short-circuit startup runtime work:
  - `--help`/`-h` and `--version`/`-v` return success without database validation/open.
- Invalid usage and argument-validation failures return exit code `2`.
- Runtime/operational failures remain exit code `1`.
- Product and technical docs are updated with references to `docs/cli-parameter-and-output-standards.md` where relevant, without duplicating standard text.

# Business Rationale and Strategic Fit

- Improves first-run and occasional-use ergonomics by making CLI behavior self-discoverable.
- Increases scriptability and CI ergonomics with deterministic version and exit-code contracts.
- Reduces ambiguity in CLI behavior by aligning visible startup contract with documented standards.
- Preserves focused scope by standardizing the touched startup CLI surface first.

# Goals

- G1. Provide built-in CLI discoverability for startup usage.
- G2. Provide deterministic startup version introspection.
- G3. Standardize startup usage-error signaling for automation reliability.
- G4. Keep product and technical documentation aligned to current behavior via standards references.

# Non-Goals

- Standardizing all TUI command-mode behavior in this release.
- Adding new startup/global flags beyond help/version and existing direct-launch options.
- Introducing machine output formats (`json`/`yaml`) in this release.
- Expanding this release into broader repo-wide wording harmonization.

# Scope (In Scope / Out of Scope)

In Scope:
- Startup CLI contract for:
  - `--help` and `-h`,
  - `--version` and `-v`,
  - informational-flag short-circuit behavior,
  - usage-error exit code `2`,
  - runtime failure exit code `1`.
- Help content contract including alias visibility and practical examples.
- Version output contract (single token hash-or-dev).
- Documentation updates in product/technical docs and README where startup behavior changes, using references to CLI standards doc instead of content duplication.
- Test updates for all new/changed startup CLI behaviors.

Out of Scope:
- Broader TUI output/error-message standardization.
- Full adoption of all CLI standard sections not directly related to startup CLI entrypoint.
- New operational telemetry systems.
- Changes to database engine scope or save/navigation core workflows.

# Functional Requirements

FR-001: The product must support `--help` and `-h` as equivalent startup informational flags.
Acceptance: Running `dbc --help` and `dbc -h` produces equivalent deterministic help output and exits with code `0`.

FR-002: Help output must include startup description, canonical usage line, documented short/long aliases, and practical examples.
Acceptance: Help text includes usage and at least two practical examples, including direct launch and version invocation.

FR-003: The product must support `--version` and `-v` as equivalent startup informational flags.
Acceptance: Running `dbc --version` and `dbc -v` prints the same single-token version value and exits with code `0`.

FR-004: Version output token must be short commit hash when available, otherwise `dev`.
Acceptance: `--version`/`-v` prints hash-or-dev on stdout only, with no extra prose.

FR-005: Informational flags must short-circuit startup runtime work.
Acceptance: When `--help`/`-h` or `--version`/`-v` is present, startup does not attempt database open/validation and exits successfully.

FR-006: Invalid startup usage/argument validation must return exit code `2` with actionable guidance.
Acceptance: Unsupported, missing-value, and malformed startup argument cases terminate with exit code `2` and include corrective hint plus usage guidance.

FR-007: Runtime/operational startup failures must continue returning exit code `1`.
Acceptance: Non-usage failures in startup/runtime paths keep existing non-zero runtime-failure semantics with code `1`.

FR-008: Product and technical documentation must reference startup CLI standards where needed without duplicating standards content.
Acceptance: Updated docs include relevant references to `docs/cli-parameter-and-output-standards.md` and remain consistent with behavior.

# Non-Functional Product Requirements

NFR-001: Help/version behavior must be deterministic for the same build artifact.

NFR-002: Startup CLI error messages must remain concise and actionable for terminal use.

NFR-003: Version output must remain automation-friendly (single token on stdout).

NFR-004: Startup CLI changes must preserve existing direct-launch user value and not introduce discoverability regressions.

# Success Metrics and Release Criteria

Primary Outcome Metric:
- M1 Startup CLI standards compliance score for scoped checklist items.
Baseline: 4/9 relevant startup CLI items pass.
Target: 9/9 relevant startup CLI items pass.
Measurement window: PR validation to merge for this release.
Measurement method: Checklist-based audit against `docs/cli-parameter-and-output-standards.md` for startup CLI scope.

Leading Indicators:
- M2 Startup CLI contract test pass rate (new/updated tests for help/version/exit-codes/precedence).
Baseline: No dedicated help/version contract tests.
Target: 100% pass.
Measurement window: CI runs for release PR.
Measurement method: Automated Go test outcomes for touched startup CLI test scope.

- M3 Usage-error exit-code correctness in defined negative scenarios.
Baseline: Usage errors not standardized to code `2`.
Target: 100% of scoped usage-error scenarios return code `2`.
Measurement window: Release validation run.
Measurement method: Scenario assertions for exit status and message contract.

Guardrail Metric:
- M4 Direct-launch startup regression count.
Baseline: 0 known direct-launch regressions.
Target: 0 regressions post-change.
Measurement window: Release validation run.
Measurement method: Existing and updated startup tests for direct-launch behavior.

Release Criteria:
- FR-001 through FR-008 acceptance statements are satisfied.
- M1, M2, M3, and M4 meet target thresholds.
- Updated docs reference standards correctly and do not duplicate standards content.

# Final Acceptance Matrix

| Requirement | Status | Evidence |
| --- | --- | --- |
| FR-001 | PASS | `PRD-003-TASK-01`, `PRD-003-TASK-02`, `PRD-003-TASK-06` |
| FR-002 | PASS | `PRD-003-TASK-02`, `PRD-003-TASK-06` |
| FR-003 | PASS | `PRD-003-TASK-01`, `PRD-003-TASK-03`, `PRD-003-TASK-06` |
| FR-004 | PASS | `PRD-003-TASK-03`, `PRD-003-TASK-06` |
| FR-005 | PASS | `PRD-003-TASK-01`, `PRD-003-TASK-06` |
| FR-006 | PASS | `PRD-003-TASK-04`, `PRD-003-TASK-06` |
| FR-007 | PASS | `PRD-003-TASK-04`, `PRD-003-TASK-06` |
| FR-008 | PASS | `PRD-003-TASK-05`, `PRD-003-TASK-06` |
| NFR-001 | PASS | `PRD-003-TASK-02`, `PRD-003-TASK-03`, `PRD-003-TASK-06` |
| NFR-002 | PASS | `PRD-003-TASK-04`, `PRD-003-TASK-06` |
| NFR-003 | PASS | `PRD-003-TASK-03`, `PRD-003-TASK-06` |
| NFR-004 | PASS | `PRD-003-TASK-01`, `PRD-003-TASK-06` |

# Risks and Dependencies

Risks:
- CLI parser updates may unintentionally regress existing direct-launch behavior.
- Help copy may drift from real behavior if docs and tests are not kept aligned.
- Build environments without revision metadata could create inconsistent expectations if fallback is unclear.

Dependencies:
- Existing startup CLI entrypoint behavior and tests.
- Existing documentation structure in `docs/` and `README.md`.
- Build metadata availability for short hash output (with explicit `dev` fallback).

# State & Failure Matrix

| Flow Area | Trigger / Failure Mode | Expected Product Response | User-Visible Recovery Path |
| --- | --- | --- | --- |
| startup | User runs `dbc --help` or `dbc -h` | Print deterministic help to stdout and exit `0` | Use provided usage/examples to run correct command |
| startup | User runs `dbc --version` or `dbc -v` | Print single-token hash-or-dev to stdout and exit `0` | Use version token for support/debug/automation needs |
| startup | User passes invalid startup args | Print actionable error/hint/usage guidance and exit `2` | Correct arguments and retry |
| config | Runtime configuration CRUD/state transitions | Out of scope for this release | Existing behavior remains unchanged |
| save | Data persistence/save workflow | Out of scope for this release | Existing behavior remains unchanged |
| navigation | In-session navigation/context switching | Out of scope for this release | Existing behavior remains unchanged |

# Assumptions

- A1 (High): For this release, repeated logical informational flags (for example both short and long alias for the same flag) are treated as invalid usage for deterministic contract enforcement.
- A2 (High): Startup CLI scope is intentionally limited to touched entrypoint behavior and does not require repo-wide UX wording harmonization.
- A3 (Medium): Release validation evidence from automated tests and checklist audit is sufficient for success-metric evaluation in this scope.
- A4 (High): Standards references in product/technical docs are sufficient and preferred over duplicating standards text.

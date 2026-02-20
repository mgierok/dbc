# Overview

We believe that defining a quality-first, full-coverage manual regression scenario suite for DBC maintainers and release validators will reduce release risk and improve confidence in startup and runtime stability.
We will know this is true when all required regression coverage and quality metrics reach target values during release validation for this scope.

# Metadata

- Status: READY

# Problem Statement

The current regression scenario set does not provide full behavioral coverage for critical DBC user journeys. This creates a risk that user-facing regressions in startup, configuration management, runtime interaction, and save/navigation safety paths can reach release without being detected by the manual regression process.

# Current State (As-Is)

- The repository currently contains one regression case: `TC-001-direct-launch-opens-main-view`.
- Existing coverage is concentrated on direct launch happy path verification.
- Full-suite coverage across all required areas (`user journey`, `TUI behavior`, `critical path`) is not yet systematically defined as a complete manual regression package.
- Failure and user-visible recovery expectations are not fully represented across all critical journeys.

# Target State After Release (To-Be)

- A complete PRD-defined manual regression scenario package exists, aligned with `docs/test-case-specification.md`.
- The scenario suite covers all critical DBC journeys end-to-end, including failure and recovery outcomes where relevant.
- Every scenario follows the required test-case contract, deterministic PASS/FAIL rules, and startup-script binding rules.
- Release validation uses explicit quality gates based on coverage and evidence artifacts rather than scenario count targets.

# Business Rationale and Strategic Fit

DBC is positioned as a predictable terminal-first data workflow tool. Regressions in startup, interaction flow, or save safety directly damage trust and increase operational risk. A quality-first full regression specification supports reliable releases, aligns with current product scope, and reduces the chance of high-impact user-facing failures escaping validation.

# Goals

- G1: Define a full-coverage manual regression scope for the current DBC product behavior, focused on quality and completeness rather than scenario quantity.
- G2: Ensure every critical journey has explicit, deterministic pass/fail assertions and clear user-visible recovery behavior for failure paths.
- G3: Establish measurable release validation criteria with concrete evidence artifacts that can be produced during execution.

# Non-Goals

- NG1: Defining or implementing automated test frameworks, pipelines, or automation code.
- NG2: Introducing new product features, behavior changes, or architecture changes.
- NG3: Optimizing for a fixed target count of scenarios as a delivery objective.

# Scope (In Scope / Out of Scope)

In Scope:
- Manual regression scenario definition for full end-to-end user journeys.
- Coverage of startup, selector/config workflows, runtime TUI behavior, save behavior, and navigation safety with dirty-state handling.
- Mandatory inclusion of critical failure and user-visible recovery expectations in scenario design.
- Compliance with `docs/test-case-specification.md` and `docs/test-case-template.md`.
- Deterministic PASS/FAIL assertion design with explicit evidence expectations.

Out of Scope:
- Automation implementation or automation tooling selection.
- Any runtime feature implementation, refactor, or behavior change in application code.
- Non-SQLite engine behavior and integrations outside documented current product scope.

# Functional Requirements

FR-001:
- The regression package must define end-to-end manual scenarios that collectively cover all critical journey areas in current DBC scope: startup, selector/config management, runtime TUI interaction, save operations, and context/navigation transitions.
- Acceptance: A coverage matrix maps every required journey area to one or more scenario IDs with no uncovered required area.

FR-002:
- Each scenario must bind exactly one startup script from the approved startup script catalog and include both script path and exact startup command.
- Acceptance: Scenario metadata shows exactly one valid startup script binding and one exact command for every scenario.

FR-003:
- Each scenario must follow the mandatory section order and required fields from `docs/test-case-template.md`.
- Acceptance: A structure audit confirms all scenario files match the required template headings and required table fields.

FR-004:
- Each step in every scenario must contain one user action, one expected outcome, and one linked assertion ID.
- Acceptance: Step-level audit confirms one-to-one mapping of action, expected outcome, and assertion reference for all steps.

FR-005:
- Each scenario must define a single observable expected result and deterministic assertion criteria with explicit evidence capture.
- Acceptance: Assertions in all scenarios are measurable and resolvable to only `PASS` or `FAIL` without ambiguity.

FR-006:
- Critical journeys must include both normal flow validation and failure/recovery validation where user-facing failure can occur.
- Acceptance: For each critical journey in scope, at least one scenario or scenario branch documents trigger/failure mode, expected response, and user-visible recovery path.

FR-007:
- The suite must prioritize expanded, context-rich scenarios instead of splitting into many low-value single-assertion files.
- Acceptance: Scenario quality review confirms each scenario contains multiple context-relevant assertions and avoids redundant fragmentation.

FR-008:
- Release readiness for this scope must require complete assertion success across the defined regression suite.
- Acceptance: Release decision record states `PASS` only when all scenario assertions are `PASS`; any unmet precondition or failed expectation results in `FAIL`.

# Non-Functional Product Requirements

NFR-001:
- The regression suite definition must be reproducible by any qualified operator using repository-documented scripts and fixture data without additional hidden setup.

NFR-002:
- Scenario language must remain clear, unambiguous, and behavior-oriented, enabling consistent interpretation by different reviewers.

NFR-003:
- Evidence references must be concrete and auditable (for example scenario results table, assertion evidence field, and coverage checklist outputs).

NFR-004:
- The specification must remain stable against scope creep by explicitly documenting non-goals and out-of-scope boundaries.

# Success Metrics and Release Criteria

Primary Outcome Metric:
- M1: Required journey-area coverage completeness.
  - Baseline: `20%` (`1/5` required journey areas currently covered by existing suite: startup only).
  - Target: `100%` (`5/5` required journey areas covered by approved scenarios).
  - Measurement window: During PRD execution and regression suite validation for this scope.
  - Measurement method: Coverage matrix audit artifact mapping scenario IDs to required journey areas (`startup`, `selector/config`, `runtime/TUI`, `save`, `navigation`) and signed review checklist.

Leading Indicators:
- M2: Critical journey failure/recovery coverage ratio.
  - Baseline: `0%` (`0/5` required journey areas currently documented with explicit failure + user-visible recovery validation in existing suite).
  - Target: `100%` (`5/5` required journey areas include explicit failure/recovery validation where applicable).
  - Measurement window: During scenario definition review and pre-release regression audit.
  - Measurement method: Failure/recovery mapping checklist artifact linked to scenario IDs and assertion IDs.
- M3: Specification compliance rate against test case standard/template.
  - Baseline: `100%` for existing single case (`1/1`) on mandatory metadata and structure fields.
  - Target: `100%` for full suite (`N/N`) on mandatory metadata, structure order, startup binding, and deterministic result policy.
  - Measurement window: During suite review before release decision.
  - Measurement method: Structured compliance audit report against `docs/test-case-specification.md` and `docs/test-case-template.md`.

Guardrail Metric:
- M4: Determinism integrity violations in final suite.
  - Baseline: `0` known violations in existing scenario set.
  - Target: `0` violations in the full suite.
  - Measurement window: Final regression suite quality audit before release decision.
  - Measurement method: Determinism audit artifact checking for forbidden result states and ambiguous pass criteria.

Release Criteria:
- All metrics (`M1`, `M2`, `M3`, `M4`) meet target values.
- Every scenario and every assertion in the scoped suite is marked `PASS`.
- No unresolved placeholders, open decisions, or contract violations remain in final scenario artifacts.

# Risks and Dependencies

Risks:
- Incomplete interpretation of current product behavior could leave hidden coverage gaps.
- Overly generic scenarios may pass formal checks but fail to detect meaningful regressions.
- Manual execution variability can reduce confidence if evidence fields are not rigorously maintained.

Dependencies:
- Current product behavior source-of-truth in `docs/product-documentation.md`.
- Current technical behavior source-of-truth in `docs/technical-documentation.md`.
- Test case contract in `docs/test-case-specification.md` and template in `docs/test-case-template.md`.
- Availability of fixture and startup scripts under `scripts/`.

# State & Failure Matrix

| Flow Area | Trigger / Failure Mode | Expected Product Response | User-Visible Recovery Path |
| --- | --- | --- | --- |
| startup | Startup mode mismatch, invalid startup target, or startup argument misuse in a scoped journey. | Product follows documented startup contracts with explicit error/status behavior and deterministic outcome. | User can rerun with valid startup mode/arguments or use selector-driven path according to scenario contract. |
| config | Config entry invalid, malformed, missing, or unreachable during selector/config workflow. | Product blocks unsafe progression, surfaces clear status/error context, and keeps user in manageable selector/config context. | User edits/adds/selects a valid entry and continues flow without process ambiguity. |
| save | Save operation failure after staged data actions. | Product retains staged state and surfaces failure status without silent data loss. | User can correct issue and retry save, or choose explicit alternative action from documented flow. |
| navigation | User attempts table/context switch or `:config` routing with dirty state present. | Product requires explicit user decision (`save`, `discard`, or `cancel`) before context transition. | User selects a decision path and product behaves predictably per selected recovery/continuation option. |

# Assumptions

- A1 (High): This PRD governs manual regression scenario quality and coverage for the current product state only.
- A2 (High): Quality-first means coverage completeness and deterministic validation are prioritized over any fixed scenario count.
- A3 (High): Existing `TC-001` is treated as factual baseline artifact for current measured state.
- A4 (Medium): Journey-area grouping (`startup`, `selector/config`, `runtime/TUI`, `save`, `navigation`) is sufficient to represent full regression scope for current release decisions.

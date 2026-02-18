# Overview

This PRD defines a testability enablement release for `dbc` focused on a stable local SQLite fixture and reproducible tmp-environment startup playbooks that support agent-driven and manual validation.

We believe that introducing a canonical test fixture and explicit tmp startup guides for AI coding agents will reduce setup friction and increase repeatable test execution quality.
We will know this is true when startup-playbook and fixture-coverage release targets are met during PR validation.

# Metadata

- Status: READY

# Problem Statement

Current testing relies on ad-hoc local databases and implicit startup knowledge, making Codex-style testing inconsistent, slower to set up, and difficult to reproduce across runs.

# Current State (As-Is)

- The repository has no canonical fixture database at `docs/test.db`.
- There is no dedicated `docs/` guide for running `dbc` in an isolated tmp context with all required startup variants.
- There is no standardized example test scenario combining fixture data, startup method, and expected navigation outcomes.

# Target State After Release (To-Be)

- The repository includes a canonical fixture database at `docs/test.db`.
- The fixture contains several related tables with coherent data and balanced edge cases while remaining intentionally small for fast tests.
- Documentation in `docs/` explains tmp-environment startup in three variants:
  - database provided via `-d`,
  - database provided via config file,
  - startup without passing a database parameter.
- Each variant includes executable command listings suitable for Codex execution.
- Each variant includes specificity notes and example usage situations.
- Documentation includes at least one example manual test using the fixture and one selected startup method.

# Business Rationale and Strategic Fit

- Improves repeatability of local validation flows for AI coding tools.
- Reduces onboarding friction for contributors executing test and navigation scenarios.
- Increases confidence in behavior checks without changing runtime feature scope.
- Supports current product direction as a terminal-first SQLite workflow.

# Goals

- G1. Provide one canonical, fast fixture DB for repeatable local testing.
- G2. Document deterministic tmp startup procedures for the three required launch variants.
- G3. Provide one reusable manual test example tied to fixture data and startup flow.
- G4. Improve pre-implementation and regression verification quality for Codex-style workflows.

# Non-Goals

- Adding new runtime capabilities or startup flags.
- Expanding support beyond SQLite.
- Delivering an automated TUI key-driving framework in this release.
- Refactoring product architecture or application behavior.

# Scope (In Scope / Out of Scope)

In Scope:
- Add `docs/test.db` as canonical repository fixture.
- Define fixture expectations: related tables, coherent records, balanced edge cases, small size.
- Add/update `docs/` guidance for tmp startup in three variants (`-d`, config file, no database argument).
- Include executable command listings for each variant.
- Include one example manual test scenario using fixture data and one startup method.

Out of Scope:
- Any runtime behavior changes.
- Additional startup entry modes beyond currently supported flows.
- CI framework changes for scripted interactive TUI automation.
- Large or performance-focused data fixture sets.

# Functional Requirements

FR-001: The release must provide `docs/test.db` as the canonical local fixture database for testing `dbc`.
Acceptance: `docs/test.db` exists in repository and is identified in docs as the standard fixture.

FR-002: The fixture must include several tables with explicit relational links and coherent cross-table data.
Acceptance: Fixture documentation confirms relation map and consistency of linked records.

FR-003: The fixture must include balanced core edge-case categories relevant for current product behavior.
Acceptance: Fixture documentation confirms inclusion of null/default/not-null/unique/foreign-key/check plus empty and long values and varied SQLite data types.

FR-004: The fixture must remain small enough for fast local testing loops.
Acceptance: Validation checklist confirms fixture size and row volume stay within agreed small-fixture threshold.

FR-005: Documentation must cover tmp startup with database path passed via `-d`, including executable command listings.
Acceptance: A tester can execute listed commands and observe expected direct-launch startup behavior.

FR-006: Documentation must cover tmp startup with database passed through config file, including executable command listings.
Acceptance: A tester can execute listed commands and observe expected config-driven startup behavior.

FR-007: Documentation must cover tmp startup without passing a database argument, including executable command listings.
Acceptance: A tester can execute listed commands and observe expected no-parameter startup behavior.

FR-008: Each startup variant section must describe variant-specific behavior and at least one practical usage situation.
Acceptance: Each section contains explicit "specific behavior" notes and at least one "when to use" scenario.

FR-009: Documentation must include one example manual test scenario based on fixture data and one startup method.
Acceptance: Scenario lists steps, expected observations, and pass/fail criteria.

# Non-Functional Product Requirements

NFR-001: Documentation must be clear enough to execute startup flows without hidden project knowledge.

NFR-002: Fixture content must prioritize deterministic repeatability over dataset breadth.

NFR-003: Added assets must not change existing startup and runtime user-visible behavior.

NFR-004: Command listings must be copy-paste ready and compatible with isolated tmp execution flow.

# Success Metrics and Release Criteria

Primary Outcome Metric:
- M1 Startup playbook completion rate.
Baseline: 0 of 3 required tmp startup variants are documented with executable listings.
Target: 3 of 3 required tmp startup variants documented and validated.
Measurement window: PR validation before merge.
Measurement method: PR checklist artifact verifying each variant section and command execution evidence.

Leading Indicators:
- M2 Fixture coverage checklist completeness.
Baseline: 0 canonical fixtures with documented balanced edge-case coverage.
Target: 1 canonical fixture with full required coverage checklist.
Measurement window: PR validation before merge.
Measurement method: Fixture audit checklist attached to PR with required categories marked pass/fail.

- M3 Example manual test reproducibility.
Baseline: 0 standardized manual scenario bound to canonical fixture.
Target: 1 standardized manual scenario executed end-to-end with expected outcomes.
Measurement window: PR validation before merge.
Measurement method: Execution notes artifact recording steps and observed outcomes for the documented scenario.

Guardrail Metric:
- M4 Startup regression count in existing behaviors.
Baseline: 0 known regressions for current startup flows.
Target: 0 regressions introduced by this release.
Measurement window: PR validation before merge.
Measurement method: Verification run against current startup behavior expectations documented in product/technical docs.

Release Criteria:
- FR-001 through FR-009 acceptance statements are satisfied.
- M1 through M4 targets are achieved.
- Out-of-scope boundaries remain unchanged (no runtime behavior expansion).

# Risks and Dependencies

Risks:
- Fixture may become too synthetic and miss practical navigation patterns.
- Fixture may grow over time and reduce speed of local testing loops.
- Documentation may drift from actual startup behavior if not validated during updates.

Dependencies:
- Existing startup support for `-d` and config-based selection flows.
- Existing no-parameter startup behavior and first-entry configuration paths.
- Existing documentation structure under `docs/`.

# State & Failure Matrix

| Flow Area | Trigger / Failure Mode | Expected Product Response | User-Visible Recovery Path |
| --- | --- | --- | --- |
| startup | Launch with `-d` and valid fixture path | Application starts through direct-launch path | Correct path and rerun if startup fails due to path issue |
| startup | Launch with config file in tmp and invalid entry | Startup shows configuration/connection failure per current behavior | Fix tmp config entry and rerun |
| startup | Launch without database parameter in tmp and no configured DB | Application follows current no-config startup behavior | Complete initial configuration flow and continue |
| config | Tmp config file is missing or malformed | Application follows existing config error/empty handling | Recreate valid config file from documented commands |
| save | Save-related behavior during example test flow | Out of scope for this release | Use existing save behavior as currently documented |
| navigation | Expected navigation checkpoint is not reached in manual scenario | Scenario identifies mismatch as failed observation | Re-run scenario and report reproducible mismatch |

# Assumptions

- A1 (High): English documentation is acceptable for the new tmp startup and fixture testing guides.
- A2 (High): Small-fixture threshold values will be explicitly defined in delivery validation materials and kept stable.
- A3 (Medium): PR-attached execution notes are sufficient evidence for release metric measurement in this scope.
- A4 (High): Committed `docs/test.db` is the preferred fixture distribution model for this release.

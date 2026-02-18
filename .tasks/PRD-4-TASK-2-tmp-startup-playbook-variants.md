# Overview

Document deterministic tmp-environment startup playbooks for the three required launch variants so contributors and agents can reproduce startup behavior consistently.

## Metadata

- Status: DONE
- PRD: PRD-4-agent-testability-tmp-startup-fixture.md
- Task ID: 2
- Task File: PRD-4-TASK-2-tmp-startup-playbook-variants.md
- PRD Requirements: FR-005, FR-006, FR-007, FR-008, NFR-001, NFR-004
- PRD Metrics: M1

## Objective

Provide executable tmp startup instructions for `-d`, config-file startup, and no-parameter startup, including variant-specific behavior and usage context.

## Working Software Checkpoint

After this task, startup behavior is unchanged, and users can execute copy-paste startup commands in isolated tmp environments for all required variants.

## Technical Scope

### In Scope

- Add or update docs under `docs/` with startup playbooks for three variants.
- Include copy-paste command listings for each variant.
- Include "specific behavior" and "when to use" notes for each variant.
- Ensure instructions are aligned with current startup behavior contracts.

### Out of Scope

- New startup flags or argument semantics.
- Changes to runtime startup implementation.
- CI automation for startup playbook execution.

## Implementation Plan

1. Draft tmp setup prerequisites and a consistent environment bootstrap sequence.
2. Add variant section for direct launch via `-d` with executable commands and expected outcomes.
3. Add variant section for config-file startup with executable commands and expected outcomes.
4. Add variant section for no-parameter startup with executable commands and expected outcomes.
5. For each variant, add explicit "specific behavior" and at least one "when to use" scenario.
6. Run all listed command flows and record observed outcomes for closure evidence.

## Verification Plan

- FR-005 happy-path check: execute documented `-d` variant commands in tmp context and confirm direct-launch startup behavior is observed.
- FR-005 negative-path check: execute `-d` flow with invalid DB path and confirm expected startup failure guidance is observed.
- FR-006 happy-path check: execute documented config-file variant commands and confirm config-driven startup behavior is observed.
- FR-006 negative-path check: execute flow with malformed or unreachable config entry and confirm expected startup/config failure behavior is observed.
- FR-007 happy-path check: execute documented no-parameter startup flow and confirm expected selector/mandatory setup behavior is observed.
- FR-007 negative-path check: execute no-parameter flow with intentionally missing tmp config and confirm expected recovery path is observed.
- FR-008 happy-path check: confirm each variant section contains explicit "specific behavior" notes and "when to use" scenario.
- FR-008 negative-path check: run section-completeness checklist and confirm it fails if either required subsection is absent.
- Metric checkpoint (M1):
  - Metric ID: M1
  - Evidence source artifact: `Completion Summary` entry in this task file with execution outcomes for all three variants.
  - Threshold/expected value: `3/3` required startup variants documented and validated.
  - Check procedure: execute each documented variant flow and record pass/fail plus observed behavior in `Completion Summary`.

## Acceptance Criteria

- Documentation contains executable startup flows for `-d`, config-file startup, and no-parameter startup.
- Each variant section includes clear behavior notes and at least one practical usage situation.
- Command listings are copy-paste ready for isolated tmp execution.
- Documented startup flows are consistent with current product and technical documentation.
- Project validation requirement: affected tests and checks pass (`go test ./...`, `golangci-lint run ./...`).

## Dependencies

- blocked-by: [PRD-4-TASK-1-fixture-foundation-and-coverage-contract](.tasks/PRD-4-TASK-1-fixture-foundation-and-coverage-contract.md)
- blocks: [PRD-4-TASK-3-manual-scenario-reproducibility](.tasks/PRD-4-TASK-3-manual-scenario-reproducibility.md), [PRD-4-TASK-4-integration-hardening](.tasks/PRD-4-TASK-4-integration-hardening.md)

## Completion Summary

Completed deliverables:

1. Updated `docs/test-fixture.md` with deterministic tmp startup playbooks for all required variants:
   - direct launch via `-d`,
   - config-file startup,
   - startup without database parameter.
2. Added shared tmp bootstrap and cleanup flow using:
   - `TMP_ROOT`, `TMP_HOME`, `TMP_DB`,
   - temporary compiled binary (`go build -o "$DBC_BIN" ./cmd/dbc`).
3. Added executable command listings for each variant (happy + negative checks), plus required:
   - `Specific behavior` notes,
   - `When to use` scenarios.

Verification evidence (FR + M1):

- FR-005:
  - happy: `HOME="$TMP_HOME" "$DBC_BIN" -d "$TMP_DB"` opened runtime directly (main view with fixture tables including `categories`, `customers`, `orders`) and exited with status `0` after `q`.
  - negative: `HOME="$TMP_HOME" "$DBC_BIN" -d "$TMP_ROOT/missing.db"` failed with direct-launch guidance (`Cannot start DBC with direct launch target ...`) and exited with status `1`.
- FR-006:
  - happy: with valid `"$TMP_HOME/.config/dbc/config.toml"`, `HOME="$TMP_HOME" "$DBC_BIN"` opened selector-first startup, showing tmp config path and `fixture` option, then exited cleanly with status `0`.
  - negative: malformed config (`name = "fixture` without closing quote) failed at startup with `toml: basic strings cannot have new lines` and exited with status `1`.
- FR-007:
  - happy: with valid config and no `-d`, `HOME="$TMP_HOME" "$DBC_BIN"` opened selector-first startup; pressing `Enter` launched runtime to the main fixture view and exited with status `0`.
  - negative: with intentionally missing config file, `HOME="$TMP_HOME" "$DBC_BIN"` opened mandatory first-entry setup (`Add database` form with `Name`/`Path`), and `Esc` provided recovery by canceling startup with status `0`.
- FR-008:
  - happy: section-completeness check on `docs/test-fixture.md` returned `happy_specific=3` and `happy_when_to_use=3`.
  - negative: simulated removal of one `When to use:` subsection reduced count to `2`, and completeness check failed as expected.

Metric checkpoint:

- M1 PASS: all `3/3` required startup variants are documented with executable flows and validated with recorded happy/negative outcomes.

Project validation:

- `golangci-lint run ./...` -> `0 issues.`
- `go test ./...` -> pass for all packages.

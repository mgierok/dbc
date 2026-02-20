# Full-Suite Release Readiness Audit

## Scope

- Scenario set: `TC-001` through `TC-008` in `test-cases/`
- Governance artifacts:
  - `test-cases/suite-coverage-matrix.md`
  - `test-cases/scenario-structure-and-metadata-checklist.md`
  - `test-cases/deterministic-result-audit-checklist.md`
  - `docs/test-case-specification.md`
  - `docs/test-case-template.md`

## Cross-Artifact Consistency Rules

Result is `FAIL` when at least one of these is true:

- metadata fields differ between specification, template, and checklist,
- assertion columns differ between template and checklist,
- startup script referenced in a scenario is not present in startup scripts catalog,
- coverage matrix omits scenario or assertion mappings for a tracked Functional Behavior reference,
- deterministic checklist omits `Violation Count` or fail trigger for non-zero count.

## Governance Contract Audit

| Contract | Audit Check | Evidence | Result (`PASS`/`FAIL`) |
| --- | --- | --- | --- |
| One-reference ownership | Scenario contract requires exactly one metadata Functional Behavior reference. | `docs/test-case-specification.md` + `docs/test-case-template.md` + `test-cases/scenario-structure-and-metadata-checklist.md` | `PASS` |
| Assertion purity | Assertion contract requires one reference per assertion and equality to metadata reference. | `docs/test-case-specification.md` + `test-cases/deterministic-result-audit-checklist.md` | `PASS` |
| Expand-first evidence model | Release audit includes explicit expanded-vs-new classification and ratio inputs. | `Expand-First Coverage Addition Evidence` section in this file | `PASS` |
| Area traceability | Coverage matrix maps Functional Behavior reference -> scenario IDs -> assertion IDs. | `test-cases/suite-coverage-matrix.md` | `PASS` |
| Informational startup coverage readiness | Startup scripts catalog contains informational startup binding command. | `docs/test-case-specification.md` startup scripts catalog + `scripts/start-informational.sh` | `PASS` |
| Cross-artifact mismatch fail policy | Cross-artifact mismatch fail triggers are explicitly defined. | `Cross-Artifact Consistency Rules` section in this file | `PASS` |

## Full-Suite Integration Audit Results

| Check | Observed | Result (`PASS`/`FAIL`) |
| --- | --- | --- |
| Ownership metadata compliance (`FR-001`) | All active scenarios (`TC-001` to `TC-008`) contain exactly one metadata `Functional Behavior Reference` row (`8/8`). | `PASS` |
| Assertion purity compliance (`FR-002`) | Every scenario resolves to one Functional Behavior anchor only (`unique anchor count = 1` for each file), with `52/52` assertion rows carrying matching reference links. | `PASS` |
| Startup script binding contract | Scenario startup scripts are within specification catalog (`start-direct-launch.sh`, `start-selector-from-config.sh`, `start-informational.sh`) with no unknown script usage. | `PASS` |
| Required structure and headings contract | Required headings `## 1` through `## 7` are present for all active scenarios (`8/8` per heading). | `PASS` |
| Deterministic binary outcomes | `52` assertion rows are binary `PASS/FAIL`; `8` final results are binary and `PASS`; no third-state markers detected. | `PASS` |
| Functional Behavior matrix completeness (`FR-004`) | `test-cases/suite-coverage-matrix.md` contains `8` rows (`4.1` to `4.8`), all with completeness/purity marked `PASS`. | `PASS` |
| Governance artifact synchronization (`FR-010`) | Matrix, structure checklist, determinism checklist, and this release audit all reference active scenario set `TC-001` to `TC-008` and aligned contracts. | `PASS` |

## Expand-First Coverage Addition Evidence

| Coverage Change ID | Delivery Method (`Expanded Existing TC` \| `New TC`) | Expanded Scenario IDs | New Scenario IDs | Expand-First Evidence | Counts Toward Expanded Numerator (`0/1`) | Counts Toward Total Denominator (`0/1`) | Result (`PASS`/`FAIL`) |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `REF-001` | `Expanded Existing TC` | `TC-001` | `none` | Refactored existing `TC-001` to area-pure `4.1` with informational startup script binding (`help`/`version`). | `1` | `1` | `PASS` |
| `REF-002` | `Expanded Existing TC` | `TC-002` | `none` | Refactored existing `TC-002` to area-pure `4.2` layout/focus ownership assertions. | `1` | `1` | `PASS` |
| `REF-003` | `Expanded Existing TC` | `TC-003` | `none` | Refactored existing `TC-003` to area-pure `4.3` table discovery and schema assertions. | `1` | `1` | `PASS` |
| `REF-004` | `Expanded Existing TC` | `TC-004` | `none` | Refactored existing `TC-004` to area-pure `4.4` records/navigation assertions. | `1` | `1` | `PASS` |
| `REF-005` | `Expanded Existing TC` | `TC-005` | `none` | Refactored existing `TC-005` to area-pure `4.6` insert/edit/delete ownership with deterministic operation assertions. | `1` | `1` | `PASS` |
| `REF-006` | `Expanded Existing TC` | `TC-006` | `none` | Refactored existing `TC-006` to area-pure `4.7` staging lifecycle, undo/redo, and dirty-decision assertions. | `1` | `1` | `PASS` |
| `REF-007` | `New TC` | `none` | `TC-007` | Added dedicated area-pure filtering scenario (`4.5`) after expand-first completion for `TC-001` to `TC-006`; new scenario was required to preserve one-area purity and keep filter-flow assertions readable. | `0` | `1` | `PASS` |
| `REF-008` | `New TC` | `none` | `TC-008` | Added dedicated area-pure visual-state scenario (`4.8`) after expand-first completion for `TC-001` to `TC-006`; new scenario was required to preserve one-area purity and keep visual-indicator checks focused. | `0` | `1` | `PASS` |

### Expand-First Ratio Formula

- Expanded-first adherence ratio = `sum(Expanded Numerator) / sum(Total Denominator)` for rows with denominator `1`.
- If denominator sum is `0`, ratio is `N/A` for the snapshot and readiness is evaluated on contract availability.
- Current ratio after current expansion evidence rows: `6/8 = 75%` (`PASS` against M3 threshold `>=70%`).

## Determinism Violation Checkpoint

- Violation Count Source: `test-cases/deterministic-result-audit-checklist.md`
- Required threshold: `0`
- Current observed full-suite value: `0` (`TC-001` to `TC-008`).
- Gate rule: any non-zero value forces `FAIL`.

## Requirements Closure (`FR`/`NFR`)

| Requirement ID | Closure Evidence | Result (`PASS`/`FAIL`) |
| --- | --- | --- |
| `FR-001` | Ownership audit across all active scenarios reports exactly one metadata Functional Behavior reference per file (`8/8`). | `PASS` |
| `FR-002` | Assertion purity audit reports no cross-area reference drift and no metadata/assertion mismatch across `52` assertion rows. | `PASS` |
| `FR-003` | Expand-first evidence table contains `REF-001` to `REF-008`, with new scenarios (`TC-007`, `TC-008`) added only after refactor-first sequence and explicit rationale. | `PASS` |
| `FR-004` | Coverage matrix maps all Functional Behavior areas `4.1` through `4.8` with non-empty scenario/assertion IDs and `PASS` status. | `PASS` |
| `FR-005` | `TC-001` validates informational startup behavior through approved `scripts/start-informational.sh` binding with deterministic `help`/`version` assertions (`A1` to `A3`). | `PASS` |
| `FR-006` | `TC-007` covers filter flow sequence (column/operator/value), apply behavior, one-active replacement, and table-switch reset (`A1` to `A7`). | `PASS` |
| `FR-007` | `TC-005` deterministically covers insert/edit/delete operation behavior, including pending-row removal semantics (`A1` to `A10`). | `PASS` |
| `FR-008` | `TC-006` deterministically covers staged lifecycle with undo/redo and dirty `:config` decisions (`cancel`, `save`, `discard`) (`A1` to `A11`). | `PASS` |
| `FR-009` | `TC-008` verifies required visual communication markers (`READ-ONLY`, `WRITE (dirty: N)`, `*`, `[INS]`, `[DEL]`) and status-line context (`A1` to `A6`). | `PASS` |
| `FR-010` | Governance artifacts remain synchronized and include active scenario ownership/traceability evidence with no stale mapping. | `PASS` |
| `NFR-001` | All active scenarios remain reproducible via repository startup scripts and fixture-backed startup flows. | `PASS` |
| `NFR-002` | Scenario metadata ownership is unambiguous (`1` Functional Behavior reference per file). | `PASS` |
| `NFR-003` | Traceability remains auditable via Functional Behavior matrix area -> scenario -> assertion mapping. | `PASS` |
| `NFR-004` | Readability guardrail maintained: `6/8` additions delivered via expansion, and `2/8` new scenarios include explicit readability/purity rationale. | `PASS` |

## Metrics Closure (`M1` to `M4`)

| Metric | Threshold | Final Evidence | Result (`PASS`/`FAIL`) |
| --- | --- | --- | --- |
| `M1` | `100%` ownership compliance | `8/8 = 100%` scenarios pass one-area ownership contract. | `PASS` |
| `M2` | `100%` area coverage (`8/8`) | Coverage matrix reports `8/8` Functional Behavior areas mapped with assertion traceability. | `PASS` |
| `M3` | `>=70%` expand-first adherence | Expand-first ratio = `6/8 = 75%`. | `PASS` |
| `M4` | `0` determinism integrity violations | Full-suite determinism violation count = `0`. | `PASS` |

## Release Criteria Closure

| Release Criterion | Evidence | Result (`PASS`/`FAIL`) |
| --- | --- | --- |
| `M1` to `M4` meet target thresholds | `M1=100%`, `M2=8/8`, `M3=75%`, `M4=0`. | `PASS` |
| Every area `4.1` to `4.8` has mapped scenario assertions | Coverage matrix rows `4.1` to `4.8` each contain non-empty scenario and assertion IDs. | `PASS` |
| Every active scenario passes ownership and determinism audits | Ownership/purity/determinism checks pass for `TC-001` to `TC-008`. | `PASS` |
| Governance artifacts are synchronized with no unresolved violations | Cross-artifact checks pass for matrix + checklists + release audit. | `PASS` |

## Release Decision (Final Snapshot)

- Governance contract readiness: `PASS`
- Scenario conformance readiness: `PASS`
- Go/No-Go: `GO`
- Rationale: All PRD-006 FR/NFR requirements and metric thresholds are closed with deterministic full-suite evidence, and release criteria are fully satisfied.

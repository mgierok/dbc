# Suite Coverage Matrix

## Purpose

Define mandatory journey-area coverage mapping and deterministic pass/fail coverage review rules for the manual regression suite.

## Required Journey Areas

- `startup`
- `selector/config`
- `runtime/TUI`
- `save`
- `navigation`

## Coverage Review Rules

1. Coverage review is `FAIL` when any required journey area has an empty `Scenario IDs` value.
2. Coverage review is `FAIL` when a listed scenario ID does not resolve to an existing `test-cases/TC-*.md` file.
3. Coverage review is `PASS` only when every required journey area has at least one valid mapped scenario ID.

## Mapping Matrix

| Journey Area | Scenario IDs | Failure/Recovery Required | Failure/Recovery Scenario IDs | Coverage Status (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- | --- |
| startup | `TC-001, TC-002` | Yes | `TC-002` | `PASS` | `test-cases/TC-001-direct-launch-opens-main-view.md`, `test-cases/TC-002-empty-config-startup-recovers-through-first-entry-setup.md` |
| selector/config | `TC-002, TC-003` | Yes | `TC-002, TC-003` | `PASS` | `test-cases/TC-002-empty-config-startup-recovers-through-first-entry-setup.md`, `test-cases/TC-003-selector-edit-invalid-path-blocks-save-until-corrected.md` |
| runtime/TUI | `TC-004` | Yes | `TC-004` | `PASS` | `test-cases/TC-004-runtime-command-failure-recovery-keeps-session-usable.md` |
| save | `TC-005` | Yes | `TC-005` | `PASS` | `test-cases/TC-005-save-failure-retains-staged-changes-until-corrected.md` |
| navigation | `TC-006` | Yes | `TC-006` | `PASS` | `test-cases/TC-006-dirty-config-navigation-requires-explicit-decision.md` |

## Current Baseline Conclusion

- Coverage review result: `PASS`
- Reason: All required journey areas are mapped to valid `TC-*` scenario files, including explicit failure/recovery coverage for each area.

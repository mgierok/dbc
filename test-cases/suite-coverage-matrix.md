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
| startup | `TC-001` | Yes | `none` | `PASS` | `test-cases/TC-001-direct-launch-opens-main-view.md` |
| selector/config | `none` | Yes | `none` | `FAIL` | No mapped selector/config scenario yet. |
| runtime/TUI | `none` | Yes | `none` | `FAIL` | No mapped runtime/TUI scenario yet. |
| save | `none` | Yes | `none` | `FAIL` | No mapped save scenario yet. |
| navigation | `none` | Yes | `none` | `FAIL` | No mapped navigation scenario yet. |

## Current Baseline Conclusion

- Coverage review result: `FAIL`
- Reason: Required journey areas `selector/config`, `runtime/TUI`, `save`, and `navigation` are currently unmapped.

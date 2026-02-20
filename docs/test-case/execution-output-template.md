# Test Case Execution Output Template

This template defines display format only. Execution results must be shown after run and must not be persisted as repository artifacts.

## Single Test Case Output

- Execution Type: `SINGLE`
- Case ID: `TC-<NNN>`
- Functional Behavior Reference: `[<4.x Section Title>](../docs/product-documentation.md#<anchor-under-4-functional-behavior>)`
- Startup Script: `scripts/<script-name>.sh`
- Startup Command: `bash scripts/<script-name>.sh`

| Assertion ID | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- |
| `A1` | `PASS` | `<observable proof>` |
| `A2` | `PASS` | `<observable proof>` |

- Test Result: `PASS` | `FAIL`
- Failed Assertions: `<IDs or 'none'>`
- Failure Reason: `<required when result is FAIL>`

## Full Suite Output

- Execution Type: `SUITE`
- Scenario Scope: `TC-<NNN> ... TC-<NNN>`
- Total Cases: `<number>`
- Passed Cases: `<number>`
- Failed Cases: `<number>`

| Case ID | Result (`PASS`/`FAIL`) | Failed Assertions | Evidence Summary |
| --- | --- | --- | --- |
| `TC-001` | `PASS` | `none` | `<short proof>` |
| `TC-002` | `FAIL` | `A3` | `<short proof>` |

- Suite Result: `PASS` | `FAIL`
- Failure Summary: `<required when suite result is FAIL>`

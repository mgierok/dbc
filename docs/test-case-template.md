# <Scenario Title>

Use a short, clear title describing the tested behavior.
Recommended filename: `TC-<NNN>-<behavior>-<expected-result>.md`.

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-<NNN>` |
| Functional Behavior Reference | `[<Section title>](../docs/product-documentation.md#<section-anchor-under-functional-behavior>)` |
| Startup Script | `scripts/<script-name>.sh` |
| Startup Command | `bash scripts/<script-name>.sh` |

`Functional Behavior Reference` must be exactly one Markdown link pointing to one subsection under `docs/product-documentation.md#4-functional-behavior`.

## 2. Scenario

- Subject under test: `<what exactly is being tested>`
- Expected result: `<single observable outcome>`

## 3. Preconditions

1. `<required environment/state>`
2. `<required data/configuration>`

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | `<action>` | `<observable outcome>` | `A1` |
| S2 | `<action>` | `<observable outcome>` | `A2` |
| S3 | `<action>` | `<observable outcome>` | `A3` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `<same reference as Metadata>` | `<clear verification rule>` | `<PASS/FAIL>` | `<what confirms result>` |
| A2 | `<same reference as Metadata>` | `<clear verification rule>` | `<PASS/FAIL>` | `<what confirms result>` |
| A3 | `<same reference as Metadata>` | `<clear verification rule>` | `<PASS/FAIL>` | `<what confirms result>` |

## 6. Final Result

- Test Result: `PASS` \| `FAIL`
- Failed Assertions: `<IDs or 'none'>`
- Failure Reason: `<required when result is FAIL>`
- Notes: `<optional>`

## 7. Cleanup

1. Exit app using `<key/command>`.
2. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`

# Informational Startup Flags Exit Without Runtime Launch

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-001` |
| Functional Behavior Reference | `[4.1 Database Configuration and Access](../docs/product-documentation.md#41-database-configuration-and-access)` |
| Startup Script | `scripts/start-informational.sh` |
| Startup Command | `bash scripts/start-informational.sh <help\|version>` |

## 2. Scenario

- Subject under test: startup informational aliases (`--help` and `--version`) executed through approved informational startup script binding.
- Expected result: informational command prints deterministic stdout and exits without opening selector or runtime.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-informational.sh` is executable in current environment.
3. Keep `TMP_ROOT` value from each script execution output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-informational.sh help`. | Help text is printed and process exits without opening selector/runtime UI. | `A1` |
| S2 | Run `bash scripts/start-informational.sh version`. | One version token is printed and process exits without opening selector/runtime UI. | `A2` |
| S3 | Inspect both command outputs in terminal. | Output shows deterministic informational behavior only (`help` usage text and `version` token). | `A3` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `[4.1 Database Configuration and Access](../docs/product-documentation.md#41-database-configuration-and-access)` | `help` output is rendered and startup short-circuits before selector/runtime launch. | `PASS` | Terminal prints help/usage content and returns to prompt immediately. |
| A2 | `[4.1 Database Configuration and Access](../docs/product-documentation.md#41-database-configuration-and-access)` | `version` output is one token (`short revision` or `dev`) and startup short-circuits before selector/runtime launch. | `PASS` | Terminal prints one version token and returns to prompt immediately. |
| A3 | `[4.1 Database Configuration and Access](../docs/product-documentation.md#41-database-configuration-and-access)` | Informational alias behavior remains deterministic and UI-free for both runs. | `PASS` | No selector/main-view UI is opened for either informational command. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Both informational paths are validated through the approved informational startup script binding.`

## 7. Cleanup

1. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT-from-help-run>`
2. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT-from-version-run>`

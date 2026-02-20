# Table Discovery and Schema View Stay Consistent

## 1. Metadata

| Field | Value |
| --- | --- |
| Case ID | `TC-003` |
| Functional Behavior Reference | `[4.3 Table Discovery and Schema View](../docs/product-documentation.md#43-table-discovery-and-schema-view)` |
| Startup Script | `scripts/start-direct-launch.sh` |
| Startup Command | `bash scripts/start-direct-launch.sh` |

## 2. Scenario

- Subject under test: table-list discovery behavior and schema rendering for selected table in the main runtime view.
- Expected result: table list remains alphabetically ordered without SQLite internal tables, and schema view renders selected-table column metadata.

## 3. Preconditions

1. Run from repository root.
2. Script `scripts/start-direct-launch.sh` is executable in current environment.
3. Keep `TMP_ROOT` value printed by startup script output for cleanup.

## 4. Test Steps

| Step ID | User Action | Expected Outcome | Assertion ID |
| --- | --- | --- | --- |
| S1 | Run `bash scripts/start-direct-launch.sh`. | Runtime opens with table list in left panel and schema view for selected table in right panel. | `A1` |
| S2 | Inspect visible table names in left panel. | SQLite internal tables are not listed. | `A2` |
| S3 | Verify order of visible table names. | Table list is alphabetically sorted. | `A3` |
| S4 | Move selection to a different table using `j`/`k`. | Right panel schema updates to show that table's column names and types. | `A4` |

## 5. Assertions

| Assertion ID | Functional Behavior Reference | Pass Criteria | Result (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- |
| A1 | `[4.3 Table Discovery and Schema View](../docs/product-documentation.md#43-table-discovery-and-schema-view)` | Runtime presents table-discovery list and schema panel for selected table. | `PASS` | Left panel contains table names and right panel shows schema rows immediately after startup. |
| A2 | `[4.3 Table Discovery and Schema View](../docs/product-documentation.md#43-table-discovery-and-schema-view)` | Table list excludes internal SQLite system tables. | `PASS` | No `sqlite_` system table names are shown in the list. |
| A3 | `[4.3 Table Discovery and Schema View](../docs/product-documentation.md#43-table-discovery-and-schema-view)` | Table list ordering is alphabetical for predictable scanning. | `PASS` | Visible table names are ordered lexicographically. |
| A4 | `[4.3 Table Discovery and Schema View](../docs/product-documentation.md#43-table-discovery-and-schema-view)` | Changing selected table refreshes schema with column name and type fields. | `PASS` | Right panel schema content updates after selection change and includes name/type information. |

## 6. Final Result

- Test Result: `PASS`
- Failed Assertions: `none`
- Failure Reason: `N/A`
- Notes: `Scenario is intentionally limited to table discovery and schema rendering ownership.`

## 7. Cleanup

1. Exit app using `q`.
2. Run:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`

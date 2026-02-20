# Suite Coverage Matrix

| Functional Behavior Reference | Scenario IDs | Assertion IDs | Mapping Completeness (`PASS`/`FAIL`) | Ownership / Purity (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- | --- |
| `[4.1 Database Configuration and Access](../docs/product-documentation.md#41-database-configuration-and-access)` | `TC-001` | `TC-001:A1,A2,A3` | `PASS` | `PASS` | `test-cases/TC-001-direct-launch-opens-main-view.md` |
| `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | `TC-002` | `TC-002:A1,A2,A3,A4` | `PASS` | `PASS` | `test-cases/TC-002-empty-config-startup-recovers-through-first-entry-setup.md` |
| `[4.3 Table Discovery and Schema View](../docs/product-documentation.md#43-table-discovery-and-schema-view)` | `TC-003` | `TC-003:A1,A2,A3,A4` | `PASS` | `PASS` | `test-cases/TC-003-selector-edit-invalid-path-blocks-save-until-corrected.md` |
| `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | `TC-004` | `TC-004:A1,A2,A3,A4,A5,A6,A7` | `PASS` | `PASS` | `test-cases/TC-004-runtime-command-failure-recovery-keeps-session-usable.md` |
| `[4.5 Filtering](../docs/product-documentation.md#45-filtering)` | `TC-007` | `TC-007:A1,A2,A3,A4,A5,A6,A7` | `PASS` | `PASS` | `test-cases/TC-007-filter-core-flow-applies-and-resets-on-table-switch.md` |
| `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | `TC-005` | `TC-005:A1,A2,A3,A4,A5,A6,A7,A8,A9,A10` | `PASS` | `PASS` | `test-cases/TC-005-save-failure-retains-staged-changes-until-corrected.md` |
| `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | `TC-006` | `TC-006:A1,A2,A3,A4,A5,A6,A7,A8,A9,A10,A11` | `PASS` | `PASS` | `test-cases/TC-006-dirty-config-navigation-requires-explicit-decision.md` |
| `[4.8 Visual State Communication](../docs/product-documentation.md#48-visual-state-communication)` | `TC-008` | `TC-008:A1,A2,A3,A4,A5,A6` | `PASS` | `PASS` | `test-cases/TC-008-visual-state-indicators-remain-visible-during-staged-changes.md` |

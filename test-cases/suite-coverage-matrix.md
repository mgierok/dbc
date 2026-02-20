# Suite Coverage Matrix

## Purpose

Define mandatory Functional Behavior mapping and traceability from Product Documentation references to scenarios and assertions.

## Coverage Review Rules

1. Coverage review is `FAIL` when a mapped reference is not a Markdown link to a subsection under `docs/product-documentation.md#4-functional-behavior`.
2. Coverage review is `FAIL` when any row has empty `Scenario IDs`.
3. Coverage review is `FAIL` when any row has empty `Assertion IDs`.
4. Coverage review is `FAIL` when a listed scenario ID does not resolve to an existing `test-cases/TC-*.md` file.
5. Mapping completeness is `PASS` only when every tracked Functional Behavior reference has non-empty scenario and assertion mappings.
6. Ownership/purity is `FAIL` when a scenario does not declare exactly one Functional Behavior reference or assertion rows contain mixed references.

## Functional Behavior Mapping Matrix

| Functional Behavior Reference | Scenario IDs | Assertion IDs | Mapping Completeness (`PASS`/`FAIL`) | Ownership / Purity (`PASS`/`FAIL`) | Evidence |
| --- | --- | --- | --- | --- | --- |
| `[4.1 Database Configuration and Access](../docs/product-documentation.md#41-database-configuration-and-access)` | `TC-001` | `TC-001:A1,A2,A3` | `PASS` | `PASS` | `test-cases/TC-001-direct-launch-opens-main-view.md` |
| `[4.2 Main Layout and Focus Model](../docs/product-documentation.md#42-main-layout-and-focus-model)` | `TC-002` | `TC-002:A1,A2,A3,A4` | `PASS` | `PASS` | `test-cases/TC-002-empty-config-startup-recovers-through-first-entry-setup.md` |
| `[4.3 Table Discovery and Schema View](../docs/product-documentation.md#43-table-discovery-and-schema-view)` | `TC-003` | `TC-003:A1,A2,A3,A4` | `PASS` | `PASS` | `test-cases/TC-003-selector-edit-invalid-path-blocks-save-until-corrected.md` |
| `[4.4 Records View and Navigation](../docs/product-documentation.md#44-records-view-and-navigation)` | `TC-004` | `TC-004:A1,A2,A3,A4,A5,A6,A7` | `PASS` | `PASS` | `test-cases/TC-004-runtime-command-failure-recovery-keeps-session-usable.md` |
| `[4.5 Filtering](../docs/product-documentation.md#45-filtering)` | `none` | `none` | `FAIL` | `FAIL` | `Area coverage intentionally pending TASK-04 scenario creation.` |
| `[4.6 Data Operations (Insert, Edit, Delete)](../docs/product-documentation.md#46-data-operations-insert-edit-delete)` | `TC-005, TC-006` | `TC-005:A4,A7; TC-006:A3,A7` | `PASS` | `FAIL` | `test-cases/TC-005-save-failure-retains-staged-changes-until-corrected.md`, `test-cases/TC-006-dirty-config-navigation-requires-explicit-decision.md` |
| `[4.7 Staging, Undo/Redo, and Save](../docs/product-documentation.md#47-staging-undoredo-and-save)` | `TC-005, TC-006` | `TC-005:A4,A5,A6,A7,A8,A9; TC-006:A3,A4,A5,A7,A8` | `PASS` | `FAIL` | `test-cases/TC-005-save-failure-retains-staged-changes-until-corrected.md`, `test-cases/TC-006-dirty-config-navigation-requires-explicit-decision.md` |
| `[4.8 Visual State Communication](../docs/product-documentation.md#48-visual-state-communication)` | `TC-005, TC-006` | `TC-005:A6,A9; TC-006:A3,A4,A7,A8` | `PASS` | `FAIL` | `test-cases/TC-005-save-failure-retains-staged-changes-until-corrected.md`, `test-cases/TC-006-dirty-config-navigation-requires-explicit-decision.md` |

## Baseline Conclusion

- Mapping completeness (areas `4.1` to `4.4`): `PASS` (`4/4` scoped references mapped with scenario and assertion IDs).
- Ownership/purity (areas `4.1` to `4.4`): `PASS` (`4/4` scoped references mapped to area-pure scenarios).
- Mapping completeness (areas `4.1` to `4.8`): `FAIL` (`7/8`; area `4.5` intentionally pending TASK-04).
- Ownership/purity (areas `4.1` to `4.8`): `FAIL` (`TC-005` and `TC-006` remain pre-refactor and include mixed-area assertions).

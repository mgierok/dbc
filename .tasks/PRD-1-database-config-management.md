# Overview

This PRD defines an in-app database configuration management capability within the database selection experience, including add, edit, and delete actions, forced first-database setup, cross-platform config path handling, and in-session return to selector/management.

We believe that enabling in-app management of configured databases for individual developers using DBC will reduce startup configuration failures and speed up first successful data access.
We will know this is true when configuration-related startup failures drop from 18% to 5% or lower within 8 weeks after release.

# Problem Statement

Users currently depend on manual configuration-file editing to maintain database entries. This creates friction, especially on first run and across different operating systems where default config locations differ. The lack of in-app management increases startup failure risk, slows onboarding, and forces context switching to external editors. Additionally, users working inside a selected database cannot directly return to the selector/management context through an in-app command workflow.

# Current State (As-Is)

- The product reads configured databases from a configuration file and requires at least one valid entry.
- Database selector behavior is limited to choosing among already defined entries.
- Add, delete, and edit of configured databases are not available in the selector flow.
- On installations without a valid configured database, users cannot complete a guided in-app first-setup flow.
- Configuration path expectations are not presented as a cross-platform, user-visible product behavior.
- While already inside an opened database session, users do not have a command-based in-app return path dedicated to selector/configuration management.

# Target State After Release (To-Be)

- The selector screen also supports configuration management for database entries: add, edit, delete.
- If no configured databases exist, the app enforces a first-database creation flow before normal browsing can start.
- Users can optionally add additional databases during the same setup context.
- The app follows OS-native default config file location rules on Windows, macOS, and Linux, and always shows the active config file path in the management UI.
- From inside an active database session, users can run `:config` to open selector/management.
- If staged data changes exist when invoking `:config`, the product requires explicit decision: save, discard, or cancel navigation.

# Business Rationale and Strategic Fit

- Improves first-run success and reduces early abandonment caused by configuration friction.
- Aligns with keyboard-first, vim-like interaction principles by introducing a command-driven in-session return (`:config`).
- Reduces support burden from cross-platform path confusion by making config location behavior explicit and visible.
- Strengthens product usability for local developer workflows, the primary segment.

# Goals

- G1. Decrease configuration-related startup failures for first-run and reconfiguration scenarios.
- G2. Reduce time required for new users to reach a successful first database selection.
- G3. Enable complete in-app lifecycle management of configured databases without external file editing.
- G4. Preserve data safety when switching to configuration management from an active session with unsaved staged changes.

# Non-Goals

- Support for non-SQLite engines in this release.
- Schema management or database administration features.
- Background sync or cloud distribution of configuration data.
- Role-based access control for configuration operations.
- Redesign of unrelated browsing, filtering, or table-edit workflows.

# Scope (In Scope / Out of Scope)

In Scope:
- In-selector actions to add, edit, and delete database entries stored in config.
- Mandatory guided first-entry creation when no database entries exist.
- Optional addition of further entries in the same setup context.
- Display of active configuration file path in management UI.
- OS-native default configuration location behavior for Windows, macOS, Linux.
- Command `:config` to return from active session to selector/management.
- Explicit safety prompt (save/discard/cancel) when unsaved staged changes exist before navigation to `:config`.

Out of Scope:
- Multi-engine runtime behavior beyond current SQLite scope.
- Advanced metadata for database entries (tags, grouping, favorites, environment labels).
- Automated migration/import of external configuration formats.
- Multiple concurrent profile files and profile-switching strategy.

# Functional Requirements

FR-001: The selector screen must provide a user-visible action to add a new configured database entry with required fields `name` and `db_path`.
Acceptance: A user can create a valid entry entirely in-app and sees it immediately in selector list without external file editing.

FR-002: The selector screen must provide a user-visible action to edit existing configured database entry fields `name` and `db_path`.
Acceptance: A user can update an existing entry and the selector reflects the updated values in the same session.

FR-003: The selector screen must provide a user-visible action to delete an existing configured database entry.
Acceptance: A deleted entry is removed from selector list and no longer available for selection after confirmation.

FR-004: On startup with zero configured database entries, the product must block normal browsing flow and enforce first-entry creation.
Acceptance: A user cannot enter main browsing UI until at least one valid entry has been created.

FR-005: During forced first-entry creation, the product must allow optional creation of additional entries before entering main browsing UI.
Acceptance: A user can add one required first entry and optionally continue adding more entries in the same flow.

FR-006: The product must support command `:config` from an active database session to open selector/management.
Acceptance: Entering `:config` from active session always navigates to selector/management context.

FR-007: If unsaved staged data changes exist when `:config` is triggered, the product must require explicit decision: save, discard, or cancel.
Acceptance: Navigation to selector/management does not proceed until one explicit option is chosen.

FR-008: The product must apply OS-native default config location behavior on Windows, macOS, and Linux where the app runs.
Acceptance: On each supported OS, first-run configuration behavior uses that OS default location rule without requiring manual path input.

FR-009: The selector/management UI must display the currently active configuration file path.
Acceptance: A user can view the exact active config path directly in-app before or during entry management.

FR-010: Add, edit, and delete actions must persist changes to configuration data so that subsequent app startups use the latest saved entries.
Acceptance: After app restart, previously saved entry changes remain consistent with the last in-app management action.

# Non-Functional Product Requirements

NFR-001: The management flow must remain keyboard-first and consistent with existing interaction language used by the product.

NFR-002: Critical destructive actions in configuration management (for example delete) must require explicit confirmation.

NFR-003: Navigation and prompts related to `:config` must be understandable to first-time users without external documentation.

NFR-004: Cross-platform behavior must be predictable: users can identify where configuration is read from regardless of OS.

NFR-005: Configuration changes must be reliable for end users, with no silent loss of successfully confirmed add/edit/delete actions.

# Success Metrics and Release Criteria

Primary Outcome Metric:
- M1 Startup configuration failure rate.
Baseline: 18% of first-run/reconfiguration sessions fail due to missing or invalid config handling.
Target: 5% or lower.
Measurement window: First 8 weeks after release.
Measurement method: Tagged startup outcome logs from QA and production-like acceptance runs, classified as success/failure for config-related startup path.
Related goals: G1.

Leading Indicators:
- M2 First-run setup completion rate (from app launch to successful database selection).
Baseline: 62%.
Target: 90% or higher.
Measurement window: First 8 weeks after release.
Measurement method: Structured usability-session tracking and QA matrix runs across Windows/macOS/Linux.
Related goals: G1, G2.

- M3 Median time to first successful database selection for new users.
Baseline: 4 minutes 30 seconds.
Target: 1 minute 30 seconds or less.
Measurement window: First 8 weeks after release.
Measurement method: Time-based observation in scripted onboarding tests for new-user scenarios.
Related goals: G2, G3.

- M4 In-session selector return success rate via `:config`.
Baseline: 0% (feature unavailable).
Target: 95% successful completion when command is invoked.
Measurement window: First 8 weeks after release.
Measurement method: Command invocation outcome tracking in acceptance tests and usability sessions.
Related goals: G3.

Guardrail Metric:
- M5 Unsaved staged-change loss incidents triggered by `:config` navigation.
Baseline: 0 known incidents (feature unavailable; baseline set from existing data-loss incident log category).
Target: 0 incidents.
Measurement window: First 8 weeks after release.
Measurement method: Incident log review and regression test evidence for save/discard/cancel decision path.
Related goals: G4.

Release Criteria:
- Release proceeds only if M1 and M2 targets are met in acceptance testing on Windows, macOS, and Linux.
- M5 must remain at target (0 incidents) throughout release qualification.
- FR-001 through FR-010 acceptance statements must be demonstrably satisfied.

# Risks and Dependencies

Risks:
- Users may misinterpret delete behavior and remove needed entries accidentally.
- Cross-platform path expectations may still be confusing if path visibility is not prominent enough.
- Frequent switching to config context could interrupt workflows if prompts are too intrusive.
- Metrics may be hard to compare if startup outcome tagging is inconsistently applied.

Dependencies:
- Clear product copy for management actions and confirmations.
- Cross-platform validation coverage on Windows, macOS, and Linux environments where DBC runs.
- Agreement on startup outcome taxonomy for consistent metric measurement.
- Existing keyboard interaction conventions must be preserved in new management flow.

# Assumptions

- A1 (High): Primary target users for this scope are individual developers using DBC locally.
- A2 (High): The command `:config` is accepted as the canonical in-session entry point to selector/management.
- A3 (Medium): Baseline metric values (M1-M3) are based on current internal QA/usability observations and can be measured consistently post-release.
- A4 (High): Advanced metadata management for configured databases is intentionally excluded from this release.
- A5 (Medium): All three target OS families are available in release validation where DBC runtime is supported.

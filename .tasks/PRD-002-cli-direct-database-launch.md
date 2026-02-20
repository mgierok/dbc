# Overview

This PRD defines a direct CLI launch mode where DBC can start with a database connection string parameter, bypass the startup selector on success, and fail fast with a clear message on connection failure.

We believe that enabling direct database launch from the command line for DBC users will reduce startup friction and time-to-first-usable-session while preserving safe access to existing selector-based database switching.
We will know this is true when median startup-to-main-view time improves to the defined target within 6 weeks after release.

# Metadata

- Status: READY

# Problem Statement

Current startup always routes through the database selector, even when a user already knows the target database and wants immediate entry. This adds avoidable startup steps for terminal-centric workflows. Users need a fast launch path that still validates connectivity before runtime and still preserves in-session ability to switch databases safely.

# Current State (As-Is)

- Startup flow is selector-first for all sessions.
- User must select a configured database before entering the main view.
- Connection failures for selected entries are handled inside selector status feedback.
- In-session database switching is available through `:config`.
- Databases available for switching are sourced from config-managed entries.

# Target State After Release (To-Be)

- User can launch DBC with a direct database connection-string parameter (with supported aliases).
- If the provided target is reachable, DBC opens directly into the selected database context, skipping selector UI.
- If direct-launch validation fails, DBC prints a readable error and exits with non-zero code.
- While inside that session, `:config` remains available and allows switching to other configured databases.
- The CLI-provided database appears in selector options for that running process so the user can return to it, but it is not auto-persisted to config.
- If the same connection string already exists in config, existing configured entry is reused (no duplicate temporary entry).

# Business Rationale and Strategic Fit

- Improves startup speed for terminal-first workflows.
- Reduces unnecessary navigation for known-target sessions.
- Preserves existing safe switching behavior and config management model.
- Strengthens product usability parity between scripted/CLI workflows and interactive selector workflows.

# Goals

- G1. Reduce time from process start to usable main view for known-target launches.
- G2. Preserve clear and safe failure behavior for invalid/unreachable direct-launch targets.
- G3. Maintain continuity of existing in-session database switching behavior.
- G4. Prevent unintended persistence side effects from ad-hoc CLI launch targets.

# Non-Goals

- Adding support for non-SQLite engines.
- Replacing or removing selector-first startup behavior for users who do not pass a direct-launch parameter.
- Introducing automatic config writes for CLI-provided databases.
- Redesigning unrelated table browsing, filtering, editing, or save flows.

# Scope (In Scope / Out of Scope)

In Scope:
- Direct launch path using a connection-string CLI parameter with multiple aliases.
- Startup pre-validation of direct-launch connection before main runtime opens.
- Readable error output and non-zero process exit on direct-launch validation failure.
- Selector bypass only for successful direct-launch attempts.
- In-session `:config` access and switching behavior preserved.
- Process-lifetime availability of CLI-provided target in selector list.
- Duplicate suppression by reusing existing configured entry when connection string already matches.

Out of Scope:
- Persisting temporary CLI entries to configuration file by default.
- Bulk import of CLI-provided targets into config.
- New configuration schema fields for source tagging.
- Any change to current selector CRUD semantics beyond temporary list visibility behavior.

# Functional Requirements

FR-001: The product must accept a direct-launch database connection-string parameter using supported aliases.
Acceptance: User can provide the parameter at startup and DBC recognizes it as a direct-launch request.

FR-002: On direct-launch request, the product must validate database connectivity before opening the main view.
Acceptance: Main view opens only after successful connectivity validation for the provided target.

FR-003: On direct-launch validation failure, the product must show a clear user-facing error and terminate startup.
Acceptance: Failure output includes actionable context and the process exits with non-zero status.

FR-004: Successful direct-launch must bypass the startup selector and enter the database session directly.
Acceptance: User reaches main view without interacting with selector when connection is valid.

FR-005: During a direct-launched session, `:config` must continue to provide access to selector and configured database switching.
Acceptance: User can invoke `:config` and select another available database without restarting the process.

FR-006: The direct-launched database must be available in selector options during the same process.
Acceptance: After entering `:config`, user can reselect the direct-launched target from the list in that running session.

FR-007: Direct-launched temporary database entries must not be automatically persisted to config.
Acceptance: After app restart, temporary entries from prior direct launches are absent unless explicitly saved via existing config management flows.

FR-008: If direct-launch connection string matches an existing configured database, product must reuse that existing entry identity.
Acceptance: Selector list does not show a duplicate temporary entry for an already configured target.

# Non-Functional Product Requirements

NFR-001: Startup behavior messaging for direct-launch success/failure must be understandable without external documentation.

NFR-002: Failed direct-launch must fail fast and not leave the user in an ambiguous startup state.

NFR-003: Direct-launch mode must preserve consistency of current navigation and safety expectations in session.

NFR-004: Temporary-entry behavior must be predictable and transparent across a single process lifecycle.

NFR-005: CLI and selector startup paths must remain behaviorally coherent for users switching between workflows.

# Success Metrics and Release Criteria

Primary Outcome Metric:
- M1 Median startup-to-main-view time for valid direct-launch sessions.
Baseline: 4.5 seconds.
Target: 2.0 seconds or less.
Measurement window: First 6 weeks after release.
Measurement method: Instrumented startup timing in QA acceptance runs and sampled real-user telemetry sessions.

Leading Indicators:
- M2 Direct-launch successful-start rate (valid parameter path).
Baseline: 0% (feature unavailable).
Target: 98% or higher.
Measurement window: First 6 weeks after release.
Measurement method: Startup outcome logs for sessions with direct-launch parameter.

- M3 Share of known-target sessions completed without selector interaction.
Baseline: 0% (feature unavailable).
Target: 70% or higher.
Measurement window: First 6 weeks after release.
Measurement method: Session classification from startup-path analytics (direct launch vs selector-first).

- M4 Error clarity satisfaction in failed direct-launch usability checks.
Baseline: Not measured.
Target: 90% of participants can correctly identify next action after reading error.
Measurement window: Release qualification period.
Measurement method: Structured usability test script with comprehension scoring.

Guardrail Metric:
- M5 Unintended config persistence incidents caused by direct-launch temporary entries.
Baseline: 0 incidents.
Target: 0 incidents.
Measurement window: First 6 weeks after release.
Measurement method: Regression test evidence plus config-diff audit in acceptance scenarios.

Release Criteria:
- M1 and M2 targets must be met in release qualification.
- M5 must remain at target (0 incidents).
- FR-001 through FR-008 acceptance statements must be satisfied in acceptance testing.

# Risks and Dependencies

Risks:
- Users may misread direct-launch errors if wording is too technical.
- Alias behavior may create confusion if not clearly documented in CLI help.
- Temporary list visibility may be misunderstood as persistence.
- Startup-path metrics may be noisy if event tagging is inconsistent.

Dependencies:
- Clear product copy for failure output and help text.
- Reliable startup-path measurement instrumentation.
- Regression coverage for selector behavior alongside direct-launch path.
- Alignment with existing configuration management user expectations.

# State & Failure Matrix

| Flow Area | Trigger / Failure Mode | Expected Product Response | User-Visible Recovery Path |
| --- | --- | --- | --- |
| startup | User provides valid direct-launch connection string | Skip selector and open main view directly | Continue normal work; use `:config` anytime to switch |
| startup | User provides invalid/unreachable direct-launch connection string | Show clear error and terminate with non-zero exit | Correct parameter/target and relaunch |
| config | User invokes `:config` during direct-launched session | Open selector with configured entries plus temporary direct-launch entry for current process | Select any available database, including returning to direct-launch target |
| config | Direct-launch target already exists in config | Reuse existing configured entry, no temporary duplicate | User sees single canonical entry and can select it normally |
| save | User tries to switch via `:config` with staged changes | Preserve existing dirty-state decision behavior (save/discard/cancel) | Choose save/discard/cancel and proceed per current workflow |
| navigation | User exits app and relaunches without explicit save of temporary entry | Temporary direct-launch entry no longer present in selector list | Re-pass CLI parameter or explicitly add database through config management |

# Assumptions

- A1 (High): Supported direct-launch aliases will be defined as one short and one long form in implementation-facing specs.
- A2 (Medium): Baseline timing and usability values are validated with release-qualification data before GA.
- A3 (High): Existing selector CRUD remains the only persistence path for database entries.
- A4 (High): Direct-launch remains optional and does not deprecate selector-first startup.

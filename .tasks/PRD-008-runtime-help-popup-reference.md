# Overview

We believe that adding an in-session `:help` popup reference for terminal-first DBC users will reduce command and shortcut discovery friction and improve first-try command success during active sessions.
We will know this is true when the updated existing help-validation test case passes at 100% in release validation.

# Metadata

- Status: READY

# Problem Statement

Users currently have no in-session command to quickly discover supported runtime commands and keyboard keywords. This creates avoidable memory burden, slows workflows, and increases command-entry errors while working in an active database session.

# Current State (As-Is)

- Runtime command entry exists during active sessions.
- `:config` is supported as an in-session command.
- Unsupported commands return an unknown-command status message.
- No runtime `:help` command exists.
- No popup lists supported runtime commands and keyboard keywords in one place.

# Target State After Release (To-Be)

- Users can execute `:help` from runtime command entry in main session contexts.
- A popup opens with two sections: `Supported Commands` and `Supported Keywords`.
- Each listed command/keyword has a short one-line description.
- If content exceeds popup height, users can scroll to reach all items.
- Executing `:help` while popup is already open keeps popup open (idempotent open behavior).
- Popup closes with `Esc` only.

# Business Rationale and Strategic Fit

This change improves in-session self-service discoverability for keyboard-first workflows. It reduces context switching to external references, lowers usage friction for less-frequent commands and shortcuts, and supports faster onboarding without broadening product scope.

# Goals

- G1: Enable runtime command discoverability directly inside an active session.
- G2: Enable runtime keyboard-keyword discoverability directly inside an active session.
- G3: Provide deterministic release readiness evidence through test validation of `:help` behavior.

# Non-Goals

- NG1: Expanding startup CLI `--help`/`-h` behavior.
- NG2: Adding new runtime command families beyond `:help`.
- NG3: Enabling `:help` in selector/startup contexts.
- NG4: Redesigning unrelated popup workflows.

# Scope (In Scope / Out of Scope)

In Scope:
- Add runtime `:help` command support in main active-session contexts.
- Show popup content split into `Supported Commands` and `Supported Keywords`.
- Show one-line descriptions for all listed entries.
- Support keyboard scrolling when popup content overflows visible space.
- Update one existing test case to validate this feature and use PASS result as release evidence.

Out of Scope:
- Startup informational help/version behavior.
- Selector/form context support for `:help`.
- Adding unrelated commands or altering existing command semantics outside help discoverability.
- UI redesign beyond needed popup behavior/content for this feature.

# Functional Requirements

FR-001:
- The product must accept `:help` in runtime command entry for active main-session contexts.
- Acceptance: Entering `:help` opens the help popup and does not show unknown-command status.

FR-002:
- The popup must contain a `Supported Commands` section.
- Acceptance: The section is visible and lists supported runtime commands with one-line descriptions.

FR-003:
- The popup must contain a `Supported Keywords` section, where keywords are keyboard shortcuts/key bindings.
- Acceptance: The section is visible and lists supported keyboard keywords with one-line descriptions.

FR-004:
- If popup content exceeds available height, users must be able to scroll within the popup.
- Acceptance: Users can navigate to the final listed item via keyboard scrolling.

FR-005:
- Running `:help` while the help popup is already open must be idempotent.
- Acceptance: Re-entering `:help` keeps the popup open and does not dismiss it.

FR-006:
- The help popup must close only on `Esc`.
- Acceptance: `Esc` closes popup; popup is not dismissed by unrelated keys.

FR-007:
- Existing behavior for unsupported runtime commands must remain intact.
- Acceptance: Unsupported commands still return unknown-command status and keep session usable.

# Non-Functional Product Requirements

NFR-001:
- Help popup interaction must remain keyboard-first and consistent with current terminal workflow.

NFR-002:
- Popup content must remain scannable, with clear section separation and concise one-line descriptions.

NFR-003:
- Feature release evidence must be deterministic and reproducible using the updated existing test case result.

# Success Metrics and Release Criteria

Primary Outcome Metric:
- M1: Updated existing test case pass rate for runtime `:help` popup behavior.
  - Baseline: 0% (no current existing test case validates runtime `:help` popup behavior).
  - Target: 100% pass.
  - Measurement window: Release validation execution for this feature.
  - Measurement method: Test execution artifact showing PASS for the updated existing test case that validates `:help` popup behavior.

Leading Indicators:
- M2: Functional requirement validation coverage for `FR-001` to `FR-007`.
  - Baseline: 0/7 requirements mapped to `:help`-specific validation.
  - Target: 7/7 requirements mapped and validated.
  - Measurement window: PR/release readiness review.
  - Measurement method: Validation checklist artifact mapping each FR to concrete test assertions with PASS outcomes.
- M3: Required popup section completeness.
  - Baseline: 0/2 required sections present in runtime help popup.
  - Target: 2/2 required sections present with one-line descriptions.
  - Measurement window: Feature validation execution.
  - Measurement method: Updated existing test case assertions confirming both required sections and description presence.

Guardrail Metric:
- M4: Regressions in unsupported-command handling behavior.
  - Baseline: 0 accepted regressions.
  - Target: 0 regressions.
  - Measurement window: Same release validation execution.
  - Measurement method: Regression assertions in the updated existing test case confirm unsupported command behavior remains unchanged.

Release Criteria:
- Updated existing test case validating runtime `:help` popup behavior is PASS.
- All `FR-001` to `FR-007` validations are PASS.
- Guardrail metric `M4` remains at target.

# Risks and Dependencies

Risks:
- Incomplete or inconsistent keyword listing can reduce discoverability value.
- Popup content growth can reduce readability if descriptions are not kept concise.
- Adding `:help` can unintentionally affect existing command-entry behavior if regression checks are weak.

Dependencies:
- Runtime command entry remains available in active main-session contexts.
- Existing test-case suite supports updating one current scenario for `:help` validation.
- Interaction model remains the source of truth for supported keyboard keywords.

# State & Failure Matrix

| Flow Area | Trigger / Failure Mode | Expected Product Response | User-Visible Recovery Path |
| --- | --- | --- | --- |
| startup | Out of scope for this feature (startup help/version flows) | Startup behavior remains unchanged | Use existing startup workflow and current startup help/version commands |
| config | User enters `:config` instead of `:help` during runtime | Existing `:config` behavior remains unchanged | Reopen command entry and run `:help` in runtime context |
| save | User attempts save workflow while help popup is open | Existing save behavior is preserved; help popup does not redefine save semantics | Close popup with `Esc` and continue normal save flow |
| navigation | Help popup content exceeds visible height | Popup allows keyboard scrolling to reach hidden entries | Scroll within popup to target item, then close with `Esc` |

# Assumptions

- A1 (High): Help popup copy is in English to match current internal product documentation language.
- A2 (High): `Supported Commands` includes at least `:config` and `:help` in v1.
- A3 (Medium): `Supported Keywords` list is based on currently documented keyboard shortcuts.
- A4 (Medium): One existing runtime-oriented test case can be extended to cover `:help` popup behavior and regression guardrails.

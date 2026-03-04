---
name: unit-test-audit
description: Audit unit tests for a full repository or a user-specified directory and produce actionable findings for stale or redundant tests plus missing coverage. Use only when the user explicitly asks for a unit-test audit, stale-test review, test-pruning review, or unit-test coverage-gap analysis. This skill MUST NOT auto-trigger in standard implementation workflows.
---

# Unit Test Audit

## Purpose

This skill performs a read-only, periodic audit of unit tests to:
- identify redundant, unnecessary, or outdated unit tests with concrete removal recommendations,
- identify missing unit-test coverage with concrete improvement guidance.

The skill is language agnostic and project agnostic. It MUST treat project-defined testing rules and conventions as the source of truth.

## Scope

- In scope:
  - audit all unit tests in repository scope or a user-selected subdirectory,
  - classify potential removal candidates with evidence and risk,
  - classify coverage gaps and recommend targeted additions.
- Out of scope:
  - modifying production code or tests,
  - deleting tests automatically,
  - enforcing one universal testing style across all projects.

## Explicit Invocation Policy

- This skill MUST run only on explicit user request.
- This skill MUST NOT be auto-invoked from standard coding or review workflows.
- If invocation intent is not explicit, the skill MUST ask the user to confirm whether the audit should run.
- The skill MUST use `update_plan` from workflow start through completion, updating statuses after each workflow step and finishing with all steps marked as `completed`.

## Inputs

- `scope`:
  - optional path to analyze,
  - default is repository root when no path is provided.
- `project testing rules`:
  - documentation, conventions, and local testing policies available in the current project.

## Applicability Rule

- The skill MUST evaluate each mandatory audit area defined in this file and decide whether it is applicable to the current project and selected scope.
- If an area is not applicable, the report MUST mark it as `Not applicable` and MUST include a brief reason.
- If an area is applicable, the report MUST include evidence-based findings and actionable recommendations.

## Audit Workflow

1. Confirm scope and inventory files
- Resolve the audit scope (`repo root` or provided directory).
- Build a list of candidate unit-test files and relevant production files.

2. Load project testing expectations
- Find project-specific test rules before applying generic heuristics.
- If project rules are missing or ambiguous, state that explicitly and proceed with heuristic fallback.

3. Build behavior-to-test map
- Map test cases to covered behavior contracts and production modules.
- Identify duplicated behavior checks and missing behavior checks.

4. Identify removal candidates (balanced policy)
- Mark findings with confidence level and risk level.
- Use two recommendation classes:
  - `HIGH_CONFIDENCE_REMOVE`: strong evidence test is redundant or stale.
  - `REVIEW_REQUIRED`: potential removal candidate that needs human validation.
- Every candidate MUST include:
  - file path and test identifier,
  - reason (`redundant`, `obsolete`, `misaligned with current contract`, or similar),
  - concrete evidence,
  - risk note,
  - suggested validation step before deletion.

5. Identify coverage gaps
- Locate untested or under-tested behavior based on project-defined expectations.
- For each gap, provide:
  - affected module or contract,
  - missing scenario type (`happy path`, `edge`, `error`, `state transition`, or equivalent),
  - why the gap matters,
  - concrete recommendation for test additions.

6. Prioritize recommendations
- Group recommendations into:
  - `High impact`,
  - `Medium impact`,
  - `Low impact`.
- Prioritization SHOULD consider defect risk, maintainability gain, and effort.

7. Produce final report
- Generate a Markdown report using `references/report-template.md`.
- Save the generated report to `unit-test-audit.md` in the project root directory.
- Follow validation points from `references/analysis-checklist.md`.

## Decision Rules

- The skill MUST prefer project-native rules over generic assumptions.
- The skill SHOULD avoid naming a test as removable when confidence is low.
- The skill MUST separate evidence from assumptions.
- The skill MUST flag uncertainty and request manual confirmation where needed.
- The skill MUST keep recommendations actionable and file-specific.

## Mandatory Audit Areas (Apply When Applicable)

1. Unit Definition and Boundaries
- The skill MUST verify whether tests target one unit (for example function, class, or use case) in isolation.
- The skill MUST verify that tests do not touch DB, file system, network, or real clock without stubs or test doubles.
- The skill MUST verify that boundaries between `unit`, `integration`, `contract`, and `e2e` tests are explicit (for example by tags, naming, or directory structure).

2. Logic and Risk Coverage (Not Only Percentage Coverage)
- The skill MUST verify coverage of edge cases, error paths, exceptions, validation, permissions logic (when implemented), and retry/timeout behavior (when implemented).
- The skill SHOULD verify the presence of regression tests for known bugs.
- The skill MUST verify that critical business rules include both happy-path and unhappy-path tests.

3. Assertion Quality and Test Value
- The skill MUST verify that assertions are specific and behavior-focused, not limited to weak checks such as only "no exception".
- The skill SHOULD flag brittle assertions (for example full-object comparisons or low-value snapshots without clear intent).
- The skill SHOULD verify behavior-level checks instead of implementation-detail checks where feasible.

4. Isolation and Dependency Handling (Mocking/Stubbing)
- The skill MUST verify that external dependencies are mocked or stubbed consistently.
- The skill SHOULD flag mock overuse where tests mostly validate mock configuration instead of business logic.
- Interaction verification (for example call counts) SHOULD be used only where behavior requires it (for example side effects).

5. Determinism and Stability (Flaky Tests)
- The skill MUST evaluate test resilience to time dependencies (timezone, DST, real `now`), randomness, execution order, parallel execution, and global/singleton state.
- The skill SHOULD assess available flaky-test reporting or rerun mechanisms.
- The skill MUST flag cases where flakes appear masked instead of removed.

6. Test Data and Readability
- The skill SHOULD verify that test data is minimal, readable, and local to each test where practical.
- The skill SHOULD assess factories/builders/fixtures for hidden "magic defaults" that obscure test intent.
- The skill MUST verify that test names communicate condition and expected result.

7. Test Structure and Maintainability
- The skill SHOULD verify consistent test style (for example Arrange-Act-Assert or Given-When-Then).
- The skill SHOULD identify excessive setup duplication and recommend reductions that preserve readability.
- The skill SHOULD verify that test organization mirrors production structure so tests are easy to locate.

8. Execution Speed and Cost
- The skill SHOULD assess unit-suite runtime and identify non-linear growth risk where measurable.
- The skill MUST flag heavy initialization in unit tests (for example booting full frameworks) as a likely boundary violation.
- The skill SHOULD verify that running a single unit test is straightforward in common IDE/CLI workflows.

9. Mutation Resistance and False-Green Risk
- The skill SHOULD evaluate whether tests detect meaningful code changes instead of passing on ineffective assertions.
- If mutation testing is available, the skill SHOULD review mutation results for critical modules and flag materially low mutation scores.

10. Contract and Compatibility at Unit Boundaries
- The skill SHOULD verify boundary-unit tests (for example serialization, mapping, validators) for null/empty inputs and invalid formats.
- If backward compatibility applies, the skill SHOULD verify coverage for backward-compatibility behavior.

11. Anti-Patterns to Detect
- The skill MUST flag order-dependent tests.
- The skill MUST flag use of `sleep()` as synchronization when stronger synchronization is feasible.
- The skill MUST flag mocking of the subject under test (including risky half-mocks).
- The skill SHOULD flag copy-paste tests with weak or duplicated assertions.
- The skill MUST flag tests that still pass after meaningful code removal (false-green indicators).

12. CI Tooling and Reporting
- The skill SHOULD assess whether CI publishes unit-test results, flaky-test indicators, coverage by module, and runtime trends.
- The skill SHOULD assess whether quality gates are practical and risk-based (not blindly set to 100%).
- The skill MUST flag misleading coverage setups (for example excludes that hide meaningful code without justification).

## Output Requirements

The final output MUST be a Markdown report with these sections:
- `Scope and Inputs`
- `Applicability Matrix`
- `Removal Candidates`
- `Coverage Gaps`
- `Recommended Actions`
- `Confidence and Risk`
- `Evidence and Limitations`

The report MUST include concrete file references and test identifiers wherever possible.
The report MUST include each mandatory audit area with status (`Applicable` or `Not applicable`) and supporting evidence.
The skill MUST write the report file to `<project-root>/unit-test-audit.md`.

## Safety and Boundaries

- This skill MUST remain analysis-only for source code and tests.
- This skill MUST NOT edit repository files except writing `<project-root>/unit-test-audit.md`.
- This skill MUST NOT infer product behavior from tests alone when code and docs conflict; current code behavior is factual unless project governance states otherwise.
- This skill MAY provide optional follow-up commands for verification, but MUST label them as optional.

## References

- `references/analysis-checklist.md`
- `references/report-template.md`

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
- Return a Markdown report using `references/report-template.md`.
- Follow validation points from `references/analysis-checklist.md`.

## Decision Rules

- The skill MUST prefer project-native rules over generic assumptions.
- The skill SHOULD avoid naming a test as removable when confidence is low.
- The skill MUST separate evidence from assumptions.
- The skill MUST flag uncertainty and request manual confirmation where needed.
- The skill MUST keep recommendations actionable and file-specific.

## Output Requirements

The final output MUST be a Markdown report with these sections:
- `Scope and Inputs`
- `Removal Candidates`
- `Coverage Gaps`
- `Recommended Actions`
- `Confidence and Risk`
- `Evidence and Limitations`

The report MUST include concrete file references and test identifiers wherever possible.

## Safety and Boundaries

- This skill MUST remain analysis-only and MUST NOT edit repository files.
- This skill MUST NOT infer product behavior from tests alone when code and docs conflict; current code behavior is factual unless project governance states otherwise.
- This skill MAY provide optional follow-up commands for verification, but MUST label them as optional.

## References

- `references/analysis-checklist.md`
- `references/report-template.md`

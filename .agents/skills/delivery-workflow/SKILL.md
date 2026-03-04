---
name: delivery-workflow
description: Execute the standard four-phase delivery workflow for project-code changes. Use only when the user explicitly requests this skill by name; this skill MUST NOT auto-trigger.
---

# Delivery Workflow

## Purpose

Provide the standard workflow for intent alignment, planning, implementation, and completion of project-code changes.

## Scope

- This skill MUST be applied only to project tasks that can result in project-code changes.
- This skill MUST NOT be applied to documentation-only or governance-only tasks.

## Trigger

- This skill MUST run only when the user explicitly requests `delivery-workflow` by name.
- This skill MUST NOT auto-trigger from inferred intent.

## Workflow

The workflow MUST be executed in four ordered stages.
The workflow MUST NOT use one global `update_plan` that spans multiple stages.
Each stage MUST use its own stage-local `update_plan`.
Before starting the next stage, the agent MUST close the current stage-local plan with all steps marked as `completed`.
At any point in time, the agent MUST keep at most one active stage-local `update_plan`.

1. Intent Alignment
2. Planning
3. Implementation
4. Completion

### Change-Set Definition

- A change-set MUST be the smallest independently reviewable implementation increment that delivers one coherent value objective and working software.
- Each change-set MUST have a unique identifier within the full plan.
- A change-set MUST be executable, verifiable, and reversible as one unit.
- A change-set MUST NOT mix unrelated value objectives.
- A change-set MAY contain a mix of code, test, and documentation changes only when those changes are directly connected.
- A change-set MUST target the smallest change that increases business value.
- A change-set MUST be complete for code consistency, tests, and documentation.

### Intent Alignment

Intent Alignment MUST create and use a stage-local `update_plan`, update statuses after each workflow step in this phase, and keep exactly one step `in_progress` at a time.

1. Intent Understanding
   - The agent MUST NOT treat the user instruction as literal and complete by default.
   - The agent MUST ask focused clarification questions to establish full intent and required context.
   - The agent MUST challenge instructions that appear unusual, inconsistent, risky, or controversial, and MUST explain concrete reasons for doubt.
   - The agent MUST NOT continue to planning when any ambiguity or contradiction remains.
2. Intent Approval
   - The phase MUST end with an explicit interpretation artifact.
   - The agent MUST obtain explicit user approval of that artifact before starting planning.

### Planning

Planning MUST create and use a stage-local `update_plan`, update statuses after each workflow step in this phase, and keep exactly one step `in_progress` at a time.

1. Measurable Success Criteria
   - The agent MUST define clear, measurable success criteria from a project-development perspective.
   - Criteria MUST be verifiable through engineering evidence (for example behavior, tests, quality gates, architecture constraints, or delivery artifacts).
   - The agent MUST avoid vague goals like "make it better" or "improve code quality".
   - Business outcome metrics (for example revenue, adoption, or installs) MUST NOT be used as success criteria in this phase.
   - For a bug fix, success criteria MUST include a regression test that fails before the fix and passes after the fix.
   - For new behavior, success criteria MUST include happy path, edge case, and error path verification.
   - For a behavior-preserving refactor, success criteria MUST include proof that behavior is unchanged.
   - For optimization, success criteria MUST include a correctness baseline first, then optimization evidence with preserved behavior.
2. Implementation Planning
   - The agent MUST create a detailed implementation plan that links product intent to technical execution.
   - For each planned change-set, the plan MUST describe product-side value delivered by the change and corresponding technical implementation vision.
   - The plan MUST present `Technical Scope` as a dedicated section inside each planned change-set.
   - The agent MUST NOT provide one aggregated technical-scope section shared across multiple change-sets.
   - The plan SHOULD be iterative and SHOULD split complex work into multiple change-sets.
   - The agent MUST default to multiple change-sets for non-trivial scope.
   - A single change-set MAY be used only when scope is trivial (for example one tightly scoped behavior in one layer) or when the user explicitly requests one change-set.
   - If a single change-set is chosen, the plan MUST include explicit justification why further decomposition would not improve delivery safety or reviewability.
3. Plan Verification
   - The agent MUST verify that the full plan achieves the intended goal.
   - The agent MUST verify that the full plan can meet the defined success criteria.
   - If gaps or risks are found, the agent MUST update the plan before implementation starts.

### Implementation

Implementation MUST iterate over all approved change-sets in the defined order.
Implementation MUST create and use a stage-local `update_plan`, update statuses after each workflow step in this phase, and keep exactly one step `in_progress` at a time.

1. Code and Test Execution
   - For project-code implementation, the agent MUST apply all active repository engineering guardrails.
   - During implementation, the agent MAY run verification tools iteratively for affected scope to speed up feedback.
2. Change-Set Verification
   - Before finalizing a change-set, the agent MUST run all mandatory verification commands required by active project governance.
   - If mandatory tests cannot run, the agent MUST explicitly report why.
3. Documentation and Test Cases
   - If a change-set modifies at least one non-documentation file in the repository, the agent MUST update required documentation according to project-defined rules.
   - If the change-set modifies product documentation, the agent MUST perform test-case impact analysis using project-defined policy.
4. Change-Set Commit
   - The agent MUST commit each completed change-set as exactly one commit.
   - The agent MUST NOT mark the change-set as completed before this commit exists.
5. Change-Set Closure Report
   - The agent MUST provide a short closure report for each completed change-set that includes:
     - change-set identifier,
     - commit hash,
     - mandatory verification command results for that change-set,
     - podsumowanie dokonanych zmian w dokumentacji.

### Completion

Completion MUST create and use a stage-local `update_plan`, update statuses after each workflow step in this phase, and keep exactly one step `in_progress` at a time.

1. Full-Plan Completion Verification
   - The agent MUST verify that all approved change-sets from the plan were implemented, or that any approved deviation is explicitly documented.
   - The agent MUST verify `one change-set = one commit` across the full plan and MUST explicitly list this mapping check result.
   - The agent MUST verify that measurable success criteria are satisfied for the full planned scope.
   - The agent MUST verify that required tests were added or updated according to active repository engineering guardrails, or that an exception is explicitly documented.
   - The agent MUST verify that all mandatory verification commands were completed for the full planned scope, or that a limitation is explicitly documented.
   - The agent MUST verify that mandatory tests pass, or that a limitation is explicitly documented.
   - The agent MUST verify that naming and terminology remain consistent across the full planned scope.
   - If any required verification item fails or is missing, the agent MUST stop completion and finish missing workflow steps first.
2. Final Completion Report
   - After completing the full planned scope, the agent MUST provide one final completion report.
   - The report MUST include `CHANGES MADE` as a file-level summary of what changed and why.
   - The report MUST include `RISKS / VERIFY` as potential regressions and additional checks to run, including manual tests.
   - The report MUST include `CHANGE-SET COMMITS` with each completed change-set mapped to exactly one commit hash.
   - The report MUST include all accepted local exceptions (for example linter or security suppressions) with concrete rationale.
   - The report SHOULD stay short and concrete so a junior engineer can quickly review and validate the result.
   - The agent MUST NOT publish a final completion response until all mandatory sections in this phase are present.

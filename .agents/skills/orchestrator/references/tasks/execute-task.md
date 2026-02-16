# How to Execute One PRD Task Safely and Consistently

## 1. Purpose
Use this guide when an AI agent must execute exactly one implementation task linked to an existing PRD.

The output must:
- complete one task end-to-end,
- keep task/PRD status lifecycle consistent,
- preserve dependency correctness,
- commit implementation changes for the completed task.

## 2. Core Rules (Non-Negotiable)
1. Use English only for all agent outputs in this workflow: clarifications, summaries, checks, and completion report.
2. This workflow executes exactly one task; do not execute multiple tasks in one run.
3. Execution mode must be `Default`; if mode is not `Default`, stop and request mode switch.
4. Execution must run through a subagent.
5. Input must identify either:
   - one explicit task file, or
   - one explicit PRD file.
6. If input is PRD-only, select the first executable task for that PRD.
7. A task is executable only if:
   - task `Status` is exactly `READY`, and
   - all tasks listed in `blocked-by` are `DONE`.
8. Execute work on branch derived from PRD filename:
   - branch name must equal lowercase PRD filename stem (`prd-[prd-id]-[short-name]`),
   - if branch does not exist, create it from `main` and checkout to it.
9. Required knowledge sources for implementation:
   - selected task content,
   - completion summaries from dependency tasks,
   - parent PRD content,
   - current codebase state and current documentation.
10. Verification must be executed according to selected task `Verification Plan`.
11. After successful implementation:
   - set task `Status` to `DONE`,
   - fill `Completion Summary` with concrete delivered changes and important follow-up context.
12. Commit completed task changes.
13. If technical/process issues are discovered and they can be prevented by AGENTS instructions, append concrete proposal(s) to `lessons-learned.md` in repository root as numbered list items.
14. If no other task in the same PRD remains in `READY`, set parent PRD `Status` to `DONE`.

## 3. Required Workflow (Execution Order)
Follow this sequence exactly:

1. Confirm execution mode.
   - Verify current mode is `Default`.
   - If mode is not `Default`, stop and request mode switch.
2. Validate input selector.
   - Accept exactly one selector: explicit task file or explicit PRD file.
   - If neither is provided, ask for one focused clarification.
3. Resolve execution target.
   - If task file is provided, use it as target.
   - If PRD file is provided, enumerate its tasks in deterministic order (`Task ID` ascending) and pick first executable task.
4. Validate task executability.
   - Confirm target task `Status: READY`.
   - Confirm every `blocked-by` dependency is `DONE`.
   - If not executable, stop and report exact blocker.
5. Resolve parent PRD and branch.
   - Identify parent PRD from task metadata.
   - Derive branch name from PRD filename stem and convert to lowercase (`prd-[prd-id]-[short-name]`).
   - If branch exists, checkout it.
   - If branch does not exist, checkout `main`, create branch, then checkout new branch.
6. Launch subagent execution.
   - Delegate implementation of selected task to subagent with strict task scope.
   - Provide subagent with required knowledge sources from Section 2, rule 9.
7. Verify implementation.
   - Run verification checks required by task `Verification Plan`.
   - If verification fails, iterate implementation until checks pass or report hard blocker.
8. Finalize task state.
   - Update task file: set `Status: DONE`.
   - Replace `Completion Summary` with factual delivery summary.
9. Capture reusable process lessons.
   - If applicable, append instruction proposals to `lessons-learned.md` as numbered list items.
10. Commit.
    - Create one commit containing task implementation and task-state update.
11. Finalize PRD state.
    - Inspect sibling tasks for same PRD.
    - If none has `Status: READY`, set parent PRD `Status: DONE`.
12. Commit PRD status change separately.
    - If parent PRD status changed in step 11, create a separate commit containing only PRD status update.
13. Publish concise completion report.
    - Include task executed, verification result, task commit hash, optional PRD-status commit hash, and PRD status result.

## 4. Task Selection Rules
1. When selector is PRD:
   - Consider only tasks belonging to that PRD.
   - Order by `Task ID` ascending.
   - Choose first task that satisfies executability rules.
2. If no executable task exists:
   - Do not implement anything.
   - Return exact reason (for example all `READY` tasks blocked, or no `READY` tasks).
3. Do not skip to later executable tasks if an earlier `READY` task is executable.

## 5. Branch Rules
1. Parent branch naming source is PRD filename, not task filename.
2. Branch name format must be PRD stem:
   - `prd-[prd-id]-[short-name]`
3. Creation flow when missing:
   - checkout `main`,
   - pull latest `main` when policy requires it,
   - create and checkout target branch.
4. Do not execute task on unrelated branch.

## 6. Verification and Completion Rules
1. Verification scope is defined by task `Verification Plan`; do not replace it with ad-hoc checks.
2. Task can be marked `DONE` only after verification passes.
3. `Completion Summary` must include:
   - key implementation outcomes,
   - tests/checks executed,
   - decisions relevant for downstream tasks.
4. Commit message should reflect executed task intent and scope.
5. If PRD status is changed to `DONE`, commit that status update in a separate commit from task implementation.

## 7. Quality Gates (All Must Pass)
1. Execution mode was `Default`.
2. Exactly one task was selected and executed.
3. Selected task was executable (`READY` + dependencies `DONE`) before implementation.
4. Task execution ran through subagent.
5. Branch matched parent PRD filename stem.
   - Name was lowercase.
6. Verification followed task `Verification Plan` and passed.
7. Task status was updated to `DONE` with non-empty `Completion Summary`.
8. Implementation was committed.
9. If PRD status changed, that change was committed separately.
10. `lessons-learned.md` entries were appended as numbered list items when instruction-worthy issues were discovered.
11. Parent PRD status was set to `DONE` when no sibling task remained `READY`.

## 8. Agent Output Contract
When running this workflow, return concise output with:
1. Selected input and resolved executed task file.
2. Executability validation result (`READY` + dependency check).
3. Branch resolution result (existing vs created).
   - Include final lowercase branch name.
4. Verification results against task `Verification Plan`.
5. Updated files list.
6. Task commit hash.
7. Optional PRD-status commit hash (when PRD status changed).
8. Whether parent PRD was moved to `DONE` or why not.

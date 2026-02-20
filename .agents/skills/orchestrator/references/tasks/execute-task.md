# How to Execute One PRD Task Safely and Consistently

## 1. Purpose
Use this guide when an AI agent must execute exactly one implementation task linked to an existing PRD.

The output must:
- complete one task end-to-end,
- keep task/PRD status lifecycle consistent,
- preserve dependency correctness,
- commit implementation changes for the completed task.

## 2. Core Rules (Non-Negotiable)
1. Apply shared baseline rules from `../../SKILL.md` section `Shared Workflow Baseline`.
2. This workflow executes exactly one task; do not execute multiple tasks in one run.
3. Input selector rules:
   - input may identify one explicit task file, one explicit PRD file, or both,
   - if both are provided and inconsistent, stop and ask for one focused selector clarification.
4. If input is PRD-only, select the first executable task for that PRD.
5. A task is executable only if:
   - task `Status` is exactly `READY`, and
   - all tasks listed in `blocked-by` are `DONE`.
6. Execute work on branch derived from PRD filename:
   - branch name must equal lowercase PRD filename stem (`prd-[prd-id]-[short-name]`, where `[prd-id]` is three digits),
   - if branch does not exist, create it from local `main` and checkout to it,
   - do not pull/sync automatically unless user explicitly requests it.
7. Required knowledge sources for implementation:
   - selected task content,
   - completion summaries from dependency tasks,
   - parent PRD content,
   - current codebase state and current documentation.
8. Verification must be executed according to selected task `Verification Plan`.
9. After successful implementation:
   - set task `Status` to `DONE`,
   - fill `Completion Summary` with concrete delivered changes and important follow-up context.
10. Commit completed task changes.
    - Task implementation commit message must include parent PRD ID reference (for example `PRD-003`).
11. Run mandatory lessons-learned harvest before final reporting:
   - check execution/user-feedback triggers: user correction/pushback, aborted/repeated turn, verification failure with rework, ambiguity clarification, documentation consistency issue,
   - if any trigger exists, append at least one concrete prevention rule to `lessons-learned.md` as numbered list item,
   - if no trigger exists, explicitly report `LESSONS LEARNED: no qualifying trigger`.
12. Set parent PRD `Status` to `DONE` only when both are true:
    - no other task in the same PRD remains in `READY`,
    - parent PRD `Release Criteria` are satisfied with explicit evidence from `DONE` task `Completion Summary` entries.
    If either condition is not met, keep parent PRD open and report the exact blocker.

## 3. Required Workflow (Execution Order)
Follow this sequence exactly:

1. Confirm execution mode.
   - Verify current mode is `Default`.
   - If mode is not `Default`, stop and request mode switch.
2. Validate input selector.
   - Accept explicit task selector, explicit PRD selector, or both.
   - If neither is provided, ask for one focused clarification.
   - If both are provided and the task belongs to the selected PRD, proceed with the explicit task file.
   - If both are provided and inconsistent, stop and ask the user to select one authoritative selector.
3. Resolve execution target.
   - If task file is provided, use it as target.
   - If PRD file is provided, enumerate its tasks in deterministic order (`Task ID` ascending) and pick first executable task.
   - Use hint `H003` from `../commands.md` when selector input is a PRD file and deterministic task order is required.
4. Validate task executability.
   - Confirm target task `Status: READY`.
   - Confirm every `blocked-by` dependency is `DONE`.
   - Parse `blocked-by` value as either:
     - `none`, or
     - comma+space separated Markdown links to `.tasks` files.
   - If `blocked-by` uses any other format, stop and report invalid dependency format.
   - Use hint `H004` from `../commands.md` when you need quick executability metadata for a candidate task.
   - Use hint `H005` from `../commands.md` when task has dependencies and sibling/dependency status snapshot is needed.
   - If not executable, stop and report exact blocker.
5. Resolve parent PRD and branch.
   - Identify parent PRD from task metadata.
   - Use hint `H006` from `../commands.md` when selector input is a task file and branch source PRD must be resolved quickly.
   - Derive branch name from PRD filename stem and convert to lowercase (`prd-[prd-id]-[short-name]`, where `[prd-id]` is three digits).
   - If branch exists, checkout it.
   - If branch does not exist, checkout local `main`, create branch from current local `main`, then checkout new branch.
6. Execute selected task implementation.
   - Perform implementation directly using required knowledge sources from Section 2, rule 8.
   - Use hint `H007` from `../commands.md` when you need `Completion Summary` context from current task or dependency tasks.
7. Verify implementation.
   - Run verification checks required by task `Verification Plan`.
   - Use hint `H008` from `../commands.md` when exact `Verification Plan` scope must be extracted without reading full task file.
   - If verification fails, iterate implementation until checks pass or report hard blocker.
8. Finalize task state.
   - Update task file: set `Status: DONE`.
   - Replace `Completion Summary` with factual delivery summary.
9. Capture reusable process lessons (mandatory scan).
   - Review the trigger set from Section 2, rule 12.
   - If at least one trigger occurred, append at least one concrete prevention rule to `lessons-learned.md` as numbered list item.
   - If no trigger occurred, record `LESSONS LEARNED: no qualifying trigger` in final report.
10. Commit.
    - Create one commit containing task implementation and task-state update.
    - Commit message must include parent PRD ID reference (for example `PRD-003`).
11. Finalize PRD state.
   - Inspect sibling tasks for same PRD.
   - Use hint `H009` from `../commands.md` when deciding whether parent PRD can be moved to `DONE`.
   - Validate parent PRD `Release Criteria` and confirm they are satisfied by explicit evidence in `DONE` task `Completion Summary` entries.
   - If none has `Status: READY` and release criteria evidence is sufficient, set parent PRD `Status: DONE`.
   - Otherwise keep parent PRD `Status` unchanged and report unmet release criteria or missing evidence.
12. Commit PRD status change separately.
    - If parent PRD status changed in step 11, create a separate commit containing only PRD status update.
13. Publish concise completion report.
    - Include task executed, verification result, task commit hash, optional PRD-status commit hash, and PRD status result.

## 4. Task Selection Rules
1. When selector is PRD:
   - Consider only tasks belonging to that PRD.
   - Use hint `H010` from `../commands.md` when identifying first executable task while preserving `Task ID` order.
   - Order by `Task ID` ascending.
   - Choose first task that satisfies executability rules.
2. If no executable task exists:
   - Do not implement anything.
   - Return exact reason (for example all `READY` tasks blocked, or no `READY` tasks).
3. Do not skip to later executable tasks if an earlier `READY` task is executable.

## 5. Branch Rules
1. Parent branch naming source is PRD filename, not task filename.
2. Branch name format must be PRD stem:
   - `prd-[prd-id]-[short-name]` where `[prd-id]` is three digits.
3. Creation flow when missing:
   - checkout `main`,
   - create and checkout target branch from current local `main` state,
   - do not pull latest `main` automatically; pull only when user explicitly asks.
4. Do not execute task on unrelated branch.

## 6. Verification and Completion Rules
1. Verification scope is defined by task `Verification Plan`; do not replace it with ad-hoc checks.
2. Task can be marked `DONE` only after verification passes.
3. `Completion Summary` must include:
   - key implementation outcomes,
   - tests/checks executed,
   - decisions relevant for downstream tasks.
4. Commit message should reflect executed task intent and scope, and must include parent PRD ID reference (for example `PRD-003`).
5. If PRD status is changed to `DONE`, commit that status update in a separate commit from task implementation.

## 7. Quality Gates (All Must Pass)
1. Execution mode was `Default`.
2. Exactly one task was selected and executed.
3. Selected task was executable (`READY` + dependencies `DONE`) before implementation.
4. Branch matched parent PRD filename stem.
   - Name was lowercase.
5. Verification followed task `Verification Plan` and passed.
6. Task status was updated to `DONE` with non-empty `Completion Summary`.
7. Implementation was committed.
   - Task implementation commit message included parent PRD ID reference (for example `PRD-003`).
8. If PRD status changed, that change was committed separately.
9. Mandatory lessons-learned scan was executed and outcome was reported.
   - If triggers occurred, `lessons-learned.md` was updated with at least one numbered prevention rule.
   - If no triggers occurred, final report contains `LESSONS LEARNED: no qualifying trigger`.
10. Parent PRD status was set to `DONE` only when no sibling task remained `READY` and parent PRD `Release Criteria` were satisfied with explicit evidence.

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
9. Release criteria evidence result used for PRD closure decision.
10. Lessons-learned harvest outcome (entries added or `LESSONS LEARNED: no qualifying trigger`).

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
3. Execution mode must be `Default`; if mode is not `Default`, stop and request mode switch.
4. Input must identify either:
   - one explicit task file, or
   - one explicit PRD file.
5. If input is PRD-only, select the first executable task for that PRD.
6. A task is executable only if:
   - task `Status` is exactly `READY`, and
   - all tasks listed in `blocked-by` are `DONE`.
7. Execute work on branch derived from PRD filename:
   - branch name must equal lowercase PRD filename stem (`prd-[prd-id]-[short-name]`),
   - if branch does not exist, create it from local `main` and checkout to it,
   - do not pull/sync automatically unless user explicitly requests it.
8. Required knowledge sources for implementation:
   - selected task content,
   - completion summaries from dependency tasks,
   - parent PRD content,
   - current codebase state and current documentation.
9. Verification must be executed according to selected task `Verification Plan`.
10. After successful implementation:
   - set task `Status` to `DONE`,
   - fill `Completion Summary` with concrete delivered changes and important follow-up context.
11. Commit completed task changes.
12. If technical/process issues are discovered and they can be prevented by AGENTS instructions, append concrete proposal(s) to `lessons-learned.md` in repository root as numbered list items.
13. Set parent PRD `Status` to `DONE` only when both are true:
    - no other task in the same PRD remains in `READY`,
    - parent PRD final acceptance matrix has passed (all required rows marked `PASS`).
    If either condition is not met, keep parent PRD open and report the exact blocker.

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
   - Hint: to list task files for one PRD in deterministic numeric order, run:
     ```bash
     rg --files .tasks | rg "^\\.tasks/PRD-[prd-id]-TASK-[0-9]+-" | sort -V
     ```
   - Use when: selector input is a PRD file and you must avoid accidental out-of-order task selection.
4. Validate task executability.
   - Confirm target task `Status: READY`.
   - Confirm every `blocked-by` dependency is `DONE`.
   - Hint: to verify one task quickly without reading full file content, run:
     ```bash
     rg -n "^- Status:|^- blocked-by:" .tasks/PRD-[prd-id]-TASK-[task-id]-*.md
     ```
   - Use when: you already have candidate task file and need only executability metadata.
   - Then verify dependency statuses for the same PRD:
     ```bash
     rg -n "^- Status:" .tasks/PRD-[prd-id]-TASK-*.md
     ```
   - Use when: task has one or more `blocked-by` dependencies and you need dependency status snapshot.
   - If not executable, stop and report exact blocker.
5. Resolve parent PRD and branch.
   - Identify parent PRD from task metadata.
   - Hint: to read parent PRD pointer without opening full task file, run:
     ```bash
     rg -n "^- PRD:" .tasks/PRD-[prd-id]-TASK-[task-id]-*.md
     ```
   - Use when: selector input is a task file and branch source PRD must be resolved quickly.
   - Derive branch name from PRD filename stem and convert to lowercase (`prd-[prd-id]-[short-name]`).
   - If branch exists, checkout it.
   - If branch does not exist, checkout local `main`, create branch from current local `main`, then checkout new branch.
6. Execute selected task implementation.
   - Perform implementation directly using required knowledge sources from Section 2, rule 8.
   - Hint: to read only `Completion Summary` of one task without opening the whole file, run:
     ```bash
     rg --multiline --multiline-dotall "^## Completion Summary\\n\\n([\\s\\S]*?)(?:\\n## [^\\n]*|\\z)" .tasks/PRD-[prd-id]-TASK-[task-id]-*.md --replace '$1'
     ```
   - Use when: you need compact status context for current task before coding or before final update.
   - Hint: to read only dependency task completion summaries, run:
     ```bash
     rg --multiline --multiline-dotall "^## Completion Summary\\n\\n([\\s\\S]*?)(?:\\n## [^\\n]*|\\z)" .tasks/PRD-[prd-id]-TASK-[dep-task-id]-*.md --replace '$1'
     ```
   - Use when: current task has `blocked-by` dependencies and you only need delivered outcomes from those tasks.
7. Verify implementation.
   - Run verification checks required by task `Verification Plan`.
   - Hint: to read only `Verification Plan` for selected task, run:
     ```bash
     rg --multiline --multiline-dotall "^## Verification Plan\\n\\n([\\s\\S]*?)(?:\\n## [^\\n]*|\\z)" .tasks/PRD-[prd-id]-TASK-[task-id]-*.md --replace '$1'
     ```
   - Use when: you need exact verification scope without loading full task details.
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
   - Hint: to quickly check whether any sibling task is still `READY`, run:
     ```bash
     rg -n "^- Status: READY$" .tasks/PRD-[prd-id]-TASK-*.md
     ```
   - Use when: deciding whether parent PRD can be moved to `DONE`.
   - Validate final acceptance matrix status in the parent PRD and confirm all required rows are `PASS`.
   - If none has `Status: READY` and matrix status is fully `PASS`, set parent PRD `Status: DONE`.
   - Otherwise keep parent PRD `Status` unchanged and report blocking rows or missing matrix evidence.
12. Commit PRD status change separately.
    - If parent PRD status changed in step 11, create a separate commit containing only PRD status update.
13. Publish concise completion report.
    - Include task executed, verification result, task commit hash, optional PRD-status commit hash, and PRD status result.

## 4. Task Selection Rules
1. When selector is PRD:
   - Consider only tasks belonging to that PRD.
   - Hint: to quickly inspect task statuses for one PRD without opening full files, run:
     ```bash
     rg -n "^- Task ID:|^- Status:" .tasks/PRD-[prd-id]-TASK-*.md | sort -V
     ```
   - Use when: identifying first executable task while preserving `Task ID` order.
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
4. Commit message should reflect executed task intent and scope.
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
8. If PRD status changed, that change was committed separately.
9. `lessons-learned.md` entries were appended as numbered list items when instruction-worthy issues were discovered.
10. Parent PRD status was set to `DONE` when no sibling task remained `READY`.
10. Parent PRD status was set to `DONE` only when no sibling task remained `READY` and final acceptance matrix rows were all `PASS`.

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
9. Final acceptance matrix result used for PRD closure decision.

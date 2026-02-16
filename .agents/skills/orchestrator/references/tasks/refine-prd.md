# How to Refine a PRD into Executable Technical Tasks

## 1. Purpose
Use this guide when an AI agent must transform an existing PRD into a set of implementation task files.

The output must:
- produce one or more task files linked to exactly one PRD,
- create tasks that can be executed incrementally,
- ensure each completed task keeps the software in a working state,
- define explicit task ordering and dependencies,
- avoid executing implementation work as part of this workflow.

## 2. Core Rules (Non-Negotiable)
1. This workflow creates/refines task files only; it does not implement code.
2. The referenced PRD is the main source of truth for scope.
3. Task planning must also reflect current application state and current documentation.
4. Every task must represent one clear technical objective.
5. Every task must include a working-software checkpoint (no broken intermediate target states).
6. Generate at least one task per PRD; most PRDs should produce multiple tasks.
7. Task status is restricted to `READY` or `DONE` only.
8. Parent PRD status must be `READY`; planning must not proceed when parent PRD status is missing, invalid, or `DONE`.
9. Dependencies are optional but, when used, must be explicit using `blocked-by` and/or `blocks`.
10. Parallel execution is allowed only for tasks with all blockers completed.
11. Every task must explicitly reference:
    - parent PRD,
    - dependency tasks (or explicit `none`).
12. All task files must be Markdown.
13. Do not include unresolved placeholders such as `TBD` or `TODO`.
14. If supplemental material is provided (for example user stories), treat it as additive guidance and never as replacement for PRD truth.
15. This task specification is project-agnostic and must not depend on repository-specific architecture assumptions.
16. This workflow is two-phase: draft tasks in `Plan` mode, save task files in `Default` mode.
17. Do not save task files while still in `Plan` mode.
18. If current mode does not match required phase, stop and request mode switch before continuing.

## 3. Required Workflow (Execution Order)
Follow this sequence exactly:

1. Confirm execution mode.
   - Verify current execution mode is `Plan` for draft phase.
   - If mode is not `Plan`, do not continue; ask for switch to `Plan` mode.
2. Validate inputs.
   - Confirm PRD file path.
   - Confirm PRD ID from filename pattern `PRD-[prd-id]-[short-name].md`.
   - Confirm PRD contains explicit `Status`.
   - Continue only if parent PRD status is exactly `READY`.
   - If PRD status is missing, invalid, or `DONE`, stop and request PRD status correction/reopening before task drafting.
   - Load current codebase state and current documentation at a high level to avoid stale planning.
3. Load optional supplemental references.
   - If user provided extra files (for example user stories), load them as secondary constraints.
4. Detect critical unknowns.
   - If task boundaries cannot be derived safely, ask focused clarifying questions before drafting.
5. Build task graph.
   - Decompose PRD into minimal vertical slices.
   - Ensure each slice is independently verifiable and leaves software working.
   - Determine ordering and dependency edges.
6. Draft task files using fixed structure (Section 5).
7. Run one internal review pass.
   - Check sequencing, dependency validity, and execution readiness.
8. Run draft quality pre-checks for gates that do not require file-save phase evidence.
9. Request mode switch for save phase.
   - Ask for switch to `Default` mode before writing task files.
10. Confirm save mode.
   - Verify current execution mode is `Default`.
   - If mode is not `Default`, do not save; request switch and wait.
11. Save task files in `.tasks` with required naming (Section 9).
12. Run final full quality gate check and publish a concise summary listing generated tasks and dependency order.

## 4. Clarification Protocol
Ask clarifying questions only when missing information blocks safe planning.

Priority clarification topics:
1. PRD scope boundaries that impact task count or order.
2. Required release slices if PRD can be delivered in multiple valid sequences.
3. Constraints that change dependency graph (for example compliance, migration windows, rollout gating).

Rules:
- Ask one focused question at a time.
- Provide clear options when possible.
- Do not generate final task files while critical unknowns remain unresolved.

## 5. Fixed Task File Structure (Must Be Exact)
Each generated task file must use these headings in this exact order:

1. `Overview`
2. `Metadata`
3. `Objective`
4. `Working Software Checkpoint`
5. `Technical Scope`
6. `Implementation Plan`
7. `Verification Plan`
8. `Acceptance Criteria`
9. `Dependencies`
10. `Completion Summary`

### 5.1 Section Rules
1. `Overview`
   - One short paragraph describing task intent and business/feature context.
2. `Metadata`
   - Must include:
     - `Status`: `READY` or `DONE`
     - `PRD`: exact PRD filename
     - `Task ID`: integer for current PRD sequence
     - `Task File`: current task filename
3. `Objective`
   - One technical objective only.
4. `Working Software Checkpoint`
   - Describe how software remains usable after this task alone.
5. `Technical Scope`
   - Split into `In Scope` and `Out of Scope`.
6. `Implementation Plan`
   - Ordered, concrete execution steps for this task only.
7. `Verification Plan`
   - Explicit checks to prove task completion.
8. `Acceptance Criteria`
   - Observable and testable outcomes only.
   - Include project validation requirement as final criterion.
9. `Dependencies`
   - Must include:
     - `blocked-by`: explicit task file links or `none`
     - `blocks`: explicit task file links or `none`
   - Dependency entries must be Markdown links to `.tasks` files (optionally with task ID label).
   - Example:
     - `[PRD-12-TASK-1-config-foundation](.tasks/PRD-12-TASK-1-config-foundation.md)`
10. `Completion Summary`
    - If `Status: READY`, set to: `Not started`.
    - If `Status: DONE`, provide concrete summary of delivered work and decisions important for follow-up tasks.

## 6. Task Splitting and Ordering Rules
1. Prefer small, single-purpose tasks that can be completed in one focused implementation iteration.
2. Split by dependency boundaries first (for example contracts/data foundations before dependent behavior).
3. Then split by independently verifiable behavior.
4. Do not create "umbrella" tasks that span unrelated subsystems.
5. Keep task descriptions specific enough to be implemented without reinterpretation.
6. Order tasks by dependency graph first, then by logical delivery sequence.

## 7. Dependency Rules
1. Allowed dependency relations:
   - `blocked-by`: tasks that must be `DONE` before current task starts.
   - `blocks`: tasks that depend on current task.
2. Dependency graph must be acyclic.
3. If no dependency exists, write `none`.
4. Every dependency reference must point to an existing task file for the same PRD.
5. Every dependency reference must be an explicit Markdown link to `.tasks/PRD-[prd-id]-TASK-[task-id]-[short-task-name].md`.
6. If task ID is additionally shown, it must match the linked filename.
7. Sequential execution is default; parallel execution is allowed only when all `blocked-by` tasks are `DONE`.

## 8. Status Lifecycle Rules
1. Allowed statuses:
   - `READY`: task planned and available to execute.
   - `DONE`: task executed and closed.
2. `DONE` tasks must include a filled `Completion Summary`.
3. Preserve and carry forward `Completion Summary` context from `DONE` tasks because downstream tasks may rely on it.
4. Do not change `DONE` back to `READY` unless user explicitly requests reopening.

## 9. File Output Rules
When saving tasks:

1. Save only in `Default` mode; if current mode is `Plan`, stop and request switch to `Default`.
2. Save in `.tasks`.
3. Use filename format:
   - `PRD-[prd-id]-TASK-[task-id]-[short-task-name].md`
4. `prd-id` must match parent PRD ID.
5. `task-id` must be sequential from `1` to `N` within that PRD.
6. Use kebab-case for `[short-task-name]`.
7. Each task file must contain explicit PRD reference and explicit linked dependency references in content.

## 10. Quality Gates (All Must Pass Before Final Output)
1. At least one task is generated for the PRD.
2. Parent PRD contains explicit valid `Status` and it is exactly `READY`.
3. Every task file follows the exact structure from Section 5.
4. Every task has exactly one technical objective.
5. Every task has a working-software checkpoint.
6. Every task has explicit `In Scope` and `Out of Scope`.
7. Every task includes verifiable acceptance criteria.
8. Every task includes `Status` and it is either `READY` or `DONE`.
9. Every task references the correct parent PRD.
10. Every task has explicit `blocked-by` and `blocks` fields (`none` allowed), using links not plain IDs.
11. Dependency references are valid, resolvable, and acyclic.
12. File names follow required naming format.
13. No task includes unresolved placeholders.
14. Draft phase execution mode was `Plan`.
15. Save phase execution mode was `Default`.

## 11. Forbidden Content
Do not include:
- implementation execution logs as part of planning output,
- multi-task combined objectives in a single task file,
- ambiguous dependency statements (for example "depends on previous work"),
- statuses other than `READY` or `DONE`,
- orphan tasks without PRD reference.

## 12. Agent Output Contract
When running this workflow:

1. Confirm draft phase execution happened in `Plan` mode.
2. Confirm the PRD input used.
3. Confirm detected PRD status is `READY`.
4. Confirm whether supplemental references were used.
5. Confirm mode switch to `Default` happened before saving files.
6. Produce/save task files in `.tasks` with required naming.
7. Return concise summary:
   - generated task count,
   - ordered task list,
   - dependency highlights,
   - quality gate result (`PASS`/`FAIL` per gate).

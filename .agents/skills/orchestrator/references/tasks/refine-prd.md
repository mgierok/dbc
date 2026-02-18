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
1. Apply shared baseline rules from `../../SKILL.md` section `Shared Workflow Baseline`.
2. This workflow creates/refines task files only; it does not implement code.
3. The referenced PRD is the main source of truth for scope.
4. Task planning must also reflect current application state and current documentation.
5. Every task must represent one clear technical objective.
6. Every task must include a working-software checkpoint (no broken intermediate target states).
7. Generate at least one task per PRD; most PRDs should produce multiple tasks.
8. Task status is restricted to `READY` or `DONE` only.
9. Parent PRD status must be `READY`; planning must not proceed when parent PRD status is missing, invalid, or `DONE`.
10. Every task must include explicit `blocked-by` and `blocks` fields.
    - Allowed value format per field:
      - `none`, or
      - one or more Markdown links to task files separated by comma+space.
11. Parallel execution is allowed only for tasks with all blockers completed.
12. Every task must explicitly reference:
    - parent PRD,
    - dependency tasks (or explicit `none`).
13. Every task must include explicit traceability to parent PRD requirement IDs (`FR-*` and/or `NFR-*`), and every parent PRD requirement must be covered by at least one task.
14. All task files must be Markdown.
15. If supplemental material is provided (for example user stories), treat it as additive guidance and never as replacement for PRD truth.
16. This task specification is repository-aware and must align task planning with repository-specific architecture assumptions, tooling, and constraints.
17. This workflow is two-phase: draft tasks in `Plan` mode, save task files in `Default` mode.
18. For every parent `FR-*`, plan at least:
    - one happy-path verification scenario,
    - one negative-path verification scenario.
    Each scenario must be mapped to a specific task and an explicit test/check in that task `Verification Plan`.
19. If the PRD is split into multiple implementation tasks and behavior spans task boundaries, include a final integration task named with `integration-hardening` scope.
    - This task must be blocked by all behavior-delivering tasks for that PRD.
    - This task must verify cross-task interactions and regression coverage before PRD closure.
20. Every parent PRD metric in `Success Metrics and Release Criteria` must be mapped to at least one task.
21. Each mapped metric must have an execution-phase measurement checkpoint in task `Verification Plan` that is measurable before PRD closure.
22. For post-release outcome metrics, define at least one delivery-phase proxy checkpoint used for go/no-go decisions.
23. Do not finalize tasks with metric placeholders (`Not measured`, `Unknown`, `TBD`, `to be defined`) in checkpoints or metric mappings.
24. Metric evidence must be recorded in the same task file `Completion Summary` when that task is marked `DONE`.
25. Do not plan separate evidence-documentation paths for metric proof unless the user explicitly asks for them.
26. Use `../templates/task-template.md` as the single source of truth for task file structure; do not infer structure from existing files in `.tasks`.

## 3. Required Workflow (Execution Order)
Follow this sequence exactly:

1. Confirm execution mode.
   - Verify current execution mode is `Plan` for draft phase.
   - If mode is not `Plan`, do not continue; ask for switch to `Plan` mode.
2. Validate inputs.
   - Confirm PRD file path.
   - Confirm PRD ID from filename pattern `PRD-[prd-id]-[short-name].md`.
   - Confirm PRD contains explicit `Status`.
   - Confirm PRD requirements are explicitly identifiable for traceability (for example `FR-*`, `NFR-*` IDs).
   - Confirm PRD metrics are explicitly identifiable for traceability (for example `M1`, `M2`, `M3`).
   - Continue only if parent PRD status is exactly `READY`.
   - If PRD status is missing, invalid, or `DONE`, stop and request PRD status correction/reopening before task drafting.
   - If PRD requirement identifiers are missing or ambiguous, stop and request PRD correction before task drafting.
   - If PRD metric identifiers are missing or ambiguous, stop and request PRD correction before task drafting.
   - Load current codebase state and current documentation at a high level to avoid stale planning.
3. Load optional supplemental references.
   - If user provided extra files (for example user stories), load them as secondary constraints.
4. Detect critical unknowns.
   - If task boundaries cannot be derived safely, ask focused clarifying questions before drafting.
5. Build task graph.
   - Decompose PRD into minimal vertical slices.
   - Ensure each slice is independently verifiable and leaves software working.
   - Determine ordering and dependency edges.
   - Build a requirement coverage map `FR/NFR -> TASK-*` and ensure full parent-PRD requirement coverage.
   - Build a metric coverage map `M* -> TASK-*` and ensure full parent-PRD metric coverage.
   - Build a verification coverage map `FR-* -> (happy-path task + test/check, negative-path task + test/check)`.
   - Build a metric checkpoint map `M* -> (task + evidence artifact + execution-phase threshold/check)`.
   - If a parent metric is post-release by nature, add delivery-phase proxy checkpoints mapped to tasks and release decision use.
   - If the feature spans multiple tasks and cross-task interactions exist, append a final `integration-hardening` task that depends on all prior behavior-delivering tasks.
6. Draft task files from template `../templates/task-template.md`.
   - Instantiate template structure for each task first.
   - Keep heading names and heading order exactly as in the template.
   - Fill all required metadata and content fields for each task.
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
4. PRD requirement statements that cannot be mapped unambiguously to task boundaries.
5. PRD metric statements that cannot be measured during execution or mapped to explicit task checkpoints.

Rules:
- Ask one focused question at a time.
- Provide clear options when possible.
- Do not generate final task files while critical unknowns remain unresolved.

## 5. Fixed Task Template (Must Be Exact)
Apply Section 2 rule 26 as the single template contract.
During drafting, instantiate the template first, keep heading names/order unchanged, and fill all sections with concrete task-specific content.

### 5.1 Section Rules
1. `Overview`
   - One short paragraph describing task intent and business/feature context.
2. `Metadata`
   - Must include:
     - `Status`: `READY` or `DONE`
     - `PRD`: exact PRD filename
     - `Task ID`: integer for current PRD sequence
     - `Task File`: current task filename
     - `Task ID` must match `[task-id]` segment in `Task File` filename.
     - `PRD Requirements`: explicit list of covered parent PRD requirement IDs (`FR-*` and/or `NFR-*`)
     - `PRD Metrics`: explicit list of covered parent PRD metric IDs (`M*`) or `none` when the task does not carry metric checkpoints
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
   - For every mapped `PRD Metrics` item, include one execution-phase metric checkpoint with:
     - metric ID,
     - evidence source artifact (must reference this task file `Completion Summary`),
     - threshold or expected value for the current phase,
     - check procedure.
8. `Acceptance Criteria`
   - Observable and testable outcomes only.
   - Include project validation requirement as final criterion.
9. `Dependencies`
   - Must include:
     - `blocked-by`: `none` or Markdown links separated by comma+space
     - `blocks`: `none` or Markdown links separated by comma+space
   - Dependency entries must be Markdown links to `.tasks` files (optionally with task ID label).
   - Example:
     - `[PRD-12-TASK-1-config-foundation](.tasks/PRD-12-TASK-1-config-foundation.md)`
10. `Completion Summary`
    - If `Status: READY`, set to: `Not started`.
    - If `Status: DONE`, provide concrete summary of delivered work and decisions important for follow-up tasks.
    - If task maps `PRD Metrics`, include metric evidence entries for each mapped metric with observed result and threshold pass/fail.

## 6. Task Splitting and Ordering Rules
1. Prefer small, single-purpose tasks that can be completed in one focused implementation iteration.
2. Split by dependency boundaries first (for example contracts/data foundations before dependent behavior).
3. Then split by independently verifiable behavior.
4. Do not create "umbrella" tasks that span unrelated subsystems.
5. Keep task descriptions specific enough to be implemented without reinterpretation.
6. Order tasks by dependency graph first, then by logical delivery sequence.
7. Ensure every parent PRD requirement (`FR-*` / `NFR-*`) is covered by at least one task, with no orphan requirements.
8. For multi-task features with cross-task interaction risk, reserve the last task for `integration-hardening` and set dependencies so it executes after all behavior-delivering tasks.

## 7. Dependency Rules
1. Allowed dependency relations:
   - `blocked-by`: tasks that must be `DONE` before current task starts.
   - `blocks`: tasks that depend on current task.
2. Dependency graph must be acyclic.
3. If no dependency exists, write `none`.
4. If multiple dependencies exist in one field, use comma+space separated Markdown links on the same line.
5. Dependency references must be Markdown links to `.tasks/PRD-[prd-id]-TASK-[task-id]-[short-task-name].md`.
6. Every dependency reference must point to an existing task file for the same PRD.
7. If task ID is additionally shown, it must match the linked filename.
8. Sequential execution is default; parallel execution is allowed only when all `blocked-by` tasks are `DONE`.

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
3. Every task file follows the template structure from `../templates/task-template.md` and Section 5.1 rules.
4. Every task has exactly one technical objective.
5. Every task has a working-software checkpoint.
6. Every task has explicit `In Scope` and `Out of Scope`.
7. Every task includes verifiable acceptance criteria.
8. Every task includes `Status` and it is either `READY` or `DONE`.
9. Every task references the correct parent PRD.
10. Every task includes explicit `PRD Requirements` with valid parent-PRD requirement IDs (`FR-*` / `NFR-*`).
11. Combined task set provides full parent-PRD requirement coverage (no orphan `FR-*` / `NFR-*`).
12. Every parent `FR-*` is mapped to at least one happy-path and one negative-path scenario, each linked to a specific task and explicit test/check in that task `Verification Plan`.
13. Every parent PRD metric is mapped to at least one task (`M* -> TASK-*`) with no orphan metrics.
14. Every mapped metric has an execution-phase checkpoint in a task `Verification Plan`, including explicit evidence source artifact and threshold or expected value.
15. If a parent metric is post-release by nature, at least one delivery-phase proxy checkpoint is defined and mapped to task(s) for release decision support.
16. No metric mapping or checkpoint includes placeholders (`Not measured`, `Unknown`, `TBD`, `to be defined`).
17. If the feature uses multiple behavior-delivering tasks with cross-task interactions, the task set includes a final `integration-hardening` task blocked by all those tasks.
18. Every task has explicit `blocked-by` and `blocks` fields (`none` allowed), using links not plain IDs.
19. Dependency references are valid, resolvable, and acyclic.
20. File names follow required naming format.
21. `Task ID` metadata value matches `[task-id]` in filename for every task.
22. No task includes unresolved placeholders.
23. Draft phase execution mode was `Plan`.
24. Save phase execution mode was `Default`.
25. Every mapped metric checkpoint uses `Completion Summary` of the same task file as the evidence source artifact.
26. Template compliance
   - Every generated task preserves heading names and order from `../templates/task-template.md`.

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
   - traceability highlights (`FR/NFR/M -> TASK-*`),
   - dependency highlights,
   - quality gate result (`PASS`/`FAIL` per gate).

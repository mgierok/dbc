---
name: orchestrator
description: Orchestrate repeatable multi-step workflows by selecting one task specification from `references/tasks/*.md` and executing it strictly. Use when the user asks for standardized process execution managed by this skill, including running the Create PRD task from `references/tasks/create-prd.md`, the Refine PRD task from `references/tasks/refine-prd.md`, or the Execute Task workflow from `references/tasks/execute-task.md`.
---

# Orchestrator

## Goal

Execute one defined workflow from `references/tasks/` at a time, with strict adherence to the selected task file contract.

## Task Routing

1. Identify user intent.
2. Match intent to one task file in `references/tasks/`.
3. If multiple tasks could match, ask one focused disambiguation question.
4. Load only the matched task file and execute it literally.

## Available Tasks

- `Create PRD` (`references/tasks/create-prd.md`)
  - Trigger examples: "create PRD", "prepare product requirements", "build PRD from short prompt".
  - Execution rule: follow this task file exactly as written; this workflow is two-phase: draft in `Plan` mode, save in `Default` mode.
- `Refine PRD` (`references/tasks/refine-prd.md`)
  - Trigger examples: "refine PRD", "break PRD into tasks", "generate implementation tasks from PRD".
  - Execution rule: follow this task file exactly as written; this workflow is two-phase: draft in `Plan` mode, save in `Default` mode.
- `Execute Task` (`references/tasks/execute-task.md`)
  - Trigger examples: "execute task", "run next task from PRD", "implement PRD task".
  - Execution rule: follow this task file exactly as written; this workflow executes exactly one task in `Default` mode.

## Execution Rules

1. Treat every selected task file as authoritative for that workflow.
2. Preserve required order, required formats, and non-negotiable constraints.
3. Do not merge or blend rules between task files unless a task file explicitly says to do so.
4. If task instructions conflict with higher-priority runtime constraints, explain the conflict and apply the safe fallback.
5. Enforce lifecycle invariants across workflows:
   - parent PRD must be `READY` for planning/refinement workflows,
   - parent PRD can move to `DONE` only after execution closes all `READY` tasks for that PRD,
   - task statuses use `READY`/`DONE` only.
6. If selected task requires mode switch, pause at phase boundary and request explicit switch before any file-save step.
7. Keep outputs concise and verifiable against the selected task's quality gates.

## Shared Workflow Baseline

These rules apply to all orchestrator workflows unless a task file defines a stricter rule:

1. Use English only for all agent outputs.
2. Respect mode boundaries strictly:
   - draft phases run in `Plan`,
   - save/write phases run in `Default`,
   - if mode mismatches the required phase, stop and request a mode switch.
3. Do not allow unresolved placeholders (`TBD`, `TODO`) in final saved artifacts.
4. End each workflow with concise quality gate results (`PASS`/`FAIL` per gate).

## Shared Templates

Use these templates as the only source of truth for target file structure:

- PRD template: `references/templates/prd-template.md`
- Task template: `references/templates/task-template.md`

Rules:

1. When creating PRDs or tasks, instantiate the proper template and fill all required sections.
2. Do not infer structure by inspecting existing files in `.tasks`.
3. If structure changes are needed, update the corresponding template and the related task specification in the same change.

## Extending This Skill

1. Add each new workflow as a separate file in `references/tasks/`.
2. Keep each task file self-contained for workflow-specific behavior and reference `Shared Workflow Baseline` for common constraints, with clear:
   - purpose,
   - required workflow/order,
   - output contract,
   - quality checks.
3. Update `Available Tasks` in this file when adding a new task.

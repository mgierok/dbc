---
name: orchestrator
description: Orchestrate repeatable multi-step workflows by selecting one task specification from `references/tasks/*.md` and executing it strictly. Use when the user asks for standardized process execution managed by this skill, including running the Create PRD task from `references/tasks/create-prd.md` or the Refine PRD task from `references/tasks/refine-prd.md`.
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
  - Execution rule: follow this task file exactly as written.
- `Refine PRD` (`references/tasks/refine-prd.md`)
  - Trigger examples: "refine PRD", "break PRD into tasks", "generate implementation tasks from PRD".
  - Execution rule: follow this task file exactly as written; this workflow must run in `Plan` mode.

## Execution Rules

1. Treat every selected task file as authoritative for that workflow.
2. Preserve required order, required formats, and non-negotiable constraints.
3. Do not merge or blend rules between task files unless a task file explicitly says to do so.
4. If task instructions conflict with higher-priority runtime constraints, explain the conflict and apply the safe fallback.
5. Enforce status lifecycle constraints defined by the selected task (for example parent PRD must be `READY` for `refine-prd`, while task statuses use `READY`/`DONE`).
6. Keep outputs concise and verifiable against the selected task's quality gates.

## Extending This Skill

1. Add each new workflow as a separate file in `references/tasks/`.
2. Keep each task file self-contained, with clear:
   - purpose,
   - required workflow/order,
   - output contract,
   - quality checks.
3. Update `Available Tasks` in this file when adding a new task.

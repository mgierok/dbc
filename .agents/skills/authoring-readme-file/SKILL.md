---
name: authoring-readme-file
description: Assess whether changes require updating `README.md`, and create/update `README.md` only when required. Use only when the user explicitly requests README assessment or README changes.
---

# Authoring README File

## Purpose and Authority

- This skill MUST be the single source of truth for deciding whether README updates are required.
- This skill MUST be the single source of truth for authoring README updates.
- This skill MUST be used only for an explicit user request to assess, create, or modify `README.md`.
- This skill MUST NOT be auto-invoked solely because a task changes one or more non-documentation files.

## Scope

- Target content file MUST be `README.md` only.
- When this skill is active, the agent MUST NOT modify files other than `README.md`.
- README content MUST stay user-facing and task-oriented.
- README content MUST NOT duplicate internal architecture/process governance details from internal documentation.
- README content SHOULD prioritize installation, supported scope, and startup usage over exhaustive runtime reference material.
- README content MUST NOT duplicate a comprehensive keyboard/control reference when that material is better owned by product documentation or in-app help, unless the user explicitly requests that duplication.

## Required Reference

- Before decision and writing, the agent MUST load `.agents/skills/authoring-readme-file/references/readme-template.md`.

## Structural Contract (Mandatory)

- `README.md` MUST stay aligned with the section set and order from `.agents/skills/authoring-readme-file/references/readme-template.md`.
- Existing heading anchors SHOULD be preserved unless restructuring is explicitly requested.
- If one fact fits multiple sections, it SHOULD be placed in the most user-actionable section, and duplication SHOULD be avoided.

## Decision Contract (Mandatory)

- The agent MUST return exactly one status:
  - `UPDATE_REQUIRED`
  - `NO_UPDATE_REQUIRED`
- Every decision output MUST include:
  - `EVIDENCE`: concrete user-facing facts supporting the decision
  - `IMPACTED SECTIONS`: sections mapped to `.agents/skills/authoring-readme-file/references/readme-template.md` (or `none`)
- The agent MUST apply strict evidence threshold:
  - `UPDATE_REQUIRED` MUST be used only when at least one documented README fact materially changed.
  - When evidence is insufficient or ambiguous, the agent MUST return `NO_UPDATE_REQUIRED`.
  - The agent MUST NOT require README updates based on speculation.

## Decision Rules (Mandatory)

Return `UPDATE_REQUIRED` when at least one of these user-facing facts materially changed:

- installation prerequisites or primary installation path
- supported database scope
- startup usage or canonical CLI examples (including direct launch examples)
- user-facing license pointer
- user-facing project value proposition in README intro/overview

Return `NO_UPDATE_REQUIRED` when changes are README-irrelevant, for example:

- internal refactors without user-visible CLI behavior or usage change
- test-only changes that do not alter user-facing usage
- formatting, naming, or comments-only edits without user-facing impact
- changes limited to documentation files (no implementation change)
- changes limited to detailed keyboard controls or command-reference material that README does not own

## Writing Rules

- Content MUST be concise, actionable, and copy-paste oriented for CLI users.
- Command examples MUST be runnable as shown.
- Content MUST describe current factual behavior only.
- Language SHOULD prioritize "what to run" and "what happens."
- README content SHOULD summarize runtime interaction only when it materially improves first-run usability.

## Workflow (Mandatory)

1. Gather factual user-facing changes from non-documentation files.
2. Apply the Decision Rules and produce `DECISION`, `EVIDENCE`, and `IMPACTED SECTIONS`.
3. If decision is `NO_UPDATE_REQUIRED`, the agent MUST stop after decision output.
4. If decision is `UPDATE_REQUIRED`, the agent MUST update `README.md` only.
5. After update, the agent MUST verify structural alignment with `.agents/skills/authoring-readme-file/references/readme-template.md`.
6. The completion output MUST include:
   - `DECISION`
   - `EVIDENCE`
   - `IMPACTED SECTIONS`
   - `CHANGES MADE`
   - `RISKS / VERIFY`

## Consistency and Integrity Contract (Mandatory)

- This file MUST remain internally consistent and free of contradictory normative rules.
- The agent MUST keep this file and `.agents/skills/authoring-readme-file/references/readme-template.md` non-contradictory.
- If structural requirements in this file change, the template MUST be updated in the same change set.

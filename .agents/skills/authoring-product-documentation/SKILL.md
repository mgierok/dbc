---
name: authoring-product-documentation
description: Assess whether non-documentation codebase changes require updating `docs/product-documentation.md`, and create/update that file only when required. Use for every non-documentation code change and for explicit product-documentation requests.
---

# Authoring Product Documentation

## Purpose and Authority

- This skill MUST be the single source of truth for deciding whether product documentation updates are required.
- This skill MUST be the single source of truth for authoring updates to product documentation.
- This skill MUST be used:
  - for every task that changes at least one non-documentation file in the codebase,
  - for every explicit request to create or modify `docs/product-documentation.md`.

## Scope

- Target content file MUST be `docs/product-documentation.md` only.
- When this skill is active, the agent MUST NOT modify files other than `docs/product-documentation.md`.
- Product documentation content MUST cover product-facing facts only:
  - product purpose and user value,
  - capability scope and boundaries,
  - user-visible behavior and outcomes,
  - user flows and interaction model,
  - user-visible constraints, non-goals, and safety semantics.
- The agent MUST NOT describe implementation mechanics unless needed to explain user-visible behavior.
- The agent MUST NOT document development workflow process details (for example branching/PR/release flow).

## Required Reference

- Before decision and writing, the agent MUST load `references/product-documentation-template.md`.

## Structural Contract (Mandatory)

- `docs/product-documentation.md` MUST stay aligned with the section set and order from `references/product-documentation-template.md`.
- Existing section numbering and anchors MUST be preserved unless structural migration is explicitly requested.
- If one fact fits multiple sections, it SHOULD be placed in the most product-specific section, and duplication SHOULD be avoided.

## Decision Contract (Mandatory)

- The agent MUST return exactly one status:
  - `UPDATE_REQUIRED`
  - `NO_UPDATE_REQUIRED`
- Every decision output MUST include:
  - `EVIDENCE`: concrete product-facing facts supporting the decision
  - `IMPACTED SECTIONS`: sections mapped to `references/product-documentation-template.md` (or `none`)
- The agent MUST apply strict evidence threshold:
  - `UPDATE_REQUIRED` MUST be used only when at least one documented product-facing fact materially changed.
  - When evidence is insufficient or ambiguous, the agent MUST return `NO_UPDATE_REQUIRED`.
  - The agent MUST NOT require documentation updates based on speculation.

## Decision Rules (Mandatory)

Return `UPDATE_REQUIRED` when at least one of these product-facing facts materially changed:

- product purpose, audience, or current user value proposition
- in-scope/out-of-scope capability boundaries
- user-visible behavior, outcomes, errors, or status communication
- user flows, navigation, controls, shortcuts, commands, or interaction model
- user-visible dialogs, confirmations, save/discard/cancel semantics, or safety guards
- user-visible constraints, non-goals, platform support statements, or glossary terminology

Return `NO_UPDATE_REQUIRED` when changes are product-documentation-irrelevant, for example:

- internal implementation refactors without user-visible behavior change
- test-only changes that do not alter user-visible behavior or scope
- formatting, naming, or comments-only edits without user-visible impact
- changes limited to documentation files (no implementation change)

## Writing Rules

- Content MUST describe current factual product state only.
- Content MUST stay explicit, user-oriented, and behavior-focused.
- Language SHOULD stay concise and unambiguous for product and engineering audiences.

## Workflow (Mandatory)

1. Gather factual product-facing changes from non-documentation files.
2. Apply the Decision Rules and produce `DECISION`, `EVIDENCE`, and `IMPACTED SECTIONS`.
3. If decision is `NO_UPDATE_REQUIRED`, the agent MUST stop after decision output.
4. If decision is `UPDATE_REQUIRED`, the agent MUST update `docs/product-documentation.md` only.
5. After update, the agent MUST verify structural alignment with `references/product-documentation-template.md`.
6. The completion output MUST include:
   - `DECISION`
   - `EVIDENCE`
   - `IMPACTED SECTIONS`
   - `CHANGES MADE`
   - `THINGS NOT TOUCHED`
   - `RISKS / VERIFY`

## Consistency and Integrity Contract (Mandatory)

- This file MUST remain internally consistent and free of contradictory normative rules.
- The agent MUST keep this file and `references/product-documentation-template.md` non-contradictory.
- If structural requirements in this file change, the template MUST be updated in the same change set.

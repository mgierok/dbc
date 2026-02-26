---
name: authoring-technical-documentation
description: Assess whether non-documentation codebase changes require updating `docs/technical-documentation.md`, and create/update that file only when required. Use for every non-documentation code change and for explicit technical-documentation requests.
---

# Authoring Technical Documentation

## Purpose and Authority

- This skill MUST be the single source of truth for deciding whether technical documentation updates are required.
- This skill MUST be the single source of truth for authoring updates to technical documentation.
- This skill MUST be used:
  - for every task that changes at least one non-documentation file in the codebase,
  - for every explicit request to create or modify `docs/technical-documentation.md`.

## Scope

- Target content file MUST be `docs/technical-documentation.md` only.
- When this skill is active, the agent MUST NOT modify files other than `docs/technical-documentation.md`.
- Technical documentation content MUST cover implementation-facing facts only:
  - architecture and boundaries,
  - components and responsibilities,
  - technical mechanisms and contracts,
  - runtime and operational behavior,
  - technical constraints, risks, and tradeoffs.
- The agent MUST NOT document product-facing behavior unless needed to explain technical mechanisms.
- The agent MUST NOT document development workflow process details (for example branching/PR/release flow).

## Required Reference

- Before decision and writing, the agent MUST load `references/technical-documentation-template.md`.

## Structural Contract (Mandatory)

- `docs/technical-documentation.md` MUST stay aligned with the section set and order from `references/technical-documentation-template.md`.
- Existing section numbering and anchors MUST be preserved unless structural migration is explicitly requested.
- If one fact fits multiple sections, it SHOULD be placed in the most implementation-specific section, and duplication SHOULD be avoided.

## Decision Contract (Mandatory)

- The agent MUST return exactly one status:
  - `UPDATE_REQUIRED`
  - `NO_UPDATE_REQUIRED`
- Every decision output MUST include:
  - `EVIDENCE`: concrete implementation facts supporting the decision
  - `IMPACTED SECTIONS`: sections mapped to `references/technical-documentation-template.md` (or `none`)
- The agent MUST apply strict evidence threshold:
  - `UPDATE_REQUIRED` MUST be used only when at least one documented implementation fact materially changed.
  - When evidence is insufficient or ambiguous, the agent MUST return `NO_UPDATE_REQUIRED`.
  - The agent MUST NOT require documentation updates based on speculation.

## Decision Rules (Mandatory)

Return `UPDATE_REQUIRED` when at least one of these implementation facts materially changed:

- architecture boundaries, dependency direction, or component responsibilities
- interfaces, contracts, schemas, or integration mechanics
- runtime lifecycle, startup/shutdown flow, operational behavior, or failure handling
- persistence, transaction, state, concurrency, resilience, security, or observability mechanisms
- technical constraints, risks, tradeoffs, or stack/version facts
- deep-dive references in technical documentation became outdated due to codebase changes

Return `NO_UPDATE_REQUIRED` when changes are technical-documentation-irrelevant, for example:

- formatting, naming, or comments-only edits without factual behavior change
- test-only changes that do not alter implementation contracts/mechanisms
- equivalent internal refactors that preserve documented contracts, behavior, and constraints
- changes limited to documentation files (no implementation change)

## Writing Rules

- Content MUST describe current factual implementation state only.
- Content MUST stay practical and code-aligned.
- Language SHOULD stay concise, concrete, and maintenance-focused for engineers.

## Workflow (Mandatory)

1. Gather factual implementation changes from non-documentation files.
2. Apply the Decision Rules and produce `DECISION`, `EVIDENCE`, and `IMPACTED SECTIONS`.
3. If decision is `NO_UPDATE_REQUIRED`, the agent MUST stop after decision output.
4. If decision is `UPDATE_REQUIRED`, the agent MUST update `docs/technical-documentation.md` only.
5. After update, the agent MUST verify structural alignment with `references/technical-documentation-template.md`.
6. The completion output MUST include:
   - `DECISION`
   - `EVIDENCE`
   - `IMPACTED SECTIONS`
   - `CHANGES MADE`
   - `RISKS / VERIFY`

## Consistency and Integrity Contract (Mandatory)

- This file MUST remain internally consistent and free of contradictory normative rules.
- The agent MUST keep this file and `references/technical-documentation-template.md` non-contradictory.
- If structural requirements in this file change, the template MUST be updated in the same change set.

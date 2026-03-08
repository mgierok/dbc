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
- Running this skill for every non-documentation change MUST NOT imply that an update is always required.
- This skill MUST decide if an update is required based on material technical impact in the latest implementation changes.

## Scope

- Target content file MUST be `docs/technical-documentation.md` only.
- When this skill is active, the agent MUST NOT modify files other than `docs/technical-documentation.md`.
- Technical documentation content MUST cover implementation-facing facts only, with focus on durable engineering guidance:
  - architecture and boundaries,
  - stable components and responsibilities,
  - technical contracts and invariants,
  - runtime constraints and failure-relevant behavior,
  - technical constraints, risks, and tradeoffs.
- The agent MUST NOT document product-facing behavior unless needed to explain technical mechanisms.
- The agent MUST NOT document development workflow process details (for example branching/PR/release flow).
- The agent MUST NOT turn documentation into a chronological implementation log.
- The agent MUST document rules and constraints maintainers need before coding, not exhaustive UI/runtime narration.

## Required Reference

- Before decision and writing, the agent MUST load `.agents/skills/authoring-technical-documentation/references/technical-documentation-template.md`.

## Structural Contract (Mandatory)

- `docs/technical-documentation.md` MUST stay aligned with the section set and order from `.agents/skills/authoring-technical-documentation/references/technical-documentation-template.md`.
- Section titles and anchors SHOULD stay stable across updates.
- Section numbering MUST NOT be used.
- If one fact fits multiple sections, it SHOULD be placed in the most implementation-specific section, and duplication SHOULD be avoided.
- Sections MAY be intentionally short when no material facts exist beyond baseline constraints.

## Decision Contract (Mandatory)

- The agent MUST return exactly one status:
  - `UPDATE_REQUIRED`
  - `NO_UPDATE_REQUIRED`
- Every decision output MUST include:
  - `EVIDENCE`: concrete implementation facts supporting the decision
  - `IMPACTED SECTIONS`: sections mapped to `.agents/skills/authoring-technical-documentation/references/technical-documentation-template.md` (or `none`)
- The agent MUST apply strict evidence threshold:
  - `UPDATE_REQUIRED` MUST be used only when at least one documented durable technical fact materially changed.
  - When evidence is insufficient or ambiguous, the agent MUST return `NO_UPDATE_REQUIRED`.
  - The agent MUST NOT require documentation updates based on speculation.

## Materiality Test (Mandatory)

Before returning `UPDATE_REQUIRED`, the agent MUST pass all checks:

1. The change affects a durable technical rule, contract, decision, or constraint that engineers rely on during implementation.
2. Omitting the update would likely cause incorrect future code changes, architecture drift, or invalid assumptions.
3. The fact is not merely local implementation detail that is already obvious from a narrow code read.

If any check fails, the agent MUST return `NO_UPDATE_REQUIRED`.

## Decision Rules (Mandatory)

Return `UPDATE_REQUIRED` when at least one of these implementation facts materially changed:

- architecture boundaries, dependency direction, or component responsibilities
- interfaces, contracts, schemas, or integration mechanics
- runtime lifecycle invariants, operational constraints, or failure handling contracts
- persistence, transaction, state, concurrency, resilience, security, or observability mechanisms
- technical constraints, risks, tradeoffs, or stack/version facts
- deep-dive references in technical documentation became outdated due to codebase changes
- a new stable engineering rule/decision became part of the codebase and should guide future development

Return `NO_UPDATE_REQUIRED` when changes are technical-documentation-irrelevant, for example:

- formatting, naming, or comments-only edits without factual behavior change
- test-only changes that do not alter implementation contracts/mechanisms
- equivalent internal refactors that preserve documented contracts, behavior, and constraints
- changes limited to documentation files (no implementation change)
- local UI copy/layout/keybinding details that do not change durable technical contracts
- step-order edits inside existing flows that do not change invariants, boundaries, or failure semantics

## Writing Rules

- Content MUST describe current factual implementation state only.
- Content MUST stay practical and code-aligned.
- Language SHOULD stay concise, concrete, and maintenance-focused for engineers.
- Each updated section SHOULD be contract-first: what rule exists, why it matters, where it is enforced.
- The agent SHOULD prefer concise bullet points over long narrative paragraphs.
- The agent MUST NOT add exhaustive step-by-step flow descriptions unless they define an invariant or failure contract.
- The agent SHOULD keep only the minimal set of facts required for safe future changes.

## Workflow (Mandatory)

1. Gather factual implementation changes from non-documentation files.
2. Apply the Materiality Test and Decision Rules to produce `DECISION`, `EVIDENCE`, and `IMPACTED SECTIONS`.
3. If decision is `NO_UPDATE_REQUIRED`, the agent MUST stop after decision output.
4. If decision is `UPDATE_REQUIRED`, the agent MUST update only impacted sections and MUST keep unchanged sections untouched.
5. If decision is `UPDATE_REQUIRED`, the agent MUST add new technical rules/decisions when they became stable in implementation.
6. After update, the agent MUST verify structural alignment with `.agents/skills/authoring-technical-documentation/references/technical-documentation-template.md`.
7. The completion output MUST include:
   - `DECISION`
   - `EVIDENCE`
   - `IMPACTED SECTIONS`
   - `CHANGES MADE`
   - `RISKS / VERIFY`

## Consistency and Integrity Contract (Mandatory)

- This file MUST remain internally consistent and free of contradictory normative rules.
- The agent MUST keep this file and `.agents/skills/authoring-technical-documentation/references/technical-documentation-template.md` non-contradictory.
- If structural requirements in this file change, the template MUST be updated in the same change set.

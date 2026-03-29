---
name: authoring-technical-documentation
description: Assess whether changes require updating `docs/technical-documentation.md`, and create/update that file only when required. Use when the user explicitly requests technical-documentation assessment or technical-documentation changes, and whenever the agent updates `docs/technical-documentation.md`.
---

# Authoring Technical Documentation

## Purpose and Authority

- This skill MUST be the single source of truth for deciding whether technical documentation updates are required.
- This skill MUST be the single source of truth for authoring updates to technical documentation.
- This skill MUST be used for an explicit user request to assess, create, or modify `docs/technical-documentation.md`.
- This skill MUST be used whenever the agent updates `docs/technical-documentation.md`, even if the user did not name this skill explicitly.
- This skill MUST NOT be auto-invoked solely because a task changes one or more non-documentation files without updating `docs/technical-documentation.md`.
- When explicitly invoked, this skill MUST decide if an update is required based on material technical impact in the latest implementation changes.
- `docs/technical-documentation.md` MUST describe current factual implementation state only.
- `docs/technical-documentation.md` MUST NOT act as the canonical architecture policy for future changes.
- When current implementation diverges from `docs/clean-architecture-ddd.md`, the skill SHOULD document the divergence explicitly when it is material.
- The skill MUST distinguish between present-state facts and canonical architecture expectations.

## Scope

- Target content file MUST be `docs/technical-documentation.md` only.
- When this skill is active, the agent MUST NOT modify files other than `docs/technical-documentation.md`.
- Technical documentation content MUST cover implementation-facing facts only, with focus on current factual implementation reference:
  - current architecture boundaries and dependency shape,
  - current stable components and responsibilities,
  - current technical contracts and mechanisms,
  - current runtime constraints, risks, and failure-relevant behavior,
  - current technical tradeoffs, stack facts, and known drift.
- The agent MUST NOT document product-facing behavior unless needed to explain technical mechanisms.
- The agent MUST NOT document development workflow process details (for example branching/PR/release flow).
- The agent MUST NOT turn documentation into a chronological implementation log.
- The agent MUST document current implementation facts maintainers need to understand before changing code, not exhaustive UI/runtime narration.
- The agent MUST distinguish factual current-state content from canonical expectations defined in `AGENTS.md` or `docs/clean-architecture-ddd.md`.
- The agent MUST NOT present `docs/technical-documentation.md` as governance for future architecture or implementation changes.

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
  - `UPDATE_REQUIRED` MUST be used only when at least one documented current technical fact materially changed.
  - When evidence is insufficient or ambiguous, the agent MUST return `NO_UPDATE_REQUIRED`.
  - The agent MUST NOT require documentation updates based on speculation.

## Materiality Test (Mandatory)

Before returning `UPDATE_REQUIRED`, the agent MUST pass all checks:

1. The change affects a material current implementation state, technical contract, currently observed boundary, documented drift, or runtime constraint/risk.
2. Omitting the update would likely cause incorrect assumptions about the current codebase, hide material drift from canonical architecture, or misstate active technical constraints.
3. The fact is not merely local implementation detail that is already obvious from a narrow code read.

If any check fails, the agent MUST return `NO_UPDATE_REQUIRED`.

## Decision Rules (Mandatory)

Return `UPDATE_REQUIRED` when at least one of these implementation facts materially changed:

- current implementation state, architecture boundaries, dependency shape, or component responsibilities
- current interfaces, contracts, schemas, or integration mechanics
- current runtime lifecycle invariants, operational constraints, failure handling contracts, or active risks
- currently implemented persistence, transaction, state, concurrency, resilience, security, or observability mechanisms
- current technical constraints, risks, tradeoffs, or stack/version facts
- currently observed boundaries or documented drift against `docs/clean-architecture-ddd.md`
- deep-dive references in technical documentation became outdated due to codebase changes

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
- Each updated section SHOULD be contract-first: what is true now, why it matters now, where the current behavior is enforced now.
- The agent SHOULD prefer concise bullet points over long narrative paragraphs.
- The agent MUST NOT add exhaustive step-by-step flow descriptions unless they define an invariant or failure contract.
- The agent SHOULD keep only the minimal set of facts required for safe future changes.
- The agent MUST distinguish present-state facts from canonical expectations; when both are relevant, the skill MUST describe the difference explicitly as current-state versus canonical expectation or drift.
- The agent MUST NOT use future-prescriptive architecture language unless the prescription is explicitly attributed to `AGENTS.md` or `docs/clean-architecture-ddd.md`.

## Workflow (Mandatory)

1. Gather factual implementation changes from non-documentation files.
2. Apply the Materiality Test and Decision Rules to produce `DECISION`, `EVIDENCE`, and `IMPACTED SECTIONS`.
3. If decision is `NO_UPDATE_REQUIRED`, the agent MUST stop after decision output.
4. If decision is `UPDATE_REQUIRED`, the agent MUST update only impacted sections and MUST keep unchanged sections untouched.
5. If decision is `UPDATE_REQUIRED`, the agent MUST update current-state facts, contracts, constraints, and material drift without introducing new governance rules.
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

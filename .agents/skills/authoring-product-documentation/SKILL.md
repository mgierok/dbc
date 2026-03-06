---
name: authoring-product-documentation
description: Assess whether non-documentation codebase changes require updating `docs/product-documentation.md`, and create/update that file only when required. Use for every non-documentation code change and for explicit product-documentation requests.
---

# Authoring Product Documentation

## Purpose and Authority

- This skill MUST be the single source of truth for deciding whether `docs/product-documentation.md` requires an update.
- This skill MUST be the single source of truth for authoring updates to `docs/product-documentation.md`.
- This skill MUST be used:
  - for every task that changes at least one non-documentation file in the codebase,
  - for every explicit request to create or modify `docs/product-documentation.md`.

## Scope

- The target content file MUST be `docs/product-documentation.md`.
- When this skill is active for a product-documentation task, the agent MUST NOT modify non-documentation files unless the user explicitly requests a related governance/template update.
- Product documentation MUST cover product-facing facts only:
  - product purpose and user value,
  - capability scope and boundaries,
  - user-visible behavior and outcomes,
  - interaction patterns when they matter to users,
  - user-visible constraints, non-goals, and safety semantics.
- The agent MUST NOT document implementation mechanics unless they are required to explain user-visible behavior.
- The agent MUST NOT document engineering workflow process details.

## Required Reference

- Before decision or writing, the agent MUST load `references/product-documentation-template.md`.

## Structural Contract

- `docs/product-documentation.md` MUST preserve the mandatory section set and order defined in `references/product-documentation-template.md`.
- The section title `## Functional Behavior` MUST be preserved exactly.
- Optional sections from `references/product-documentation-template.md` MAY be included only when they add unique product-facing information.
- Existing section titles and order MUST NOT change unless the user explicitly requests a structural migration.
- Facts MUST be placed in the single best-fit section and MUST NOT be repeated across multiple sections.

## Decision Contract

- The agent MUST return exactly one decision status:
  - `UPDATE_REQUIRED`
  - `NO_UPDATE_REQUIRED`
- Every decision output MUST include:
  - `EVIDENCE`: concrete product-facing facts that support the decision
  - `IMPACTED SECTIONS`: section names from `references/product-documentation-template.md`, or `none`
- `UPDATE_REQUIRED` MUST be used only when at least one documented product-facing fact materially changed.
- When evidence is insufficient or ambiguous, the agent MUST return `NO_UPDATE_REQUIRED`.
- The agent MUST NOT require documentation updates based on speculation.

## Decision Rules

Return `UPDATE_REQUIRED` when at least one of these materially changed:

- product purpose, audience, or current user value
- in-scope or out-of-scope capability boundaries
- user-visible behavior, outcomes, errors, or status communication
- user-visible navigation, controls, commands, or interaction rules
- user-visible confirmations, save/discard/cancel semantics, or safety guards
- user-visible constraints, non-goals, platform support statements, or glossary terms

Return `NO_UPDATE_REQUIRED` for changes such as:

- internal refactors with no user-visible behavior change
- test-only changes with no user-visible behavior or scope change
- comments, formatting, or naming-only edits with no user-visible impact
- documentation-only changes outside the need to update product facts

## Authoring Goals

- Content MUST describe the current factual product state only.
- Content MUST stay explicit, user-oriented, and behavior-focused.
- Content SHOULD be concise enough to support fast scanning by product and engineering readers.
- The document SHOULD prefer grouped summaries over exhaustive enumerations when the shorter form preserves the same product meaning.
- The document MUST treat `Functional Behavior` as the canonical detailed behavior section.
- The document SHOULD default to the smallest section set that still preserves clarity.

## Section Roles and Concision Rules

### Product Overview

- This section MUST summarize product purpose, target users, current value, and high-level scope boundaries.
- This section SHOULD absorb high-level capability summary instead of using a separate capabilities section.
- This section SHOULD stay high level and SHOULD NOT restate detailed behavior from later sections.

### Functional Behavior

- This section MUST remain present and MUST be the most detailed section in the document.
- This section MUST describe user-visible behavior, outcomes, error handling, and state transitions in a way that supports deterministic test-case mapping.
- This section SHOULD organize behavior by product area or workflow stage.
- This section MAY include exact user-visible commands, messages, or states when those details are behavior-defining.

### Interaction Model

- This section MAY be omitted when interaction rules are simple enough to describe inside `Functional Behavior`.
- This section MUST describe the stable interaction grammar that users operate with.
- This section SHOULD use concise tables for shortcuts, commands, or control patterns when that is clearer than prose.
- This section MUST NOT repeat behavior outcomes that are already documented in `Functional Behavior` unless a brief pointer is needed for clarity.

### Constraints and Non-Goals

- This section MUST contain only current user-visible limits and explicit non-goals.
- This section SHOULD absorb cross-cutting safety limits when they do not justify a standalone section.
- This section MUST NOT repeat in-scope capabilities or detailed flow logic already documented elsewhere unless the limit itself is the product fact.

### Glossary

- This section MUST contain only terms that materially improve reader understanding.
- Terms that are obvious from common product language SHOULD be omitted.

## Default Minimal Structure

- The default document structure SHOULD be:
  - `## Product Overview`
  - `## Functional Behavior`
  - `## Constraints and Non-Goals`
  - `## Glossary`
- `## Interaction Model` SHOULD be added only when the product uses a stable control grammar, shortcut system, or command language that would otherwise clutter `Functional Behavior`.
- Separate sections for capability summaries, user flows, or safety semantics SHOULD NOT be added by default; those facts SHOULD be absorbed into the core sections unless a standalone section clearly improves comprehension without duplicating content.

## Compression Heuristics

- If the same fact appears in more than one section, the agent MUST keep the most specific occurrence and remove the rest.
- If a section can be replaced by a short grouped summary without losing product meaning, the agent SHOULD prefer the summary.
- Long bullet lists SHOULD be collapsed into grouped bullets when the individual items do not need separate emphasis.
- Repeated shortcut aliases, repeated navigation steps, and repeated confirmation semantics SHOULD appear once in the strongest-fit section.
- Visual styling details SHOULD be documented only when they are user-visible product semantics, not merely implementation or layout trivia.
- If a standalone section contains only a few facts that fit cleanly into an adjacent section, the agent SHOULD merge it instead of preserving the section.

## Workflow

1. Gather product-facing facts from the codebase or the user request.
2. Load `references/product-documentation-template.md`.
3. Apply the Decision Rules and produce `DECISION`, `EVIDENCE`, and `IMPACTED SECTIONS`.
4. If the decision is `NO_UPDATE_REQUIRED`, the agent MUST stop after the decision output.
5. If the decision is `UPDATE_REQUIRED`, the agent MUST update `docs/product-documentation.md`.
6. After the update, the agent MUST verify:
   - mandatory section order matches the template,
   - `Functional Behavior` is preserved,
   - optional sections were included only when justified,
   - duplicated facts were removed or reduced,
   - the final document is concise without losing product facts.
7. The completion output MUST include:
   - `DECISION`
   - `EVIDENCE`
   - `IMPACTED SECTIONS`
   - `CHANGES MADE`
   - `RISKS / VERIFY`

## Consistency and Integrity Contract

- This file MUST remain internally consistent and free of contradictory rules.
- This file and `references/product-documentation-template.md` MUST stay non-contradictory.
- If section-role or structural requirements change in this file, the template MUST be updated in the same change set.

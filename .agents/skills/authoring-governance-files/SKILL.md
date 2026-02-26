---
name: authoring-governance-files
description: Create or edit governance files for AI coding agents. Use when the user asks to create or update SKILL.md or AGENTS.md files. Enforce junior-friendly clarity, RFC normative wording, explicit normative-compliance checks, and minimal redundancy through edit-first consolidation.
---

# Authoring Governance Files

## Purpose
Standardize creation and editing of governance files for AI coding agents, focused only on `SKILL.md` and `AGENTS.md`.

## Scope
- In scope: create or edit `SKILL.md`, create or edit `AGENTS.md`.
- Out of scope: other file types, unless explicitly requested by the user.

## Rule Groups (Modular Structure)
- `G1` Generic Rules
- `G2` SKILL.md-Specific Rules
- `G3` AGENTS.md-Specific Rules
- `G4` Anti-Patterns and Quality Checks

When modifying this skill:
1. Extend or refine an existing rule group first.
2. Add a new rule only when extension is not possible.
3. Add a short rationale when a new rule is introduced.

## G1 Generic Rules

### G1.1 Language and Readability
- Target governance content MUST be written in English.
- Instructions MUST be understandable for a Junior Software Engineer.
- Sentences SHOULD be short, direct, and concrete.
- Writers SHOULD avoid legalese and overloaded phrasing.
- Writers SHOULD avoid vague qualifiers such as `properly`, `somehow`, `etc.`, `as needed`, unless bounded by clear conditions.

### G1.2 RFC Normative Vocabulary
Use RFC-style keywords only for normative statements:
- `MUST`: absolute requirement.
- `MUST NOT`: absolute prohibition.
- `SHOULD`: strong default; deviations require explicit justification.
- `SHOULD NOT`: strong discouragement; deviations require explicit justification.
- `MAY`: optional behavior.

### G1.5 Normative Coverage Requirement
- Every added or edited normative instruction MUST include at least one RFC keyword (`MUST`, `MUST NOT`, `SHOULD`, `SHOULD NOT`, `MAY`).
- Imperative governance text without RFC keywords MUST be rewritten as:
  - normative text with RFC keywords, or
  - explicitly non-normative explanatory context.
- RFC keywords MUST be uppercase and used with their RFC semantics from `G1.2`.
- Non-normative context SHOULD avoid imperative verbs that can be misread as hidden requirements.

### G1.3 Consolidation-First Editing Policy
When adding or changing instructions:
1. Locate existing rules covering the same intent.
2. Modify, merge, or generalize existing rules first.
3. Add a new rule only if safe consolidation is impossible.
4. Keep final output optimized, cohesive, and free from unnecessary redundancy.

### G1.4 Consistency Requirement
- Updated governance content MUST NOT contain contradictory normative statements.
- If two rules overlap, they SHOULD be unified into one clearer rule.

## G2 SKILL.md-Specific Rules (Codex-Adapted Best Practices)

### G2.1 Core Skill File Design
- A skill MUST contain `SKILL.md`.
- Optional folders MAY include `references/`, `scripts/`, and `assets/`.
- `SKILL.md` frontmatter SHOULD use `name` and `description` as default minimal metadata.
- `name` MUST be kebab-case.
- `description` MUST define both:
  1. what the skill does,
  2. when it should be triggered.

### G2.2 Trigger Quality
- Trigger phrasing MUST be specific enough to reduce over-triggering and under-triggering.
- “When to use” criteria MUST be present in frontmatter `description`, not only in body sections.

### G2.3 Progressive Disclosure
- `SKILL.md` SHOULD contain compact procedural instructions.
- Detailed examples, references, and long-form material SHOULD be placed in `references/`.
- Reference files SHOULD be linked directly from `SKILL.md` (one-level navigation).

### G2.4 External Best-Practices Usage Rule
- Canonical best-practices reference file: `references/The-Complete-Guide-to-Building-Skill-for-Claude.md`.
- Best-practices material from external guides SHOULD be referenced:
  1. when creating a new skill, or
  2. when explicitly requested by the user during skill editing.
- Claude-specific guidance that conflicts with Codex behavior MUST be removed or rewritten for Codex compatibility.

## G3 AGENTS.md-Specific Rules

### G3.1 Policy Intent
- `AGENTS.md` MUST define practical operating rules for coding agents within the repository.
- Rules MUST be testable, unambiguous, and operational.

### G3.2 Authoring Quality
- Each rule SHOULD define condition, expected behavior, and boundary/exception (if relevant).
- Overlapping instructions SHOULD be merged.
- Duplicative restatements SHOULD be removed.

### G3.3 Editing Discipline
When editing `AGENTS.md`:
1. Prefer updating existing rules in-place.
2. Preserve section coherence and terminology consistency.
3. Add new rules only when no existing section can safely absorb the intent.
4. Numeric heading prefixes (for example `1.`, `1.1`, `2.`) MUST remain sequential and consistent after edits.
5. Whenever Markdown headings are changed (title or numeric prefix), the agent MUST update inbound heading references across the repository in the same change set.

## G4 Anti-Patterns and Quality Checks

### G4.1 Anti-Patterns
- Duplicate normative rules in multiple sections.
- Conflicting `MUST`/`SHOULD` statements for the same behavior.
- Imperative rule-like statements without RFC keywords (hidden normative requirements).
- Vague requirements without verification criteria.
- Overly long and multi-clause sentences that reduce junior readability.
- Skill descriptions that explain only “what” but not “when”.

### G4.2 Final Quality Gate
Before finalizing governance output, verify:
1. Scope is limited to `SKILL.md` and/or `AGENTS.md`.
2. Content is in English and junior-readable.
3. Every added/edited normative instruction uses RFC keywords with uppercase form and correct semantics.
4. No imperative hidden requirements remain without RFC keywords.
5. Existing rules were refined first; new rules were added only when necessary.
6. No contradiction or avoidable redundancy remains.
7. Numeric heading prefixes remain sequential and consistent after edits.
8. Inbound heading references were updated when heading titles or numeric prefixes were changed.

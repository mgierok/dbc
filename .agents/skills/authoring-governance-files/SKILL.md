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

## Generic Rules

### Role, Purpose, and Structural Flow

- Target governance documents MUST make role, purpose, and scope clear near the top.
- A dedicated role/purpose section MAY be omitted when the title, frontmatter, and opening sections already make that context explicit.
- Scope MUST be explicit in opening sections and top-level constraints.
- Document structure MUST flow from global constraints to local details.

### Language and Readability

- Target governance content MUST be written in English.
- Instructions MUST be understandable for a Junior Software Engineer.
- Sentences SHOULD be short, direct, and concrete.
- Writers SHOULD avoid legalese and overloaded phrasing.
- Writers SHOULD avoid vague qualifiers such as `properly`, `somehow`, `etc.`, `as needed`, unless bounded by clear conditions.

### Section and Heading Structure

- Markdown headings MUST use stable levels (`#`, `##`, `###`) and predictable hierarchy.
- Section titles MUST be explicit and topic- or action-oriented.
- Each section SHOULD answer one operational question (or one tightly related cluster).

### Normative Language and Definitions

Use RFC-style keywords only for normative statements:

- `MUST`: absolute requirement.
- `MUST NOT`: absolute prohibition.
- `SHOULD`: strong default; deviations require explicit justification.
- `SHOULD NOT`: strong discouragement; deviations require explicit justification.
- `MAY`: optional behavior.
- Every added or edited normative instruction MUST include at least one RFC keyword (`MUST`, `MUST NOT`, `SHOULD`, `SHOULD NOT`, `MAY`).
- Normative text MUST define required, prohibited, recommended, discouraged, or optional behavior using RFC keywords.
- The keywords `MUST`, `MUST NOT`, `SHOULD`, `SHOULD NOT`, `MAY` MUST be interpreted with RFC 2119 semantics.
- RFC keywords MUST be uppercase.
- Imperative governance text without RFC keywords MUST be rewritten as:
  - normative text with RFC keywords, or
  - explicitly non-normative explanatory context.
- Non-normative text MAY provide rationale, examples, or context, but MUST NOT introduce new requirements.
- Non-normative context SHOULD avoid imperative verbs that can be misread as hidden requirements.
- Mixed strength levels in a single sentence SHOULD be avoided; when used, conditions and exceptions MUST be explicit.

### Operational Style, Behavior Boundaries, and Exceptions

- Rules MUST describe concrete actions whenever possible.
- Allowed and forbidden behavior MUST be discoverable, either in dedicated sections or clear thematic sections.
- Exceptions MUST be explicitly labeled (for example `Exception:`) and MUST define trigger conditions and boundaries.
- Rationale and examples MAY be included when they improve implementation quality.
- Limited subjective wording MAY be used for design direction only when paired with actionable constraints.

### Consolidation-First Editing Policy

When adding or changing instructions:

1. Locate existing rules covering the same intent.
2. Modify, merge, or generalize existing rules first.
3. Add a new rule only if safe consolidation is impossible.
4. Keep final output optimized, cohesive, and free from unnecessary redundancy.

### Consistency and Redundancy Control

- Updated governance content MUST NOT contain contradictory normative statements.
- If two rules overlap, they SHOULD be unified into one clearer rule.
- Formatting conventions for equivalent structures in updated governance content MUST remain uniform and coherent within each file (for example heading style, list style, lead-in punctuation, capitalization, and terminology).
  Rationale: consistent formatting improves scanability and reduces interpretation mistakes.

### Process Order and Finalization

- Numbered phases MUST be used when strict workflow order is required.
- For lightweight workflows, ordered lists SHOULD be used without mandatory goal/entry/exit fields.
- Finalization MUST happen only after required validation steps are complete.

### Output and Formatting Requirements

- Output requirements and formatting requirements MUST both be explicit; they MAY be colocated or split.
- Required formatting constraints MUST be concrete and testable.
- Examples SHOULD be used to reduce ambiguity and MAY serve as quick templates.

## SKILL.md-Specific Rules (Codex-Adapted Best Practices)

### Core Skill File Design

- A skill MUST contain `SKILL.md`.
- Optional folders MAY include `references/`, `scripts/`, and `assets/`.
- `SKILL.md` frontmatter SHOULD use `name` and `description` as default minimal metadata.
- `name` MUST be kebab-case.
- `description` MUST define both:
  1. what the skill does,
  2. when it should be triggered.

### Trigger Quality

- Trigger phrasing MUST be specific enough to reduce over-triggering and under-triggering.
- “When to use” criteria MUST be present in frontmatter `description`, not only in body sections.

### Progressive Disclosure

- `SKILL.md` SHOULD contain compact procedural instructions.
- Detailed examples, references, and long-form material SHOULD be placed in `references/`.
- Reference files SHOULD be linked directly from `SKILL.md` (one-level navigation).

### External Best-Practices Usage Rule

- Canonical best-practices reference file: `.agents/skills/authoring-governance-files/references/The-Complete-Guide-to-Building-Skill-for-Claude.md`.
- Best-practices material from external guides SHOULD be referenced:
  1. when creating a new skill, or
  2. when explicitly requested by the user during skill editing.
- Claude-specific guidance that conflicts with Codex behavior MUST be removed or rewritten for Codex compatibility.

## AGENTS.md-Specific Rules

### Policy Intent

- `AGENTS.md` MUST define practical operating rules for coding agents within the repository.
- Rules MUST be testable, unambiguous, and operational.

### Authoring Quality

- Each rule SHOULD define condition, expected behavior, and boundary/exception (if relevant).
- Overlapping instructions SHOULD be merged.
- Duplicative restatements SHOULD be removed.
- Source-document routing guidance SHOULD be kept separate from implementation guardrails when both concerns exist.
- Sections that point to source documents SHOULD state what operational questions those documents answer for the agent.
- `AGENTS.md` sections such as architecture or dependencies/toolchain SHOULD stay governance-focused and SHOULD NOT restate detailed technical facts already delegated to source documents, unless a short enforcement summary materially improves compliance.

### Editing Discipline

When editing `AGENTS.md`:

1. Prefer updating existing rules in-place.
2. Preserve section coherence and terminology consistency.
3. Add new rules only when no existing section can safely absorb the intent.
4. If numeric heading prefixes are used (for example `1.`, `1.1`, `2.`), they MUST remain sequential and consistent after edits.
5. Whenever Markdown headings are changed (title or numeric prefix), the agent MUST update inbound heading references across the repository in the same change set.
6. If source-of-truth document guidance exists, it SHOULD appear before engineering guardrails that depend on it.
7. Standalone documentation-policy sections SHOULD be removed when their content fits more cleanly into global rules or source-document guidance.

## Anti-Patterns and Quality Checks

### Anti-Patterns

- Duplicate normative rules in multiple sections.
- Conflicting `MUST`/`SHOULD` statements for the same behavior.
- Imperative rule-like statements without RFC keywords (hidden normative requirements).
- Vague requirements without verification criteria.
- Overly long and multi-clause sentences that reduce junior readability.
- Skill descriptions that explain only “what” but not “when”.
- Mixed formatting conventions for equivalent structures in the same governance file (`AGENTS.md` or `SKILL.md`).
- Architecture or dependency sections in `AGENTS.md` that mostly duplicate facts already assigned to source documents.
- Source-document guidance that only lists file names without stating what the agent should use each file for.

### Final Quality Gate

Before finalizing governance output, verify:

1. Scope is limited to `SKILL.md` and/or `AGENTS.md`.
2. Target documents make role/purpose and scope clear near the top and keep a global-to-local structure.
3. Content is in English, junior-readable, and operationally concrete.
4. Normative statements use uppercase RFC keywords with RFC 2119 semantics.
5. Hidden normative imperatives without RFC keywords are removed.
6. Non-normative sections do not introduce requirements.
7. Existing rules were refined first; new rules were added only when necessary.
8. No contradiction or avoidable redundancy remains.
9. Allowed, forbidden, and exception behavior is discoverable, with explicit exception boundaries.
10. Process order is explicit where strict ordering is required.
11. Output and formatting requirements are explicit, concrete, and testable.
12. Formatting conventions are uniform across equivalent structures in each updated governance file.
13. For `AGENTS.md`, heading numbering remains sequential and inbound heading references are updated when headings change.

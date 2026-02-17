# Technical Documentation Template

## Purpose

Use this template for engineering-facing documentation. Technical documentation defines architecture, mechanisms, contracts, decisions, and operational constraints.

Do:

- explain how behavior is implemented
- define boundaries between layers/components
- document core technical mechanisms and engineering decisions
- document cross-cutting technical mechanisms even if they are not tied to one specific user flow

Do not:

- restate product intent and UX narrative in full
- duplicate product-scope definitions that already exist

## Writing Standard

- Write for engineers and maintainers.
- Keep content practical, implementation-oriented, and code-aligned.
- Link to deep-dive references instead of repeating long theory.
- Describe the current state only, not development process.
- Keep the document updated with every technical-impacting code change.

## Mandatory Markdown Structure (Strict)

```markdown
# Technical Documentation Title

## 1. Table of Contents

## 2. Technical Overview
- High-level technical intent and context.

## 3. Architecture and Boundaries
- Layering, dependency direction, and invariants.
- Boundary responsibilities.

## 4. Components and Responsibilities
- Package/module/component ownership.
- Responsibilities and collaboration model.

## 5. Core Technical Mechanisms
- Key implementation mechanisms and patterns.
- Cross-cutting concerns (for example state handling, orchestration, safety mechanisms).

## 6. Data and Interface Contracts
- Public/internal interfaces and integration contracts.
- Key payload or schema constraints.

## 7. Runtime and Operational Considerations
- Runtime behavior, lifecycle, and operational characteristics.
- Execution constraints relevant for maintainers.

## 8. Technical Decisions and Tradeoffs
- Decision records and rationale.

## 9. Technology Stack and Versions
- Runtime dependencies and versions relevant to current behavior.

## 10. Technical Constraints and Risks
- Current limits and known tradeoffs.

## 11. Deep-Dive References
- Architecture deep dives.
- Testing deep dives.

## 12. Cross-References to Product Documentation
- Link to behavior/scope source sections.
```

Structure rule:

- Keep all sections in the exact order shown above.
- Do not remove, merge, or reorder sections.
- If a section has no applicable content, write `Not applicable in current state.`.

## Section Ownership Boundaries

- Keep product rationale short and link to product documentation for user-value context.
- Keep implementation facts canonical in technical documentation.
- Do not force one-to-one mapping to product sections; structure by technical concerns when that improves maintainability.
- Do not include development workflow instructions.

## Cross-Reference Pattern

Use explicit links instead of duplication:

- `See product documentation: {path}#{section-anchor}`

Replace placeholders with concrete paths and anchors in the target project.

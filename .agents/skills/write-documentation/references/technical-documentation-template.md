# Technical Documentation Template

## Purpose

Use this template for engineering-facing documentation. Technical documentation defines architecture, runtime behavior, interfaces, and constraints.

Do:

- explain how behavior is implemented
- define boundaries between layers/components
- document runtime flow, decisions, and test strategy

Do not:

- restate product intent and UX narrative in full
- duplicate product-scope definitions that already exist

## Writing Standard

- Write for engineers and maintainers.
- Keep content practical, implementation-oriented, and code-aligned.
- Link to deep-dive references instead of repeating long theory.
- Describe the current state only, not development process.
- Keep the document updated with every technical-impacting code change.

## Recommended Markdown Structure

```markdown
# Technical Documentation Title

## Document Control
- Owner:
- Status:
- Last updated:
- Related product doc:

## Table of Contents

## 1. Technical Overview
- High-level system intent and context.

## 2. Project Structure
- Folder/package/module responsibilities.

## 3. Architecture Guidelines
- Dependency direction.
- Boundaries and invariants.

## 4. Runtime Flow
- Startup flow.
- Read/query flow.
- Write/mutation flow.

## 5. Data and Interface Contracts
- Public/internal interfaces and integration contracts.
- Key payload or schema constraints.

## 6. Technical Decisions and Tradeoffs
- Decision records and rationale.

## 7. Technology Stack and Versions
- Runtime dependencies and versions relevant to current behavior.

## 8. Technical Constraints and Risks
- Current limits and known tradeoffs.

## 9. Deep-Dive References
- Architecture deep dives.
- Testing deep dives.

## 10. Cross-References to Product Documentation
- Link to behavior/scope source sections.
```

## Section Ownership Boundaries

- Keep product rationale short and link to product documentation for user-value context.
- Keep implementation facts canonical in technical documentation.
- Do not include development workflow instructions.

## Cross-Reference Pattern

Use explicit links instead of duplication:

- `See product documentation: {path}#{section-anchor}`

Replace placeholders with concrete paths and anchors in the target project.

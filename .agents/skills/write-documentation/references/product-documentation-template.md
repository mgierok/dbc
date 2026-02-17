# Product Documentation Template

## Purpose

Use this template for product-facing documentation. Product documentation defines available capabilities, user-visible behavior, constraints, and user flows.

Do:

- describe available capabilities and user-visible behavior
- describe user workflows and interaction expectations
- define constraints and non-goals from the user perspective

Do not:

- duplicate deep implementation details
- describe architecture internals that belong to technical documentation

## Writing Standard

- Write in plain language.
- Optimize for product stakeholders and junior engineers.
- Prefer explicit statements over implicit assumptions.
- Describe the current state only, not planned implementation work.
- Keep the document updated with every product-impacting code change.

## Mandatory Markdown Structure (Strict)

```markdown
# Product Documentation Title

## 1. Table of Contents

## 2. Product Overview
- What the application does for users.
- Core user value in current state.

## 3. Available Capabilities
- What users can do today.
- Capability boundaries in current state.

## 4. Functional Behavior
- User-visible behavior rules.
- States, outcomes, and error behavior visible to users.

## 5. User Flows
- Main user journeys.
- Alternative/error journeys when relevant.

## 6. Interaction Model
- Primary interaction patterns and controls.
- Navigation and state transitions visible to users.

## 7. Constraints and Non-Goals
- User-visible limits.
- Explicit non-goals.

## 8. Safety and Governance
- User-facing safety rules and confirmations.
- Behavioral guardrails that affect user actions.

## 9. Glossary
- Shared user/product terms.

## 10. Cross-References to Technical Documentation
- Link to relevant implementation sections.
```

Structure rule:

- Keep all sections in the exact order shown above.
- Do not remove, merge, or reorder sections.
- If a section has no applicable content, write `Not applicable in current state.`.

## Section Ownership Boundaries

- Keep implementation details out unless they change product behavior.
- If implementation detail is needed for context, summarize in one sentence and link to technical documentation.
- Do not expect one-to-one section mapping with technical documentation; the technical document may organize content by mechanisms and architecture rather than user flows.
- Do not include development workflow instructions.

## Cross-Reference Pattern

Use explicit links instead of duplication:

- `See technical documentation: {path}#{section-anchor}`

Replace placeholders with concrete paths and anchors in the target project.

# Product Documentation Template

## Purpose

Use this template for product-facing documentation. Product documentation defines user value, behavior, scope, and constraints.

Do:

- explain what the product does and why it exists
- describe user workflows and expected behavior
- define in-scope and out-of-scope boundaries

Do not:

- duplicate deep implementation details
- describe architecture internals that belong to technical documentation

## Writing Standard

- Write in plain language.
- Optimize for product stakeholders and junior engineers.
- Prefer explicit statements over implicit assumptions.
- Describe the current state only, not planned implementation work.
- Keep the document updated with every product-impacting code change.

## Recommended Markdown Structure

```markdown
# Product Documentation Title

## Document Control
- Owner:
- Status:
- Last updated:
- Related technical doc:

## Table of Contents

## 1. Product Summary
- What the product capability is.
- Why it matters.

## 2. Problem Statement and Value Proposition
- User problem.
- Product value.

## 3. Target Users and Core Jobs
- User segments.
- Jobs-to-be-done.

## 4. Current Product Scope
### In Scope
### Out of Scope

## 5. Experience Principles
- UX rules and interaction expectations.

## 6. End-to-End User Journey
- Main workflow steps.
- Alternate/error paths where relevant.

## 7. Functional Specification
- Capability-by-capability behavior.
- Input/output expectations.
- User-visible states.

## 8. Interaction Model
- Key actions, shortcuts, UI state transitions.

## 9. Data Safety and Governance
- User-facing safety constraints and confirmations.

## 10. Constraints and Non-Goals
- Current limits.
- Explicit non-goals.

## 11. Glossary
- Shared terms and definitions.

## 12. Cross-References to Technical Documentation
- Link to relevant implementation sections.
```

## Section Ownership Boundaries

- Keep implementation details out unless they change product behavior.
- If implementation detail is needed for context, summarize in one sentence and link to technical documentation.
- Do not include development workflow instructions.

## Cross-Reference Pattern

Use explicit links instead of duplication:

- `See technical documentation: {path}#{section-anchor}`

Replace placeholders with concrete paths and anchors in the target project.

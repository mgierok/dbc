# Product Documentation Rules

## Target

- Target file: `docs/product-documentation.md`
- Related template: `references/product-documentation-template.md`

## Scope

Document product-facing facts only:

- business rules and capability boundaries,
- user-visible behavior and outcomes,
- user flows and interaction model,
- constraints and non-goals.

Do not document implementation mechanics unless they directly affect user-visible behavior.

## Allowed Topics (Mandatory Allowlist)

Each proposed product-documentation change MUST map to at least one topic below:

- Product purpose, audience, and user value proposition.
- Problem statement and user needs addressed by the product.
- Capability scope (in-scope and out-of-scope features).
- User-visible feature descriptions and expected outcomes.
- User-visible workflows, journeys, and task sequences.
- User-visible navigation and interaction model.
- User-visible controls, actions, and shortcuts.
- User-visible commands and command outcomes.
- User-visible screen/panel/view behavior.
- User-visible dialogs, popups, and confirmation choices.
- User-visible states and state transitions.
- User-visible status indicators and visual cues.
- User-visible validation rules and error messages.
- User-visible save/discard/cancel semantics.
- User-visible filtering, sorting, and search behavior.
- User-visible configuration and preference behavior.
- User-visible security/safety guards and confirmations.
- User-visible permissions/roles behavior (if applicable).
- User-visible data-handling expectations (for example, staging, drafts, retries).
- User-visible constraints, limitations, and non-goals.
- User-visible accessibility and usability behavior that is factually implemented.
- User-visible platform/environment support statements.
- User-visible output/result formatting rules.
- User-visible recovery behavior after failures.
- User-facing terminology and glossary entries.

## Out-of-Allowlist Handling (Mandatory)

If a proposed change does not map to any allowed topic:

- skip the change item, or
- adapt it to the nearest allowed topic while preserving factual meaning and without adding scope.

Never introduce product-documentation content outside this allowlist.

## Topic Mapping Output (Mandatory)

When adding or changing product-documentation information, explicitly indicate which allowed topic(s) the change maps to.

Minimum requirement:

- provide mapping per applied change item,
- use exact allowed-topic labels from this file,
- include this mapping in the completion output.

## Writing Rules

- Write in plain, explicit language.
- Describe current state only.
- Keep terminology consistent with product-facing vocabulary.
- Do not describe development workflow (branching, PR flow, delivery process).
- Preserve existing section numbering and anchors unless structural migration is explicitly requested.

## Consistency and Integrity Contract (Mandatory)

- Treat this rules file as the parent normative document for product documentation writing.
- Treat `references/product-documentation-template.md` as the structural contract for the product document.
- Keep the generated/updated product document structurally aligned with the template section set and order.
- If this rules file changes structural requirements, update the related template in the same change set.
- If the related template changes structure, update this rules file in the same change set.
- Do not leave contradictions between this rules file and the related template.

# Product Documentation Rules

## Target

- Target file: `docs/product-documentation.md`
- Related template: `references/product-documentation-template.md`

## Scope

Document product-facing facts only:

- available capabilities,
- user-visible behavior,
- user flows and interaction model,
- constraints and non-goals.

Do not document architecture internals unless they directly affect user-visible behavior.

## Writing Rules

- Write in plain, explicit language.
- Describe current state only.
- Keep terminology consistent with technical documentation.
- Do not describe development workflow (branching, PR flow, delivery process).
- Preserve existing section numbering and anchors unless structural migration is explicitly requested.

## Consistency and Integrity Contract (Mandatory)

- Treat this rules file as the parent normative document for product documentation writing.
- Treat `references/product-documentation-template.md` as the structural contract for the product document.
- Keep the generated/updated product document structurally aligned with the template section set and order.
- If this rules file changes structural requirements, update the related template in the same change set.
- If the related template changes structure, update this rules file in the same change set.
- Do not leave contradictions between this rules file and the related template.

# Technical Documentation Rules

## Target

- Target file: `docs/technical-documentation.md`
- Related template: `references/technical-documentation-template.md`

## Scope

Document implementation-facing facts only:

- architecture and boundaries,
- components and responsibilities,
- technical mechanisms and contracts,
- runtime/operational constraints and risks.

Do not define or restate business rules; keep focus on technical mechanisms and implementation behavior.

## Writing Rules

- Write for engineering and maintenance audiences.
- Describe current factual implementation state only.
- Keep content practical, code-aligned, and free from repeated product-level behavior narratives.
- Do not describe development workflow (branching, PR flow, delivery process).
- Preserve existing section numbering and anchors unless structural migration is explicitly requested.

## Consistency and Integrity Contract (Mandatory)

- Treat this rules file as the parent normative document for technical documentation writing.
- Treat `references/technical-documentation-template.md` as the structural contract for the technical document.
- Keep the generated/updated technical document structurally aligned with the template section set and order.
- If this rules file changes structural requirements, update the related template in the same change set.
- If the related template changes structure, update this rules file in the same change set.
- Do not leave contradictions between this rules file and the related template.

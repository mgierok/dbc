---
name: write-documentation
description: Create and update product documentation and technical documentation in structured Markdown with a clear split of responsibilities, explicit cross-references, and no unnecessary duplication. Use when asked to write, revise, standardize, or reorganize product docs, technical docs, doc templates, documentation rules, or consistency between product and technical documentation. Trigger on requests like "update docs", "write product documentation", "write technical documentation", "create documentation template", "zaktualizuj dokumentacje", "uzupelnij dokumentacje produktowa", or "uzupelnij dokumentacje techniczna".
---

# Write Documentation

## Goal

Produce clear, structured, and complementary product and technical documentation in Markdown.

## Critical Rules

- Separate intent first:
  - Product documentation explains what and why from a product perspective.
  - Technical documentation explains how from an implementation perspective.
- Treat product and technical documentation as different artifacts with different goals:
  - product doc focuses on user-visible behavior, available capabilities, constraints, and user flows,
  - technical doc focuses on architecture, mechanisms, contracts, and engineering decisions (including cross-cutting topics not tied to one specific user flow).
- Document only the current, factual application state.
- Update impacted documentation in the same change set as every codebase change.
- Keep content complementary, not duplicated; use explicit cross-references instead of repeating the same requirement text.
- Treat template structures as strict contracts:
  - keep required section headings and order exactly as defined in the template,
  - do not remove, merge, or reorder required sections,
  - if a required section has no current content, write `Not applicable in current state.`.
- Use a structured Markdown format with numbered H2 sections and stable heading hierarchy.
- Default documentation root to `docs/`.
- Allow location override when the user or project defines a different documentation root.
- Keep wording explicit, plain, and audience-appropriate:
  - Product docs: understandable for product and junior engineering audiences.
  - Technical docs: practical and implementation-oriented for engineering audiences.
- Do not define or describe development flow (for example contributor workflow, feature delivery steps, branch strategy, PR process).
- If scope is ambiguous and affects behavior or interfaces, ask one focused question with exactly three options.

## Workflow

1. Classify the request:
   - `product`, `technical`, or `both`.
2. Resolve documentation root in this order:
   - explicit user path,
   - documented project convention,
   - existing documentation location,
   - fallback `docs/`.
3. Load only required context:
   - existing target documents,
   - style/policy files relevant to documentation governance,
   - `references/product-documentation-template.md` for product changes,
   - `references/technical-documentation-template.md` for technical changes.
4. Draft or update only the sections impacted by the request.
5. Add targeted cross-references instead of duplicated text:
   - product -> technical when implementation detail is needed for context,
   - technical -> product when user-facing behavior/scope context is needed.
   - do not require direct mapping of every product flow/feature to one technical section.
6. Enforce strict template structure:
   - verify headings/order match the selected template exactly,
   - fill non-applicable sections with `Not applicable in current state.` instead of deleting them.
7. Ensure no development-flow guidance is introduced.
8. Run `references/documentation-coherence-checklist.md` before finalizing.
9. Return a short completion report with:
   - `CHANGES MADE`
   - `THINGS NOT TOUCHED`
   - `RISKS / VERIFY`

## Writing Rules by Document Type

- Product documentation:
  - Prioritize behavior, user value, scope, UX flows, and constraints.
  - Avoid deep implementation details unless they directly affect product behavior.
- Technical documentation:
  - Prioritize architecture, runtime flow, interfaces, technical decisions, and constraints.
  - Document cross-cutting technical mechanisms even when they are not directly tied to a single user flow or feature section.
  - Avoid restating product rationale; reference product sections where needed.

## Output Rules

- Preserve existing section numbering and anchors when updating existing files.
- Prefer additive, localized edits over full rewrites.
- If required section data is unknown, write `Not applicable in current state.` and call it out in `RISKS / VERIFY`.
- Keep terminology consistent across product and technical documents.
- Follow strict template section order for all template-governed documentation updates.
- Do not add sections that describe development flow.

## Troubleshooting

- Missing counterpart document:
  - Create a minimal scaffold from the relevant template and add cross-reference placeholders.
- Conflicting statements between product and technical docs:
  - Treat product document as source for behavior intent and technical document as source for implementation details.
  - Resolve by editing both sides or linking to one canonical section.
- Duplicate text across documents:
  - Keep the most appropriate canonical version and replace duplication with concise context plus explicit links.

## References

- Use `references/product-documentation-template.md` for product structure and writing boundaries.
- Use `references/technical-documentation-template.md` for technical structure and writing boundaries.
- Use `references/documentation-coherence-checklist.md` to verify consistency and complementarity.

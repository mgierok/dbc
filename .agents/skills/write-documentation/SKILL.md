---
name: write-documentation
description: Create and update only `docs/product-documentation.md` and `docs/technical-documentation.md` in structured Markdown with a clear split of responsibilities, explicit cross-references, and no unnecessary duplication. Use only when requests create or modify these two files, or review consistency between them.
---

# Write Documentation

## Goal

Produce clear, structured, and complementary documentation in Markdown only for `docs/product-documentation.md` and `docs/technical-documentation.md`.

## Critical Rules

- Allowed targets are strict and exclusive:
  - `docs/product-documentation.md`
  - `docs/technical-documentation.md`
  - Do not use this skill for `README.md` or any other file.
- Separate intent first:
  - Product documentation explains what and why from a product perspective.
  - Technical documentation explains how from an implementation perspective.
- Treat product and technical documentation as different artifacts with different goals:
  - product doc focuses on user-visible behavior, available capabilities, constraints, and user flows,
  - technical doc focuses on architecture, mechanisms, contracts, and engineering decisions (including cross-cutting topics not tied to one specific user flow).
- Document only the current, factual application state.
- If documentation conflicts with current code behavior, treat code as factual source and update documentation accordingly in the same change set.
- Update impacted documentation in the same change set as every codebase change.
- Keep content complementary, not duplicated; use explicit cross-references instead of repeating the same requirement text.
- Prevent instruction redundancy inside documentation specs:
  - define each normative rule once in its canonical section and reference it elsewhere,
  - avoid repeating the same constraint across `Critical Rules`, `Workflow`, `Output Rules`, and `Troubleshooting`,
  - when updating wording, remove stale duplicates in the same change set.
- Treat template structures as strict contracts:
  - keep required section headings and order exactly as defined in the template,
  - do not remove, merge, or reorder required sections,
  - if a required section has no current content, write `Not applicable in current state.`.
- Use a structured Markdown format with numbered H2 sections and stable heading hierarchy.
- Keep wording explicit, plain, and audience-appropriate:
  - Product docs: understandable for product and junior engineering audiences.
  - Technical docs: practical and implementation-oriented for engineering audiences.
- Do not define or describe development flow (for example contributor workflow, feature delivery steps, branch strategy, PR process).
- If scope is ambiguous and affects behavior or interfaces, ask one focused question; when options help, provide 2-3 options with clear tradeoffs.

## Workflow

1. Classify the request:
   - `product`, `technical`, or `both`.
2. Resolve target files from the strict allowlist:
   - product target: `docs/product-documentation.md`,
   - technical target: `docs/technical-documentation.md`.
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
8. Validate cross-reference links and anchors changed in this task:
   - ensure link targets exist in current target documents,
   - fix stale anchors introduced by section title changes.
9. Run `references/documentation-coherence-checklist.md` before finalizing.
10. Return a short completion report with:
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
- For existing files, keep current numbering/anchors unless an explicit migration task requests structural alignment to the template.
- Prefer additive, localized edits over full rewrites.
- If required section data is unknown, write `Not applicable in current state.` and call it out in `RISKS / VERIFY`.
- Keep terminology consistent across product and technical documents.
- Follow strict template section order for all template-governed documentation updates.
- Do not add sections that describe development flow.

## Troubleshooting

- Missing counterpart document:
  - Create only the missing allowlisted file (`docs/product-documentation.md` or `docs/technical-documentation.md`) from the relevant template and add cross-reference placeholders.
- Conflicting statements between product and technical docs:
  - Treat product document as source for behavior intent and technical document as source for implementation details.
  - Resolve by editing both sides or linking to one canonical section.
- Duplicate text across documents:
  - Keep the most appropriate canonical version and replace duplication with concise context plus explicit links.

## References

- Use `references/product-documentation-template.md` for product structure and writing boundaries.
- Use `references/technical-documentation-template.md` for technical structure and writing boundaries.
- Use `references/documentation-coherence-checklist.md` to verify consistency and complementarity.

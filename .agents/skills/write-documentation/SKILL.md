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
- Document only the current, factual application state.
- Update impacted documentation in the same change set as every codebase change.
- Keep content complementary, not duplicated; prefer explicit links to counterpart sections.
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
5. Add explicit cross-references instead of duplicating content:
   - product -> technical for implementation details,
   - technical -> product for behavior/scope intent.
6. Ensure no development-flow guidance is introduced.
7. Run `references/documentation-coherence-checklist.md` before finalizing.
8. Return a short completion report with:
   - `CHANGES MADE`
   - `THINGS NOT TOUCHED`
   - `RISKS / VERIFY`

## Writing Rules by Document Type

- Product documentation:
  - Prioritize behavior, user value, scope, UX flows, and constraints.
  - Avoid deep implementation details unless they directly affect product behavior.
- Technical documentation:
  - Prioritize architecture, runtime flow, interfaces, technical decisions, and constraints.
  - Avoid restating product rationale; reference product sections where needed.

## Output Rules

- Preserve existing section numbering and anchors when updating existing files.
- Prefer additive, localized edits over full rewrites.
- If required section data is unknown, mark as `TBD` and call it out in `RISKS / VERIFY`.
- Keep terminology consistent across product and technical documents.
- Do not add sections that describe development flow.

## Troubleshooting

- Missing counterpart document:
  - Create a minimal scaffold from the relevant template and add cross-reference placeholders.
- Conflicting statements between product and technical docs:
  - Treat product document as source for behavior intent and technical document as source for implementation details.
  - Resolve by editing both sides or linking to one canonical section.
- Duplicate text across documents:
  - Keep the most appropriate canonical version and replace duplication with explicit links.

## References

- Use `references/product-documentation-template.md` for product structure and writing boundaries.
- Use `references/technical-documentation-template.md` for technical structure and writing boundaries.
- Use `references/documentation-coherence-checklist.md` to verify consistency and complementarity.

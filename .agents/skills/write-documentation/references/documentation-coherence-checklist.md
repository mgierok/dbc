# Documentation Coherence Checklist

Use this checklist before finalizing product and/or technical documentation updates.

## 1. Purpose Separation

- Product document answers what and why.
- Technical document answers how.
- No section mixes both perspectives without clear reason.

## 2. Scope Alignment

- Product in-scope and out-of-scope statements are reflected in technical constraints.
- Technical implementation does not imply unsupported product capabilities.

## 3. Terminology Consistency

- Shared terms mean the same thing across both documents.
- Glossary terms match names used in architecture/runtime sections.

## 4. Non-Duplication

- No large duplicated paragraphs across product and technical docs.
- Repeated content is replaced by explicit cross-reference links.

## 5. Traceability

- Each major product capability maps to at least one technical implementation section.
- Each major technical flow maps back to a product behavior or scope section.

## 6. Change Integrity

- Updated sections have coherent metadata (status/updated date if used by project convention).
- Existing anchors and section numbering remain stable unless intentionally changed.
- Documentation reflects current code behavior after this change set.

## 7. No Development Flow Content

- Product and technical documents do not describe development workflow.
- No contributor process instructions (for example PR flow, branching, delivery steps) are present.

## 8. Quality Gate

- Markdown hierarchy is valid and readable.
- Unknowns are marked as `TBD` with follow-up notes in the completion report.
- Completion report includes:
  - `CHANGES MADE`
  - `THINGS NOT TOUCHED`
  - `RISKS / VERIFY`

## Optional Traceability Matrix Template

```markdown
| Product Section | Technical Section | Relationship | Notes |
| --- | --- | --- | --- |
| {path}#{product-anchor} | {path}#{technical-anchor} | Implements / Constrains / Depends on | ... |
```

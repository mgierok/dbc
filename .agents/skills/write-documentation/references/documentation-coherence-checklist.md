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

- Traceability links are present where they improve understanding and navigation.
- Do not force one-to-one mapping between product flows/features and technical sections.
- Technical documentation may contain cross-cutting sections that do not map directly to a single product flow.
- Product documentation remains centered on capabilities/behavior/flows; technical documentation remains centered on architecture/mechanisms/contracts.

## 6. Change Integrity

- Existing anchors and section numbering remain stable unless intentionally changed.
- Documentation reflects current code behavior after this change set.
- Section headings and order match the strict template contract.

## 7. No Development Flow Content

- Product and technical documents do not describe development workflow.
- No contributor process instructions (for example PR flow, branching, delivery steps) are present.

## 8. Quality Gate

- Markdown hierarchy is valid and readable.
- Unknown or non-applicable required sections are explicitly filled with `Not applicable in current state.` and noted in the completion report.
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

Use this matrix only when a direct mapping is helpful for the current change; it is not a mandatory artifact.

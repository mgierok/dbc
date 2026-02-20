# Documentation Coherence Checklist

Use this checklist before finalizing product and/or technical documentation updates.

## 1. Purpose Separation

- Product document is the canonical source for business rules and user-visible behavior.
- Technical document is the canonical source for implementation details and runtime mechanics.
- Technical sections do not redefine product business rules.

## 2. Scope Alignment

- Product in-scope and out-of-scope statements are reflected in technical constraints.
- Technical implementation does not imply unsupported product capabilities.

## 3. Terminology Consistency

- Shared terms mean the same thing across both documents.
- Glossary terms match names used in architecture/runtime sections.

## 4. Non-Duplication

- No large duplicated paragraphs across product and technical docs.
- Repeated content is rewritten to keep one perspective-specific description per document.

## 5. Traceability

- Traceability is clear from section structure and terminology, without mandatory cross-document links.
- Do not force one-to-one mapping between product flows/features and technical sections.
- Technical documentation may contain cross-cutting sections that do not map directly to a single product flow.
- Product documentation remains centered on capabilities/behavior/flows; technical documentation remains centered on architecture/mechanisms/contracts.

## 6. Change Integrity

- Existing anchors and section numbering remain stable unless intentionally changed.
- Documentation reflects current code behavior after this change set.

## 7. No Development Flow Content

- Product and technical documents do not describe development workflow.
- No contributor process instructions (for example PR flow, branching, delivery steps) are present.

## 8. Quality Gate

- Markdown hierarchy is valid and readable.

## Optional Traceability Matrix Template

Traceability matrix is intentionally omitted in this skill variant to avoid introducing cross-document link requirements.

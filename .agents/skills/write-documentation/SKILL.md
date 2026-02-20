---
name: write-documentation
description: Create and update only `docs/product-documentation.md` and `docs/technical-documentation.md` in structured Markdown with a clear split of responsibilities and no unnecessary duplication. Use only when requests create or modify these two files, or review consistency between them.
---

# Write Documentation (Orchestrator)

## Role

Orchestrate documentation tasks for two allowlisted targets only:

- `docs/product-documentation.md`
- `docs/technical-documentation.md`

Do not use this skill for `README.md` or any other file.

## Progressive Disclosure

Load only references required for the selected pass:

- Product pass:
  - `references/product-documentation-rules.md`
  - `references/product-documentation-template.md`
- Technical pass:
  - `references/technical-documentation-rules.md`
  - `references/technical-documentation-template.md`
- Always before finalizing:
  - `references/documentation-coherence-checklist.md`

## Orchestration Workflow

1. Classify request scope as `product`, `technical`, or `both`.
2. Resolve target files from the allowlist only.
3. Execute exactly one pass at a time:
   - If scope is `product`, run Product Pass only.
   - If scope is `technical`, run Technical Pass only.
   - If scope is `both`, run Product Pass first, then Technical Pass.
4. Never update product and technical documentation in the same execution step.
5. Run `references/documentation-coherence-checklist.md` after all requested passes are complete.
6. Return completion report with:
   - `CHANGES MADE`
   - `THINGS NOT TOUCHED`
   - `RISKS / VERIFY`

## Product Pass

1. Load:
   - `references/product-documentation-rules.md`
   - `references/product-documentation-template.md`
2. Update `docs/product-documentation.md` only.
3. Follow the rules file as normative guidance and use the template as structural contract.

## Technical Pass

1. Load:
   - `references/technical-documentation-rules.md`
   - `references/technical-documentation-template.md`
2. Update `docs/technical-documentation.md` only.
3. Follow the rules file as normative guidance and use the template as structural contract.

## Guardrails

- Document only current factual state for the selected perspective: product pass covers business rules/user-visible behavior, technical pass covers implementation mechanics.
- If docs conflict with code behavior, update docs to match code in the same change set.
- Do not add development workflow/process guidance.
- Preserve existing numbering/anchors unless structural migration is explicitly requested.
- If ambiguity affects behavior, interfaces, or architecture boundaries, ask one focused clarification question before writing.

## Reference Index

- `references/product-documentation-rules.md`
- `references/technical-documentation-rules.md`
- `references/product-documentation-template.md`
- `references/technical-documentation-template.md`
- `references/documentation-coherence-checklist.md`

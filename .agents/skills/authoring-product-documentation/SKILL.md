---
name: authoring-product-documentation
description: Create and update only `docs/product-documentation.md` as product-facing factual documentation. Use when requests create or modify product documentation.
---

# Authoring Product Documentation

## Role

Author product documentation for one allowlisted target only:

- `docs/product-documentation.md`

## Required References

Load both files before writing:

- `references/product-documentation-rules.md`
- `references/product-documentation-template.md`

## Workflow

1. Resolve requested change items for `docs/product-documentation.md`.
2. Map every change item to allowed topic labels from `references/product-documentation-rules.md`.
3. If a change item does not map to any allowed topic, skip it or adapt it to the nearest allowed topic without scope expansion.
4. Apply changes in `docs/product-documentation.md` only.
5. Keep structure aligned with `references/product-documentation-template.md`.
6. Return completion report with:
   - `CHANGES MADE`
   - `ALLOWED-TOPIC MAPPING`
   - `THINGS NOT TOUCHED`
   - `RISKS / VERIFY`

## Guardrails

- Document current factual state only.
- Do not modify files other than `docs/product-documentation.md` when this skill is active.
- Preserve existing numbering and anchors unless structural migration is explicitly requested.

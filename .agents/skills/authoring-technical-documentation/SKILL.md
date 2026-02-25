---
name: authoring-technical-documentation
description: Create and update only `docs/technical-documentation.md` as implementation-facing factual documentation. Use when requests create or modify technical documentation.
---

# Authoring Technical Documentation

## Role

Author technical documentation for one allowlisted target only:

- `docs/technical-documentation.md`

## Required References

Load both files before writing:

- `references/technical-documentation-rules.md`
- `references/technical-documentation-template.md`

## Workflow

1. Resolve requested change items for `docs/technical-documentation.md`.
2. Map every change item to allowed topic labels from `references/technical-documentation-rules.md`.
3. If a change item does not map to any allowed topic, skip it or adapt it to the nearest allowed topic without scope expansion.
4. Apply changes in `docs/technical-documentation.md` only.
5. Keep structure aligned with `references/technical-documentation-template.md`.
6. Return completion report with:
   - `CHANGES MADE`
   - `ALLOWED-TOPIC MAPPING`
   - `THINGS NOT TOUCHED`
   - `RISKS / VERIFY`

## Guardrails

- Document current factual state only.
- Do not modify files other than `docs/technical-documentation.md` when this skill is active.
- Preserve existing numbering and anchors unless structural migration is explicitly requested.

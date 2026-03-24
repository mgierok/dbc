# Targeting Rules

## Scope Grammar

- `--scope <n>` resolves the whole line `n`.
- `--scope <start,end>` resolves the inclusive line range from `start` through `end`.
- `--scope <symbol.path>` resolves one symbol from `gopls symbols`.

## Symbol Path Rules

- Symbol paths match the outline path emitted by the helper.
- Nested outline entries use dot paths such as `Worker.Work`.
- Receiver methods accept both raw and normalized spellings.
  Example: `(OldName).Work` and `OldName.Work` resolve the same target.
- `scope_not_found` means no outline symbol matched.
- `ambiguous_scope` means more than one outline symbol matched.

## `--find` Rules

- `--find` searches only inside the selected scope.
- Matching is whitespace-insensitive.
- Matching is case-sensitive.
- `--find` may contain one `<|>` cursor marker.
- More than one cursor marker returns `invalid_cursor_marker`.
- No matches returns `find_not_found`.
- More than one match returns `ambiguous_find`.

## Cursor Resolution

- Without `--find`, the cursor resolves to the start of the selected scope.
- With `--find` and no `<|>`, the cursor resolves to the start of the unique match.
- With `--find` and `<|>`, the cursor resolves to the exact marker position within the unique match.

## Examples

- One line:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py locate --file ./main.go --scope 21`
- Inclusive line range:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py locate --file ./main.go --scope 21,24`
- Symbol path:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py locate --file ./main.go --scope Worker.Work`
- Receiver alias:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py locate --file ./main.go --scope OldName.Work`
- Whitespace-insensitive search:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py locate --file ./main.go --scope 17,18 --find "return worker. Work()"`
- Explicit cursor marker:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py locate --file ./main.go --scope 21 --find "value<|>"`

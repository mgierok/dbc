# gopls Workflows

## Task Recipes

- Find a symbol across the workspace:
  `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py workspace_symbol --workspace <path> --query <text>`
- Jump to the declaration or type definition for a symbol under the cursor:
  `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py definition --file <path> --line <n> --column <n>`
- List all semantic references for a symbol:
  `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py references --file <path> --line <n> --column <n>`
- Find concrete implementations of an interface or method:
  `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py implementation --file <path> --line <n> --column <n>`
- Trace callers and callees for a function:
  `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py call_hierarchy --file <path> --line <n> --column <n>`
- Read semantic diagnostics for one Go file:
  `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py check --file <path>`
- Check whether a rename is legal at a location:
  `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py prepare_rename --file <path> --line <n> --column <n>`
- Preview a rename without modifying files:
  `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py rename_diff --file <path> --line <n> --column <n> --name <newName>`

## Inspection Guidance

- Use `workspace_symbol` first when you know the symbol name but not the file.
- Use `definition`, `references`, `implementation`, or `call_hierarchy` after you have a precise file position.
- After `gopls` identifies the semantic target, inspect the smallest relevant file set manually instead of broad `rg` sweeps.
- When reviewing a possible rename, use `prepare_rename` to verify validity before looking at `rename_diff`.
- Treat `rename_diff` as the safe default preview. Escalate to write-capable operations only if the task explicitly requires applying the rename.

## Troubleshooting

- If the helper reports `workspace_load_failed`, ensure the file is inside a directory tree containing `go.mod` or `go.work`.
- If the helper reports `invalid_position`, re-check the 1-based line and column against the current file contents.
- If `gopls` is missing, install it or ensure it is on `PATH` before retrying.
- If output is unexpectedly sparse, verify the workspace root and inspect whether build tags or generated files change package loading.

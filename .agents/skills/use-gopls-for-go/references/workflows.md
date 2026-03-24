# gopls Workflows

## Task-Driven Recipes

- Search the workspace by symbol name:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py search --workspace <path> --query <text>`
- Show the semantic outline for one Go file:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py outline --file <path>`
- Resolve `--scope` and optional `--find` before deeper commands:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py locate --file <path> --scope <scope> [--find <pattern>]`
- Jump to the definition of the resolved target:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py definition --file <path> --scope <scope> [--find <pattern>]`
- List semantic references:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py references --file <path> --scope <scope> [--find <pattern>]`
- List implementations:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py implementation --file <path> --scope <scope> [--find <pattern>]`
- Trace callers and callees:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py call_hierarchy --file <path> --scope <scope> [--find <pattern>]`
- Read the active call signature:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py signature --file <path> --scope <scope> [--find <pattern>]`
- Show identifier highlights:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py highlight --file <path> --scope <scope> [--find <pattern>]`
- Read semantic diagnostics for one Go file:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py diagnostics --file <path>`
- Check whether a rename is valid:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py rename_validate --file <path> --scope <scope> [--find <pattern>]`
- Preview a rename without editing files:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py rename_preview --file <path> --scope <scope> [--find <pattern>] --name <newName>`

## Recommended Sequences

1. If only the symbol name is known, run `search`.
2. If the file is known but the exact target is not, run `outline`.
3. Run `locate` to confirm the resolved cursor and range before any deeper semantic command.
4. Run `definition`, `references`, `implementation`, `call_hierarchy`, `signature`, or `highlight` from the resolved target.
5. For rename checks, run `rename_validate` first and `rename_preview` second.

## Examples

- Find an interface by name, then inspect its outline:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py search --workspace . --query Worker`
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py outline --file ./internal/worker.go`
- Resolve a method by symbol path and then inspect references:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py locate --file ./internal/worker.go --scope Worker.Run`
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py references --file ./internal/worker.go --scope Worker.Run`
- Resolve a line range with a precise cursor marker before signature lookup:
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py locate --file ./internal/worker.go --scope 40,48 --find "client.<|>Do(req)"`
  `.agents/skills/use-gopls-for-go/scripts/gopls_query.py signature --file ./internal/worker.go --scope 40,48 --find "client.<|>Do(req)"`

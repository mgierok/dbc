---
name: use-gopls-for-go
description: Use gopls for semantic navigation, diagnostics, and safe rename preview in any Go workspace. Use when the task asks to find definitions, references, implementations, call hierarchy, workspace symbols, semantic diagnostics, or to check whether a Go rename is safe before editing.
---

# Use gopls For Go

## Purpose

Use `gopls` as the default semantic tool for Go workspace analysis.
This skill MUST stay repo-agnostic and MUST NOT depend on repository-specific product or architecture knowledge.

## Trigger

- This skill MUST be used for Go tasks where semantic navigation, diagnostics, or refactor preview is needed.
- This skill SHOULD trigger for prompts such as `find all implementations of this interface`, `trace where this function is called`, `show whether a rename is safe before editing`, or `diagnose this Go file semantically`.
- This skill MUST NOT require repository documentation unless the user separately asks for repository-specific interpretation.

## Operating Rules

- The skill MUST prefer `gopls` over plain-text search when the question is semantic rather than lexical.
- The skill MUST prefer the bundled helper `scripts/gopls_query.py` for the supported commands in this skill.
- The skill MUST keep the default workflow read-only or diff-only.
- The skill MUST NOT use write-capable `gopls` operations such as `rename -w` as the default path.
- The skill SHOULD combine `gopls` results with focused manual code inspection after the semantic target has been identified.

## Workflow

1. Determine whether the task is workspace-wide (`workspace_symbol`) or file-position-based (`definition`, `references`, `implementation`, `call_hierarchy`, `check`, `prepare_rename`, `rename_diff`).
2. Run the matching helper command from `scripts/gopls_query.py`.
3. Inspect the normalized JSON output and open only the most relevant files for follow-up reading.
4. If the task is a rename assessment, run `prepare_rename` first and `rename_diff` second.
5. If `gopls` cannot load the workspace, read `references/workflows.md` troubleshooting notes before falling back to lower-signal methods.

## Commands

- `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py workspace_symbol --workspace <path> --query <text>`
- `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py definition --file <path> --line <n> --column <n>`
- `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py references --file <path> --line <n> --column <n>`
- `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py implementation --file <path> --line <n> --column <n>`
- `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py call_hierarchy --file <path> --line <n> --column <n>`
- `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py check --file <path>`
- `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py prepare_rename --file <path> --line <n> --column <n>`
- `python3 .agents/skills/use-gopls-for-go/scripts/gopls_query.py rename_diff --file <path> --line <n> --column <n> --name <newName>`

## References

- Read `references/workflows.md` for task-to-command recipes, inspection guidance, and troubleshooting.

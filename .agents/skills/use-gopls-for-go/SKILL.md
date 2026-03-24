---
name: use-gopls-for-go
description: Use `gopls` and the bundled executable `scripts/gopls_query.py` for Go semantic search, file outline, target resolution, diagnostics, caller/callee tracing, signature lookup, highlight tracing, and safe rename validation or preview. Use when a Go task needs symbol-aware navigation or rename safety that plain-text search cannot provide.
---

# Use gopls For Go

## Purpose and Scope

- This skill MUST use `gopls` as the semantic backend for Go workspace analysis.
- This skill MUST stay repo-agnostic and MUST NOT depend on repository-specific product knowledge.
- This skill MUST keep the default workflow read-only or diff-only.
- This skill MUST NOT perform write-capable refactors in v1.

## Prerequisites

- `gopls` MUST be available on `PATH`.
- The target file SHOULD live under a directory tree that `gopls` can load through `go.mod` or `go.work`.
- The bundled helper SHOULD be invoked through `.agents/skills/use-gopls-for-go/scripts/gopls_query.py`.

## When To Prefer `gopls` Vs Text Search

- This skill MUST prefer `gopls` for definitions, references, implementations, call hierarchy, signatures, highlights, semantic diagnostics, workspace symbol search, and rename safety checks.
- This skill SHOULD prefer plain-text search only for raw string lookup, comments, non-Go content, or fallback investigation after semantic loading fails.
- This skill SHOULD start with `outline` or `locate` when the semantic target is not yet precise enough for a position-based command.

## Targeting Model

- Scope-based commands MUST use `--scope`.
- `--scope <n>` MUST resolve one full line.
- `--scope <start,end>` MUST resolve one inclusive line range.
- `--scope <symbol.path>` MUST resolve through `gopls symbols`.
- Symbol-path matching MUST accept both `Type.Method` and `(Type).Method`.
- `--find` MUST match inside the selected scope with whitespace-insensitive comparison.
- `--find` MAY contain exactly one `<|>` marker to place the semantic cursor precisely.
- If `--find` is omitted, the cursor MUST default to the start of the resolved scope or symbol.
- Detailed rules and examples SHOULD be read from [references/targeting.md](./references/targeting.md).

## Default Workflows

- Unknown file, known symbol name: run `search`, inspect the smallest useful hit set, then switch to `locate`, `definition`, `references`, or `implementation`.
- Known file, unclear target: run `outline`, then `locate` with `--scope` and optional `--find`.
- Caller or callee tracing: run `locate` first if needed, then `call_hierarchy`.
- Signature or highlight inspection: run `locate` first if needed, then `signature` or `highlight`.
- Rename safety: run `rename_validate` first and `rename_preview` second.
- Task-driven command sequences SHOULD be read from [references/workflows.md](./references/workflows.md).

## Output Expectations

- The helper MUST return JSON with `command`, `target`, `result`, and `meta`.
- `target` MUST include original inputs and the resolved cursor or range for scope-based commands.
- `meta` MUST include `workspace` and SHOULD include result counts where relevant.
- When this skill is used, the agent SHOULD report the resolved target, the most relevant semantic locations, any ambiguity, and the next recommended semantic step.

## Failure Handling

- The agent SHOULD run `locate` before deeper semantic commands when target ambiguity is likely.
- `scope_not_found`, `ambiguous_scope`, `find_not_found`, `ambiguous_find`, and `invalid_cursor_marker` MUST be treated as user-targeting problems, not as `gopls` failures.
- `workspace_load_failed` SHOULD trigger workspace verification before any plain-text fallback.
- Troubleshooting details and fallback guidance SHOULD be read from [references/troubleshooting.md](./references/troubleshooting.md).

## Limits and Non-Goals

- This skill MUST NOT use write-capable rename in v1.
- This skill MUST NOT run formatting, imports rewrites, or code actions as part of the default workflow.
- This skill MUST NOT generalize to non-Go languages.
- This skill MAY be combined with repository-specific skills after the semantic target is identified.

## References

- [references/workflows.md](./references/workflows.md)
- [references/targeting.md](./references/targeting.md)
- [references/troubleshooting.md](./references/troubleshooting.md)

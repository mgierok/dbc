# Troubleshooting

## Common Error Codes

- `gopls_not_available`: `gopls` is missing from `PATH`.
- `workspace_load_failed`: `gopls` could not load the module or workspace rooted around the target file.
- `invalid_scope`: the requested line or line range is malformed or outside the file.
- `scope_not_found`: the requested symbol path was not found in the file outline.
- `ambiguous_scope`: the requested symbol path matched more than one outline symbol.
- `find_not_found`: `--find` did not match inside the selected scope.
- `ambiguous_find`: `--find` matched more than one location inside the selected scope.
- `invalid_cursor_marker`: `--find` used more than one `<|>` marker.
- `invalid_position`: `gopls` rejected the resolved cursor.
- `gopls_command_failed`: `gopls` returned a non-classified failure.

## Workspace-Loading Checks

1. Confirm the file sits under a `go.mod` or `go.work` tree that `gopls` can load.
2. Confirm the selected file actually belongs to the intended module or workspace.
3. Re-run `locate` first to verify that the resolver selected the expected cursor and range.

## Fallback Guidance

- If semantic loading fails, the agent SHOULD explain that the fallback has lower signal than `gopls`.
- Plain-text search MAY be used only after the semantic failure is made explicit.
- Rename safety SHOULD stop at `rename_validate` or `rename_preview`; the default fallback MUST NOT apply a write-capable rename.

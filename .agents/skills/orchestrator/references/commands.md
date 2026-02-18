# Reusable Command Hints

Replace placeholders (for example `[prd-id]`, `[task-id]`) with concrete values before execution.
Task specifications reference these hints by `Hint ID` and expected usage context.

### Hint ID: `H001`
Use when: you need a clean list of existing PRD files (excluding task files) before assigning a next ID.

```bash
rg --files .tasks | rg "^\\.tasks/PRD-[0-9]+-" | rg -v -- "-TASK-" | sort -V
```

### Hint ID: `H002`
Use when: calculating the next numeric PRD ID directly from filenames.

```bash
last_id="$(rg --files .tasks | rg "^\\.tasks/PRD-[0-9]+-" | rg -v -- "-TASK-" | sed -E 's#\\.tasks/PRD-([0-9]+)-.*#\\1#' | sort -n | tail -1)"
if [ -z "$last_id" ]; then last_id=0; fi
echo "$((last_id + 1))"
```

### Hint ID: `H003`
Use when: tasks for one PRD must be listed in deterministic order.

```bash
rg --files .tasks | rg "^\\.tasks/PRD-[prd-id]-TASK-[0-9]+-" | sort -V
```

### Hint ID: `H004`
Use when: validating a candidate task `Status` and `blocked-by` metadata quickly.

```bash
rg -n "^- Status:|^- blocked-by:" .tasks/PRD-[prd-id]-TASK-[task-id]-*.md
```

### Hint ID: `H005`
Use when: checking sibling/dependency task statuses for the same PRD.

```bash
rg -n "^- Status:" .tasks/PRD-[prd-id]-TASK-*.md
```

### Hint ID: `H006`
Use when: resolving parent PRD pointer from a selected task file.

```bash
rg -n "^- PRD:" .tasks/PRD-[prd-id]-TASK-[task-id]-*.md
```

### Hint ID: `H007`
Use when: extracting `Completion Summary` from any task file (current task or dependency task).

```bash
rg --multiline --multiline-dotall "^## Completion Summary\\n\\n([\\s\\S]*?)(?:\\n## [^\\n]*|\\z)" .tasks/PRD-[prd-id]-TASK-[task-id]-*.md --replace '$1'
```

### Hint ID: `H008`
Use when: extracting exact `Verification Plan` scope for a selected task.

```bash
rg --multiline --multiline-dotall "^## Verification Plan\\n\\n([\\s\\S]*?)(?:\\n## [^\\n]*|\\z)" .tasks/PRD-[prd-id]-TASK-[task-id]-*.md --replace '$1'
```

### Hint ID: `H009`
Use when: deciding whether parent PRD can be moved to `DONE` by checking sibling `READY` tasks.

```bash
rg -n "^- Status: READY$" .tasks/PRD-[prd-id]-TASK-*.md
```

### Hint ID: `H010`
Use when: selecting the first executable task by ordered `Task ID` and status overview.

```bash
rg -n "^- Task ID:|^- Status:" .tasks/PRD-[prd-id]-TASK-*.md | sort -V
```

# README Writing Guidelines

## 1. Purpose and Audience

- `README.md` MUST be the primary end-user guide for installing, launching, and operating the CLI in everyday use.
- The intended audience SHOULD be technical users/operators (including first-time users) who are comfortable with terminal tooling.

## 2. Writing Style

- Content SHOULD stay concise, task-oriented, and actionable.
- Language SHOULD focus on "what to run" and "what happens."
- Command examples MUST be copy-paste ready.
- Internal jargon and deep implementation narrative SHOULD be minimized.

## 3. Required Content

`README.md` MUST include:

- installation and setup prerequisites
- one primary installation path unless additional paths are explicitly required
- supported database scope
- core startup usage examples (`dbc` and `dbc -d <sqlite-db-path>`)
- a `Keyboard Controls and Commands` section covering keybindings and command-mode commands (for example `:config`)
- license pointer

## 4. Excluded Content

`README.md` MUST NOT include:

- architecture internals, dependency-direction rules, or package-level design details
- standards-heavy normative contracts duplicated from internal docs
- PRD/task lifecycle content, acceptance matrices, or implementation checkpoints
- contributor workflow/process guidance (branching, PR flow, delivery steps)

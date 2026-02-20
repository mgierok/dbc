# AGENTS

## 1. Purpose

This file defines operating instructions for AI coding agents (for example Codex, Claude Code) working in this repository during:

- planning future tasks
- implementing new features
- changing existing behavior
- refactoring and maintenance

Goal: keep product and code changes consistent, safe, and aligned with project standards.

## 2. Scope and Priority

- This file applies to the whole repository (project root level).
- Source of truth split:
  - Product perspective: `docs/product-documentation.md`
  - Technical perspective: `docs/technical-documentation.md`
If any documentation conflicts with current code behavior:

1. Treat current code as factual state.
2. Update the relevant documentation in the same change set.

## 3. Mandatory Context Loading

Before planning or coding, load only relevant sections:

- Product behavior and scope:
  - `docs/product-documentation.md#3-available-capabilities`
  - `docs/product-documentation.md#4-functional-behavior`
  - `docs/product-documentation.md#7-constraints-and-non-goals`
- Technical implementation baseline:
  - `docs/technical-documentation.md#4-components-and-responsibilities`
  - `docs/technical-documentation.md#3-architecture-and-boundaries`
  - `docs/technical-documentation.md#5-core-technical-mechanisms`
  - `docs/technical-documentation.md#55-testing-strategy-and-coverage`
  - `docs/technical-documentation.md#54-technical-interaction-patterns`
- Deep-dive references (when needed):
  - `docs/clean-architecture-ddd.md`
  - `docs/test-driven-development.md`

### 3.1 Mandatory Unit-Test Skill Reference Loading

Explicitly load `.agents/skills/create-unit-tests/references/unit-testing-guide.md` before implementation when at least one of these situations is true:

- creating new unit tests
- modifying existing unit tests
- fixing failing unit tests
- refactoring brittle/flaky unit tests
- reviewing unit-test quality, scope, or structure
- designing unit-test cases for changed behavior

Do not skip this reference for unit-test work, regardless of language, framework, or methodology (`TDD`, `BDD`, or test-after).

### 3.2 Mandatory Commit-Message Skill Invocation

Explicitly invoke skill `write-commit-messages` when at least one of these situations is true:

- user asks to create/propose/write a commit message
- user asks to run `commit` without providing a message
- user asks to improve or validate an existing commit message
- user asks to classify commit type/scope according to Conventional Commits

When this skill is invoked, generate commit messages in Conventional Commits format and use the changed files/diff as primary context.

### 3.3 Mandatory Documentation Skill Invocation

Explicitly invoke skill `write-documentation` when at least one of these situations is true:

- creating `docs/product-documentation.md`
- creating `docs/technical-documentation.md`
- modifying `docs/product-documentation.md`
- modifying `docs/technical-documentation.md`
- reviewing consistency, complementarity, or cross-references between `docs/product-documentation.md` and `docs/technical-documentation.md`
- implementing any codebase change that affects documented behavior, scope, architecture, runtime, interfaces, or constraints reflected in `docs/product-documentation.md` and/or `docs/technical-documentation.md`

When this skill is invoked:

- follow `.agents/skills/write-documentation/SKILL.md` as the primary writing procedure
- limit scope strictly to `docs/product-documentation.md` and `docs/technical-documentation.md`
- document only the current factual application state
- update impacted documentation in the same change set as code changes
- do not define or describe development flow in product or technical documentation
- do not use this skill for `README.md` or any other documentation file outside the two files above

## 4. Agent Workflow Standard

### 4.1 Planning

For each task:

1. Define expected product outcome (what changes for the user).
2. Map affected layers/packages.
3. Identify test impact (new tests or updates).
4. Identify documentation impact (`docs/technical-documentation.md` and/or `docs/product-documentation.md`).

### 4.1.1 Assumptions and Ambiguity Protocol

For any non-trivial task, state assumptions before coding.

Use this format:

`ASSUMPTIONS:`
`1. ...`
`2. ...`
`-> Confirm or correct before I proceed.`

If requirements are ambiguous or inconsistent, follow this flow:

1. Stop.
2. Name the exact conflict.
3. Ask one focused clarifying question or present two options with tradeoffs.
4. Wait for user decision before continuing.

Wait for user decision when ambiguity affects behavior, architecture boundaries, data safety, or public interfaces.
You may proceed without waiting only for low-risk mechanical work (for example naming, formatting, or obvious local cleanup), but still state assumptions.

Quick examples:

- Good: "File A says read-only, file B says write allowed. Which is correct?"
- Bad: silently choosing one interpretation and implementing it.

### 4.2 Implementation

- Keep changes minimal and scoped to task intent.
- Respect architecture boundaries and dependency direction.
- Prefer interface-driven changes through application ports.
- Do not bypass use cases from interface adapters.
- Use TDD as default for behavior changes: add/update the relevant failing test before implementation.
- If any requirement, product behavior, or technical decision is unclear, ask the user before implementing assumptions.

### 4.2.1 Mandatory TDD Execution Rules

For every feature change, bug fix, or behavior-impacting refactor, execute `Red -> Green -> Refactor`:

1. `Red`: add or update a test that fails for the target behavior.
2. `Green`: implement the minimal production change required to pass.
3. `Refactor`: improve code/test structure while keeping tests green.

Additional mandatory rules:

- For bug fixes, write a regression test reproducing the bug before applying the fix.
- Do not weaken assertions just to make failing behavior pass.
- Do not skip the `Red` step unless technically impossible.
- If `Red` is technically impossible (for example missing seam in legacy code), explicitly document why and apply test-after only as a justified exception.

### 4.2.2 Simplicity and Scope Discipline (Mandatory)

Apply minimum-change rules:

- Prefer the simplest solution that satisfies requirements.
- Do not add speculative abstractions, configurability, or extensibility that were not requested.
- When asked to add a new instruction, first verify whether the intent can be covered by extending or generalizing an existing instruction; add a new instruction only when no safe merge is possible.
- Keep changes surgical; every changed line must map directly to task intent.
- Do not refactor adjacent or orthogonal code unless explicitly requested.
- If unrelated issues are discovered, report them separately instead of changing them.

Quick examples:

- Good: edit one use case and its tests for one behavior change.
- Bad: adding new generic helper layers "for future reuse" when only one call site exists.

### 4.2.3 Pushback and Decision Checkpoints

When a proposed direction has a clear technical downside:

- push back directly and explain the concrete risk
- propose a safer or simpler alternative
- proceed with the user's choice after the risk is made explicit

For multi-step tasks, include short checkpoints in this format:

- `STEP`: what will be done now
- `VERIFY`: how success will be checked
- `DECISION`: what needs user confirmation before next step

### 4.3 Verification

- Apply quality gates for code changes:
  - run formatter for changed code (for Go: `gofmt`)
  - run linter for affected scope (for Go: `golangci-lint run`)
  - run tests; for feature/code changes run `go test ./...`
- If tests cannot run, explicitly report why.

### 4.3.1 Goal-Driven Verification

Before coding, define clear success criteria that can be verified:

- bug fix: add failing regression test first, then make it pass
- new behavior: cover happy path, edge cases, and error path
- refactor: verify no behavior change with tests before and after
- optimization: implement obviously-correct baseline first, then optimize while preserving behavior

Avoid vague goals like "make it better" or "improve code quality" without measurable checks.

### 4.3.2 Lint and Safety Prevention Rules

For every code change, apply the following non-negotiable rules:

1. No new lint debt:
   - Do not introduce new lint violations.
   - Before finalizing, run `golangci-lint run ./...` and keep it clean.
2. Resource closing discipline:
   - Do not use unchecked `defer x.Close()` in production code.
   - Handle `Close()` errors explicitly or justify a deliberate ignore.
3. SQL safety:
   - Do not build runtime SQL using unvalidated string interpolation.
   - Use placeholders for values.
   - For dynamic identifiers (table/column names), use strict allowlist and/or safe identifier quoting.
4. Security findings policy:
   - Do not disable security linters globally to silence findings.
   - If an exception is required, apply it locally (`#nosec` / `nolint`) with a concrete inline justification.
   - Every local linter exception requires explicit user approval each time (no blanket pre-approval).
5. Mandatory verification evidence in completion report:
   - Include `golangci-lint run ./...` result.
   - Include `go test ./...` result.
   - List any accepted local exceptions (`#nosec` / `nolint`) with rationale.

### 4.4 Completion

A task is complete when:

- code change is implemented
- required tests are added/updated according to TDD rules (or exception is explicitly documented)
- tests pass (or limitation is explicitly documented)
- impacted documentation is updated
- naming and terminology remain consistent

### 4.4.1 Mandatory Completion Report

After each completed implementation, report:

- `CHANGES MADE`: file-level summary of what changed and why
- `THINGS NOT TOUCHED`: areas intentionally left unchanged
- `RISKS / VERIFY`: potential regressions and additional checks to run

Keep this report short and concrete so a junior engineer can quickly review and validate the result.

### 4.4.2 Mandatory Lessons-Learned Harvest

After each completed implementation, run a short retrospective scan focused on reusable process improvements.

Use these triggers:

- user correction, pushback, or repeated clarification on the same topic
- aborted/rewound turn caused by avoidable execution or communication issue
- failed verification that required rework
- avoidable ambiguity discovered during implementation
- documentation consistency issue discovered during implementation/review

Rules:

1. If at least one trigger occurred, append at least one numbered entry to `lessons-learned.md`.
2. Each entry must be concrete and prevention-oriented, in this format:
   - `When [context/trigger], [required rule/change], so that [expected prevention result].`
3. Do not add vague notes; entries must be actionable and testable in future tasks.
4. If no trigger occurred, state that explicitly in the completion report (`LESSONS LEARNED: no qualifying trigger`).

## 5. Engineering Guardrails

### 5.1 Language and Style

- Use English for identifiers and internal technical documentation.
- Write idiomatic Go.
- Keep functions focused and explicit in error handling.

### 5.2 Architecture

Use `docs/technical-documentation.md#3-architecture-and-boundaries` as primary architecture guide.

Non-negotiable summary:

- Dependencies point inward.
- Domain must stay isolated from outer layers.
- TUI is an adapter (no direct database/business rule implementation).
- Infrastructure implements ports; it does not drive use case logic.

### 5.2.1 Architecture Rule for New Features

When adding functionality:

1. Start from domain model/service changes if behavior changes domain rules.
2. Add/update use case orchestration.
3. Extend port interfaces only when a new boundary is required.
4. Implement infrastructure adapters for new port behavior.
5. Connect UI adapter to use case, not to infrastructure.

### 5.3 Dependencies and Toolchain

- Dependency/toolchain baseline is defined in:
  - `docs/technical-documentation.md#9-technology-stack-and-versions`
  - `go.mod`
- Adding third-party dependencies requires explicit approval.

## 6. Documentation Policy

Documentation creation and modification are skill-governed:

- mandatory procedure and structure: `.agents/skills/write-documentation/SKILL.md` only for `docs/product-documentation.md` and `docs/technical-documentation.md`
- product and technical documentation must stay complementary and consistent via explicit cross-references
- documentation must describe current factual state only, stay aligned with code/runtime behavior, and be updated with every relevant codebase change
- prefer deep-dive references for extended technical theory instead of duplicating long-form explanations in product/technical docs
- documentation maintenance/meta-guidance belongs in `AGENTS.md`, not in `docs/product-documentation.md` or `docs/technical-documentation.md`
- `README.md` must be kept up to date for user-facing CLI basics; when setup, installation, supported database scope, core startup usage, keybindings, or license details change, update `README.md` in the same change set
- documentation must not define or describe development flow
- whenever any file is renamed or moved, update inbound references to that file across the repository in the same change set; exclude completed PRD and TASK artifacts
- whenever Markdown headings are changed (title or numeric prefix), update inbound heading references across the repository in the same change set

### 6.1 README Purpose, Audience, and Writing Rules

- Purpose: `README.md` is the primary end-user guide for installing, launching, and operating the CLI in everyday use.
- Intended audience: technical users/operators of the application (including first-time users) who are generally familiar with terminal and CLI tooling.
- Form of expression:
  - concise, task-oriented, and actionable,
  - plain user language focused on "what to run" and "what happens",
  - command examples that are copy-paste ready,
  - minimal internal jargon and no implementation-deep narrative.
- `README.md` should include:
  - installation and setup prerequisites,
  - one primary installation path unless additional paths are explicitly required,
  - supported database scope,
  - core startup usage examples (`dbc` and `dbc -d <sqlite-db-path>`),
  - a `Keyboard Controls and Commands` section covering keybindings and command-mode commands (for example `:config`),
  - license pointer.
- `README.md` should not include:
  - architecture internals, dependency-direction rules, or package-level design details,
  - standards-heavy normative contracts duplicated from internal docs,
  - PRD/task lifecycle content, acceptance matrices, or implementation checkpoints,
  - contributor workflow/process guidance (branching, PR flow, delivery steps).

## 7. Quick Reference

- Product source of truth: `docs/product-documentation.md`
- Technical source of truth: `docs/technical-documentation.md`
- Architecture deep dive: `docs/clean-architecture-ddd.md`
- TDD deep dive: `docs/test-driven-development.md`
- Lessons learned log: `lessons-learned.md`
- Run/setup basics: `README.md`

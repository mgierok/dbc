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
  - Product perspective: `docs/PRODUCT_DOCUMENTATION.md`
  - Technical perspective: `docs/TECHNICAL_DOCUMENTATION.md`
- `docs/BRD.md` is legacy/planning context, not authoritative for current state.

If any documentation conflicts with current code behavior:

1. Treat current code as factual state.
2. Update the relevant documentation in the same change set.

## 3. Mandatory Context Loading

Before planning or coding, load only relevant sections:

- Product behavior and scope:
  - `docs/PRODUCT_DOCUMENTATION.md#4-current-product-scope`
  - `docs/PRODUCT_DOCUMENTATION.md#7-functional-specification-current-state`
  - `docs/PRODUCT_DOCUMENTATION.md#10-known-constraints-and-non-goals`
- Technical implementation baseline:
  - `docs/TECHNICAL_DOCUMENTATION.md#3-project-structure`
  - `docs/TECHNICAL_DOCUMENTATION.md#4-architecture-guidelines`
  - `docs/TECHNICAL_DOCUMENTATION.md#5-runtime-flow`
  - `docs/TECHNICAL_DOCUMENTATION.md#8-testing-strategy-and-workflow`
  - `docs/TECHNICAL_DOCUMENTATION.md#9-feature-delivery-guide`
- Deep-dive references (when needed):
  - `docs/CLEAN_ARCHITECTURE_DDD.md`
  - `docs/TEST_DRIVEN_DEVELOPMENT.md`

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

## 4. Agent Workflow Standard

### 4.1 Planning

For each task:

1. Define expected product outcome (what changes for the user).
2. Map affected layers/packages.
3. Identify test impact (new tests or updates).
4. Identify documentation impact (`PRODUCT` and/or `TECHNICAL` docs).

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

## 5. Engineering Guardrails

### 5.1 Language and Style

- Use English for identifiers and internal technical documentation.
- Write idiomatic Go.
- Keep functions focused and explicit in error handling.

### 5.2 Architecture

Use `docs/TECHNICAL_DOCUMENTATION.md#4-architecture-guidelines` as primary architecture guide.

Non-negotiable summary:

- Dependencies point inward.
- Domain must stay isolated from outer layers.
- TUI is an adapter (no direct database/business rule implementation).
- Infrastructure implements ports; it does not drive use case logic.

### 5.3 Dependencies and Toolchain

- Dependency/toolchain baseline is defined in:
  - `docs/TECHNICAL_DOCUMENTATION.md#7-technology-stack-and-versions`
  - `go.mod`
- Adding third-party dependencies requires explicit approval.

## 6. Documentation Policy

### 6.1 Product Documentation

`docs/PRODUCT_DOCUMENTATION.md` must be updated for every change affecting:

- product behavior
- feature scope
- user workflows
- UX constraints or shortcuts
- product terminology

Writing standard:

- understandable for Junior Product Manager and Junior Software Engineer
- clear, plain language
- no unnecessary technical internals (except product-level specs, e.g., SQLite support)

### 6.2 Technical Documentation

`docs/TECHNICAL_DOCUMENTATION.md` must be updated for every change affecting:

- architecture and boundaries
- technical decisions
- runtime flow
- dependencies/toolchain versions
- test strategy or engineering workflow

Writing standard:

- understandable for Junior Software Engineer
- practical, implementation-oriented, and code-aligned
- link to deep-dive docs instead of duplicating long conceptual content

## 7. Quick Reference

- Product source of truth: `docs/PRODUCT_DOCUMENTATION.md`
- Technical source of truth: `docs/TECHNICAL_DOCUMENTATION.md`
- Architecture deep dive: `docs/CLEAN_ARCHITECTURE_DDD.md`
- TDD deep dive: `docs/TEST_DRIVEN_DEVELOPMENT.md`
- Run/setup basics: `README.md`

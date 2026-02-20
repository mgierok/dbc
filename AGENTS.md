# AGENTS

## 1. Purpose

This file defines operating instructions for AI coding agents (for example Codex, Claude Code) working in this repository during:

- planning future tasks
- implementing new features
- changing existing behavior
- refactoring and maintenance

Goal: keep product and code changes consistent, safe, and aligned with project standards.

### 1.1 Normative Keywords

- `MUST` / `MUST NOT`: absolute requirement.
- `SHOULD` / `SHOULD NOT`: strong default; deviations require explicit justification.
- `MAY`: optional action based on context.

## 2. Scope and Priority

- This file MUST be treated as applying to the whole repository (project root level).
- Source of truth split:
  - Product perspective: `docs/product-documentation.md`
  - Technical perspective: `docs/technical-documentation.md`
If any documentation conflicts with current code behavior:

1. You MUST treat current code as factual state.
2. You MUST update the relevant documentation in the same change set.

## 3. Mandatory Context Loading

Before planning or coding, the agent MUST load both full source-of-truth documents:

- `docs/product-documentation.md`
- `docs/technical-documentation.md`

Deep-dive references SHOULD be loaded only when task complexity requires normative detail:

- `docs/clean-architecture-ddd.md`:
  - when introducing or changing architecture boundaries,
  - when changing dependency direction or package responsibilities,
  - when adding features that require new ports/adapters or cross-layer orchestration.
- `docs/test-driven-development.md`:
  - when implementing behavior changes (feature/bug fix/refactor with behavior impact),
  - when designing or updating test strategy/coverage,
  - when deciding Red-Green-Refactor execution details for non-trivial changes.

### 3.1 Mandatory Unit-Test Skill Reference Loading

You MUST explicitly load `.agents/skills/create-unit-tests/references/unit-testing-guide.md` before implementation when at least one of these situations is true:

- creating new unit tests
- modifying existing unit tests
- fixing failing unit tests
- refactoring brittle/flaky unit tests
- reviewing unit-test quality, scope, or structure
- designing unit-test cases for changed behavior

You MUST NOT skip this reference for unit-test work, regardless of language, framework, or methodology (`TDD`, `BDD`, or test-after).

### 3.2 Mandatory Commit-Message Skill Invocation

You MUST explicitly invoke skill `write-commit-messages` when at least one of these situations is true:

- user asks to create/propose/write a commit message
- user asks to run `commit` without providing a message
- user asks to improve or validate an existing commit message
- user asks to classify commit type/scope according to Conventional Commits

When this skill is invoked, commit messages MUST use Conventional Commits format and SHOULD use the changed files/diff as primary context.

### 3.3 Mandatory Documentation Skill Invocation

You MUST explicitly invoke skill `write-documentation` when at least one of these situations is true:

- creating `docs/product-documentation.md`
- creating `docs/technical-documentation.md`
- modifying `docs/product-documentation.md`
- modifying `docs/technical-documentation.md`
- reviewing consistency or complementarity between `docs/product-documentation.md` and `docs/technical-documentation.md`
- implementing any codebase change that affects documented behavior, scope, architecture, runtime, interfaces, or constraints reflected in `docs/product-documentation.md` and/or `docs/technical-documentation.md`

When this skill is invoked:

- you MUST follow `.agents/skills/write-documentation/SKILL.md` as the primary writing procedure
- you MUST limit scope strictly to `docs/product-documentation.md` and `docs/technical-documentation.md`
- you MUST document only the current factual application state

## 4. Agent Workflow Standard

### 4.1 Planning

For each task, the agent MUST:

1. Define expected product outcome (what changes for the user).
2. Map affected layers/packages.
3. Identify test impact (new tests or updates).
4. Identify documentation impact (`docs/technical-documentation.md` and/or `docs/product-documentation.md`).

### 4.1.1 Assumptions and Ambiguity Protocol

For any non-trivial task, the agent MUST state assumptions before coding.

Use this format:

`ASSUMPTIONS:`
`1. ...`
`2. ...`
`-> Confirm or correct before I proceed.`

If requirements are ambiguous or inconsistent, the agent MUST follow this flow:

1. Stop.
2. Name the exact conflict.
3. Ask one focused clarifying question or present two options with tradeoffs.
4. Wait for user decision before continuing.

The agent MUST wait for user decision when ambiguity affects behavior, architecture boundaries, data safety, or public interfaces.
The agent MAY proceed without waiting only for low-risk mechanical work (for example naming, formatting, or obvious local cleanup), but MUST still state assumptions.

Quick examples:

- Good: "File A says read-only, file B says write allowed. Which is correct?"
- Bad: silently choosing one interpretation and implementing it.

### 4.2 Implementation

- Changes MUST stay minimal and scoped to task intent.
- The implementation MUST respect architecture boundaries and dependency direction.
- The implementation SHOULD prefer interface-driven changes through application ports.
- Interface adapters MUST NOT bypass use cases.
- If any requirement, product behavior, or technical decision is unclear, the agent MUST ask the user before implementing assumptions.

### 4.2.1 Mandatory TDD Execution Rules

TDD MUST be applied for every feature change, bug fix, and behavior-impacting refactor.

Execution and quality rules for TDD are defined in `docs/test-driven-development.md` and MUST be treated as normative for implementation work in this repository.

Repository-enforced TDD guardrails:

- For bug fixes, you MUST add a regression test that reproduces the bug before applying the fix.
- You MUST NOT weaken assertions only to make failing behavior pass.
- You MUST NOT skip the `Red` step unless technically impossible; if impossible, you MUST document the reason and treat test-after as an explicit exception.

### 4.2.2 Simplicity and Scope Discipline (Mandatory)

Apply minimum-change rules:

- The implementation SHOULD prefer the simplest solution that satisfies requirements.
- You MUST NOT add speculative abstractions, configurability, or extensibility that were not requested.
- When asked to add or change instructions/rules, you MUST first verify whether the intent can be covered by extending, generalizing, or refactoring an existing instruction; you MUST add a new instruction only when no safe merge is possible; this applies to all governance artifacts, including `AGENTS.md` and `.agents/skills/**` definitions (`SKILL.md` and `references/*`).
- Changes MUST stay surgical; every changed line MUST map directly to task intent.
- You MUST NOT refactor adjacent or orthogonal code unless explicitly requested.
- If unrelated issues are discovered, you MUST report them separately instead of changing them.

Quick examples:

- Good: edit one use case and its tests for one behavior change.
- Bad: adding new generic helper layers "for future reuse" when only one call site exists.

### 4.2.3 Pushback and Decision Checkpoints

When a proposed direction has a clear technical downside:

- you MUST push back directly and explain the concrete risk
- you SHOULD propose a safer or simpler alternative
- you MUST proceed with the user's choice after the risk is made explicit

For multi-step tasks, the agent MUST include short checkpoints in this format:

- `STEP`: what will be done now
- `VERIFY`: how success will be checked
- `DECISION`: what needs user confirmation before next step

### 4.3 Verification

- Apply quality gates for code changes:
  - during iteration, you MAY run formatter/linter/tests for affected scope to speed up feedback
  - before finalizing, you MUST run:
    - formatter for changed Go code (`gofmt`)
    - linter for full repository scope (`golangci-lint run ./...`)
    - tests for full repository scope (`go test ./...`)
- If tests cannot run, you MUST explicitly report why.

### 4.3.1 Goal-Driven Verification

Before coding, the agent SHOULD define clear success criteria that can be verified:

- bug fix: add failing regression test first, then make it pass
- new behavior: cover happy path, edge cases, and error path
- refactor: verify no behavior change with tests before and after
- optimization: implement obviously-correct baseline first, then optimize while preserving behavior

The agent MUST avoid vague goals like "make it better" or "improve code quality" without measurable checks.

### 4.3.2 Lint and Safety Prevention Rules

For every code change, the agent MUST apply the following non-negotiable rules:

1. No new lint debt:
   - You MUST NOT introduce new lint violations.
   - Before finalizing, you MUST run `golangci-lint run ./...` and keep it clean.
2. Resource closing discipline:
   - You MUST NOT use unchecked `defer x.Close()` in production code.
   - You MUST handle `Close()` errors explicitly or justify a deliberate ignore.
3. SQL safety:
   - You MUST NOT build runtime SQL using unvalidated string interpolation.
   - You MUST use placeholders for values.
   - For dynamic identifiers (table/column names), you MUST use strict allowlist and/or safe identifier quoting.
4. Security findings policy:
   - You MUST NOT disable security linters globally to silence findings.
   - If an exception is required, you MUST apply it locally (`#nosec` / `nolint`) with a concrete inline justification.
   - Every local linter exception MUST have explicit user approval each time (no blanket pre-approval).
5. Mandatory verification evidence in completion report:
   - You MUST include `golangci-lint run ./...` result.
   - You MUST include `go test ./...` result.
   - You MUST list any accepted local exceptions (`#nosec` / `nolint`) with rationale.

### 4.4 Completion

A task is complete only when all conditions below are met:

- code change is implemented
- required tests are added/updated according to TDD rules (or exception is explicitly documented)
- tests pass (or limitation is explicitly documented)
- naming and terminology remains consistent

### 4.4.1 Mandatory Completion Report

After each completed implementation, the agent MUST report:

- `CHANGES MADE`: file-level summary of what changed and why
- `THINGS NOT TOUCHED`: areas intentionally left unchanged
- `RISKS / VERIFY`: potential regressions and additional checks to run

The report SHOULD stay short and concrete so a junior engineer can quickly review and validate the result.

### 4.4.2 Mandatory Lessons-Learned Harvest

After each completed implementation, the agent MUST run a short retrospective scan focused on reusable process improvements.

Use these triggers:

- user correction, pushback, or repeated clarification on the same topic
- aborted/rewound turn caused by avoidable execution or communication issue
- failed verification that required rework
- avoidable ambiguity discovered during implementation
- documentation consistency issue discovered during implementation/review

Rules:

1. If at least one trigger occurred, the agent MUST propose at least one numbered lesson entry and MUST ask the user whether each proposed lesson should be saved or skipped before writing to `lessons-learned.md`.
2. Each entry MUST be concrete and prevention-oriented, in this format:
   - `When [context/trigger], [required rule/change], so that [expected prevention result].`
3. The agent SHOULD prefer generalized lessons that apply across similar tasks and SHOULD avoid overly case-specific wording unless specificity is required for prevention value.
4. The agent MUST NOT add vague notes; entries MUST be actionable and testable in future tasks.
5. If no trigger occurred, the agent MUST state that explicitly in the completion report (`LESSONS LEARNED: no qualifying trigger`).

## 5. Engineering Guardrails

### 5.1 Language and Style

- The agent MUST use English for identifiers and internal technical documentation.
- The agent MUST write idiomatic Go.
- The agent SHOULD keep functions focused and explicit in error handling.

### 5.2 Architecture

The agent MUST use `docs/technical-documentation.md#3-architecture-and-boundaries` as the primary architecture guide.

Non-negotiable summary:

- Dependencies MUST point inward.
- Domain MUST stay isolated from outer layers.
- TUI MUST remain an adapter (no direct database/business rule implementation).
- Infrastructure MUST implement ports and MUST NOT drive use case logic.

### 5.2.1 Architecture Rule for New Features

When adding functionality, the agent MUST follow this order:

1. Start from domain model/service changes if behavior changes domain rules.
2. Add/update use case orchestration.
3. Extend port interfaces only when a new boundary is required.
4. Implement infrastructure adapters for new port behavior.
5. Connect UI adapter to use case, not to infrastructure.

### 5.3 Dependencies and Toolchain

- Dependency/toolchain baseline MUST be taken from:
  - `docs/technical-documentation.md#9-technology-stack-and-versions`
  - `go.mod`
- Adding third-party dependencies MUST have explicit approval.

## 6. Documentation Policy

Documentation creation and modification MUST be skill-governed:

- Mandatory procedure and structure MUST come from `.agents/skills/write-documentation/SKILL.md` only for `docs/product-documentation.md` and `docs/technical-documentation.md`.
- Product and technical documentation MUST stay complementary and consistent without mandatory cross-document links.
- Documentation MUST describe current factual state only, MUST stay aligned with code/runtime behavior, and MUST be updated with every relevant codebase change.
- The agent SHOULD prefer deep-dive references for extended technical theory instead of duplicating long-form explanations in product/technical docs.
- Documentation maintenance/meta-guidance MUST stay in `AGENTS.md`, not in `docs/product-documentation.md` or `docs/technical-documentation.md`.
- `README.md` MUST be kept up to date for user-facing CLI basics; when setup, installation, supported database scope, core startup usage, keybindings, or license details change, the agent MUST update `README.md` in the same change set.
- Documentation MUST NOT define or describe development flow.
- Whenever any file is renamed or moved, the agent MUST update inbound references to that file across the repository in the same change set; exclude completed PRD and TASK artifacts.
- Whenever Markdown headings are changed (title or numeric prefix), the agent MUST update inbound heading references across the repository in the same change set.

### 6.1 README Policy

- Detailed README writing rules MUST be taken from `docs/readme-guidelines.md`.
- When README-related conditions from Section 6 apply, the agent MUST update `README.md` in the same change set and MUST follow `docs/readme-guidelines.md`.

## 7. Quick Reference

- Product source of truth: `docs/product-documentation.md`
- Technical source of truth: `docs/technical-documentation.md`
- Architecture deep dive: `docs/clean-architecture-ddd.md`
- TDD deep dive: `docs/test-driven-development.md`
- Lessons learned log: `lessons-learned.md`
- Run/setup basics: `README.md`
- README writing rules: `docs/readme-guidelines.md`

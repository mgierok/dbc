# AGENTS

## 1. Scope and Priority

- This file MUST be treated as applying to the whole repository (project root level).
- Whenever the agent asks the user a question, it MUST present exactly four numbered response options:
  - options `1`, `2`, and `3` MUST be predefined choices
  - option `4` MUST allow the user to provide a custom response
- Source of truth split:
  - Product perspective: `docs/product-documentation.md`
  - Technical perspective: `docs/technical-documentation.md`
If any documentation conflicts with current code behavior:

1. You MUST treat current code as factual state.
2. You MUST resolve the documentation update through the applicable documentation skill(s) in the same change set.

## 2. Mandatory Context Loading

Before planning or coding project changes (for example feature work, bug fixes, refactors, or future project planning), the agent MUST load both full source-of-truth documents:

- `docs/product-documentation.md`
- `docs/technical-documentation.md`

This requirement MUST NOT apply when the task scope is limited to governance-only changes (for example updating `AGENTS.md` or `.agents/skills/**/SKILL.md`) and no project behavior is being changed.

Deep-dive references SHOULD be loaded only when task complexity requires normative detail:

- `docs/clean-architecture-ddd.md`:
  - when introducing or changing architecture boundaries,
  - when changing dependency direction or package responsibilities,
  - when adding features that require new ports/adapters or cross-layer orchestration.
- `docs/test-driven-development.md`:
  - when implementing behavior changes (feature/bug fix/refactor with behavior impact),
  - when designing or updating test strategy/coverage,
  - when deciding Red-Green-Refactor execution details for non-trivial changes.

### 2.1 Mandatory Unit-Test Skill Reference Loading

You MUST explicitly load `.agents/skills/create-unit-tests/references/unit-testing-guide.md` before implementation when at least one of these situations is true:

- creating new unit tests
- modifying existing unit tests
- fixing failing unit tests
- refactoring brittle/flaky unit tests
- reviewing unit-test quality, scope, or structure
- designing unit-test cases for changed behavior

You MUST NOT skip this reference for unit-test work, regardless of language, framework, or methodology (`TDD`, `BDD`, or test-after).

### 2.2 Mandatory Commit-Message Skill Invocation

You MUST explicitly invoke skill `write-commit-messages` when at least one of these situations is true:

- user asks to create/propose/write a commit message
- user asks to run `commit` without providing a message
- user asks to improve or validate an existing commit message
- user asks to classify commit type/scope according to Conventional Commits

When this skill is invoked, commit messages MUST use Conventional Commits format and SHOULD use the changed files/diff as primary context.

## 3. Agent Workflow Standard

### 3.1 Planning

This section applies only to project tasks that can result in project-code changes.
This section MUST NOT be applied to documentation-only or governance-only tasks.

For each in-scope task, the agent MUST execute planning in the following order:

1. Step 1: Intent Understanding
   - The agent MUST NOT treat the user instruction as literal and complete by default.
   - The agent MUST ask focused clarification questions to establish full intent and required context.
   - The agent MUST challenge instructions that appear unusual, inconsistent, risky, or controversial, and MUST explain concrete reasons for doubt.
   - The agent MUST NOT continue to Step 2 when any ambiguity or contradiction remains.
   - The step MUST end with an explicit interpretation artifact.
   - The agent MUST obtain explicit user approval of that artifact before continuing.
2. Step 2: Measurable Success Criteria
   - The agent MUST define measurable success criteria from a project-development perspective.
   - Criteria MUST be verifiable through engineering evidence (for example behavior, tests, quality gates, architecture constraints, delivery artifacts).
   - Business outcome metrics (for example revenue, adoption, installs) MUST NOT be used as success criteria in this step.
3. Step 3: Implementation Planning
   - The agent MUST create a detailed implementation plan that links product intent to technical execution.
   - For each planned change set, the agent MUST describe:
     - product-side value delivered by the change,
     - corresponding technical implementation vision.
   - The plan SHOULD be iterative and split complex work into multiple change sets.
   - Each change set MUST deliver working software.
   - Each change set MUST target the smallest change that increases business value.
   - Each change set MUST be complete for code consistency, tests, and documentation.
   - Each change set MUST end with a commit.
4. Step 4: Plan Verification
   - The agent MUST verify that the full plan achieves the intended goal.
   - The agent MUST verify that the full plan can meet the defined success criteria.
   - If gaps or risks are found, the agent MUST update the plan before implementation starts.

### 3.2 Implementation

- Changes MUST stay minimal and scoped to task intent.
- The implementation MUST respect architecture boundaries and dependency direction.
- The implementation SHOULD prefer interface-driven changes through application ports.
- Interface adapters MUST NOT bypass use cases.
- If any requirement, product behavior, or technical decision is unclear, the agent MUST ask the user before implementing assumptions.

### 3.2.1 Mandatory TDD Execution Rules

TDD MUST be applied for every feature change, bug fix, and behavior-impacting refactor.

Execution and quality rules for TDD are defined in `docs/test-driven-development.md` and MUST be treated as normative for implementation work in this repository.

Repository-enforced TDD guardrails:

- For bug fixes, you MUST add a regression unit test that reproduces the bug before applying the fix.
- You MUST NOT weaken assertions only to make failing behavior pass.
- You MUST NOT skip the `Red` step unless technically impossible; if impossible, you MUST document the reason and treat test-after as an explicit exception.

### 3.2.2 Simplicity and Scope Discipline (Mandatory)

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

### 3.2.3 Pushback and Decision Checkpoints

When a proposed direction has a clear technical downside:

- you MUST push back directly and explain the concrete risk
- you SHOULD propose a safer or simpler alternative
- you MUST proceed with the user's choice after the risk is made explicit

For multi-step tasks, the agent MUST include short checkpoints in this format:

- `STEP`: what will be done now
- `VERIFY`: how success will be checked
- `DECISION`: what needs user confirmation before next step

### 3.2.4 Reference Integrity

- Whenever any file is renamed or moved, the agent MUST update inbound references to that file across the repository in the same change set; exclude completed PRD and TASK artifacts.
- Whenever Markdown headings are changed (title or numeric prefix), the agent MUST update inbound heading references across the repository in the same change set.

### 3.3 Verification

- Apply quality gates for code changes:
  - during iteration, you MAY run formatter/linter/tests for affected scope to speed up feedback
  - before finalizing, you MUST run:
    - formatter for changed Go code (`gofmt`)
    - linter for full repository scope (`golangci-lint run ./...`)
    - tests for full repository scope (`go test ./...`)
- If tests cannot run, you MUST explicitly report why.
- For manual `TC-*` execution and reporting rules (`single test case` and `full test case suite`), use `docs/test-case-execution-reporting-specification.md`.

### 3.3.1 Goal-Driven Verification

Before coding, the agent SHOULD define clear success criteria that can be verified:

- bug fix: add failing regression test first, then make it pass
- new behavior: cover happy path, edge cases, and error path
- refactor: verify no behavior change with tests before and after
- optimization: implement obviously-correct baseline first, then optimize while preserving behavior

The agent MUST avoid vague goals like "make it better" or "improve code quality" without measurable checks.

### 3.3.2 Lint and Safety Prevention Rules

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

### 3.4 Completion

A task is complete only when all conditions below are met:

- code change is implemented
- required tests are added/updated according to TDD rules (or exception is explicitly documented)
- tests pass (or limitation is explicitly documented)
- naming and terminology remains consistent

### 3.4.1 Mandatory Completion Report

After each completed implementation, the agent MUST report:

- `CHANGES MADE`: file-level summary of what changed and why
- `RISKS / VERIFY`: potential regressions and additional checks to run

The report SHOULD stay short and concrete so a junior engineer can quickly review and validate the result.

## 4. Engineering Guardrails

### 4.1 Language and Style

- The agent MUST use English for identifiers and internal technical documentation.
- The agent MUST write idiomatic Go.
- The agent SHOULD keep functions focused and explicit in error handling.

### 4.2 Architecture

The agent MUST use `docs/technical-documentation.md#3-architecture-and-boundaries` as the primary architecture guide.

Non-negotiable summary:

- Dependencies MUST point inward.
- Domain MUST stay isolated from outer layers.
- TUI MUST remain an adapter (no direct database/business rule implementation).
- Infrastructure MUST implement ports and MUST NOT drive use case logic.

### 4.2.1 Architecture Rule for New Features

When adding functionality, the agent MUST follow this order:

1. Start from domain model/service changes if behavior changes domain rules.
2. Add/update use case orchestration.
3. Extend port interfaces only when a new boundary is required.
4. Implement infrastructure adapters for new port behavior.
5. Connect UI adapter to use case, not to infrastructure.

### 4.3 Dependencies and Toolchain

- Dependency/toolchain baseline MUST be taken from:
  - `docs/technical-documentation.md#9-technology-stack-and-versions`
  - `go.mod`
- Adding third-party dependencies MUST have explicit approval.

## 5. Documentation Policy

Documentation creation and modification MUST be skill-governed:

For trigger evaluation, documentation files MUST include Markdown/governance documentation artifacts (for example `docs/**`, `README.md`, `AGENTS.md`, `.agents/skills/**/*.md`).

- If multiple documentation perspectives are affected, the agent MUST invoke all applicable skills independently and apply each skill decision.

### 5.1 Product Documentation Policy

- The agent MUST explicitly invoke skill `authoring-product-documentation` when at least one of these situations is true:
  - task changes at least one non-documentation file in the repository
  - creating `docs/product-documentation.md`
  - modifying `docs/product-documentation.md`
- Product documentation policy is governed exclusively by skill `authoring-product-documentation`; `AGENTS.md` MUST NOT define additional or duplicate product-documentation authoring/decision rules.
- For every change in `docs/product-documentation.md`, the agent MUST verify whether existing test cases require updates and whether new test cases must be added to keep aligned with documented behavior.
- The agent MUST accept the invoked skill decision (`UPDATE_REQUIRED` or `NO_UPDATE_REQUIRED`) and proceed accordingly.

### 5.2 Technical Documentation Policy

- The agent MUST explicitly invoke skill `authoring-technical-documentation` when at least one of these situations is true:
  - task changes at least one non-documentation file in the repository
  - creating `docs/technical-documentation.md`
  - modifying `docs/technical-documentation.md`
- Technical documentation policy is governed exclusively by skill `authoring-technical-documentation`; `AGENTS.md` MUST NOT define additional or duplicate technical-documentation authoring/decision rules.
- The agent MUST accept the invoked skill decision (`UPDATE_REQUIRED` or `NO_UPDATE_REQUIRED`) and proceed accordingly.

### 5.3 README Documentation Policy

- The agent MUST explicitly invoke skill `authoring-readme-file` when at least one of these situations is true:
  - task changes at least one non-documentation file in the repository
  - creating `README.md`
  - modifying `README.md`
- README policy is governed exclusively by skill `authoring-readme-file`; `AGENTS.md` MUST NOT define additional or duplicate README authoring/decision rules.
- The agent MUST accept the invoked skill decision (`UPDATE_REQUIRED` or `NO_UPDATE_REQUIRED`) and proceed accordingly.

## 6. Quick Reference

- Product source of truth: `docs/product-documentation.md`
- Technical source of truth: `docs/technical-documentation.md`
- Architecture deep dive: `docs/clean-architecture-ddd.md`
- TDD deep dive: `docs/test-driven-development.md`
- Run/setup basics: `README.md`

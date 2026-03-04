# AGENTS

## 1. Scope and Priority

- This file MUST be treated as applying to the whole repository (project root level).
- Source of truth split:
  - Product perspective: `docs/product-documentation.md`
  - Technical perspective: `docs/technical-documentation.md`
- If any documentation conflicts with current code behavior, the agent MUST treat current code as factual state.

## 2. Mandatory Context Loading

Before planning or coding project changes (for example feature work, bug fixes, refactors, or future project planning), the agent MUST load both full source-of-truth documents:

- `docs/product-documentation.md`
- `docs/technical-documentation.md`

This requirement MUST NOT apply when the task scope is limited to governance-only changes (for example updating `AGENTS.md` or `.agents/skills/**/SKILL.md`) and no project behavior is being changed.

## 3. Agent Workflow

This section MUST be applied only to project tasks that can result in project-code changes.
This section MUST NOT be applied to documentation-only or governance-only tasks.
The workflow MUST use `update_plan` from start through completion, update statuses after each workflow step, and finish with all steps marked as `completed`.

1. Intent Alignment
2. Planning
3. Implementation
4. Completion

### Change-Set Definition

- A change-set MUST be the smallest independently reviewable implementation increment that delivers one coherent value objective and working software.
- Each change-set MUST have a unique identifier within the full plan.
- A change-set MUST be executable, verifiable, and reversible as one unit.
- A change-set MUST NOT mix unrelated value objectives.
- A change-set MAY contain a mix of code, test, and documentation changes only when those changes are directly connected.
- A change-set MUST target the smallest change that increases business value.
- A change-set MUST be complete for code consistency, tests, and documentation.

### Intent Alignment

Intent Alignment MUST use `update_plan`, update statuses after each workflow step in this phase, and keep exactly one step `in_progress` at a time.

1. Intent Understanding
   - The agent MUST NOT treat the user instruction as literal and complete by default.
   - The agent MUST ask focused clarification questions to establish full intent and required context.
   - The agent MUST challenge instructions that appear unusual, inconsistent, risky, or controversial, and MUST explain concrete reasons for doubt.
   - The agent MUST NOT continue to planning when any ambiguity or contradiction remains.
2. Intent Approval
   - The phase MUST end with an explicit interpretation artifact.
   - The agent MUST obtain explicit user approval of that artifact before starting planning.

### Planning

Planning MUST use `update_plan`, update statuses after each workflow step in this phase, and keep exactly one step `in_progress` at a time.

1. Measurable Success Criteria
   - The agent MUST define clear, measurable success criteria from a project-development perspective.
   - Criteria MUST be verifiable through engineering evidence (for example behavior, tests, quality gates, architecture constraints, or delivery artifacts).
   - The agent MUST avoid vague goals like "make it better" or "improve code quality".
   - Business outcome metrics (for example revenue, adoption, or installs) MUST NOT be used as success criteria in this phase.
   - For a bug fix, success criteria MUST include a regression test that fails before the fix and passes after the fix.
   - For new behavior, success criteria MUST include happy path, edge case, and error path verification.
   - For a behavior-preserving refactor, success criteria MUST include proof that behavior is unchanged.
   - For optimization, success criteria MUST include a correctness baseline first, then optimization evidence with preserved behavior.
2. Implementation Planning
   - The agent MUST create a detailed implementation plan that links product intent to technical execution.
   - For each planned change-set, the plan MUST describe product-side value delivered by the change and corresponding technical implementation vision.
   - The plan MUST present `Technical Scope` as a dedicated section inside each planned change-set.
   - The agent MUST NOT provide one aggregated technical-scope section shared across multiple change-sets.
   - The plan SHOULD be iterative and SHOULD split complex work into multiple change-sets.
   - The agent MUST default to multiple change-sets for non-trivial scope.
   - A single change-set MAY be used only when scope is trivial (for example one tightly scoped behavior in one layer) or when the user explicitly requests one change-set.
   - If a single change-set is chosen, the plan MUST include explicit justification why further decomposition would not improve delivery safety or reviewability.
3. Plan Verification
   - The agent MUST verify that the full plan achieves the intended goal.
   - The agent MUST verify that the full plan can meet the defined success criteria.
   - If gaps or risks are found, the agent MUST update the plan before implementation starts.

### Implementation

Implementation MUST iterate over all approved change-sets in the defined order.
Implementation MUST use `update_plan`, update statuses after each workflow step in this phase, and keep exactly one step `in_progress` at a time.

1. Code and Test Execution
   - For project-code implementation, the agent MUST apply all coding rules from Section 4 (`Engineering Guardrails`).
   - For each change-set, the agent MUST apply TDD according to Section `4.3.2 TDD Rules`.
   - During implementation, the agent MAY run verification tools iteratively for affected scope to speed up feedback.
2. Change-Set Verification
   - Before finalizing a change-set, the agent MUST run all mandatory verification commands defined in Section 4.
   - If mandatory tests cannot run, the agent MUST explicitly report why.
3. Documentation and Test Cases
   - If a change-set modifies at least one non-documentation file in the repository, the agent MUST invoke one after another:
     - `authoring-technical-documentation` skill,
     - `authoring-product-documentation` skill,
     - `authoring-readme-file` skill.
   - If the change-set modifies `docs/product-documentation.md`, the agent MUST perform test-case impact analysis using `docs/test-case-authoring-specification.md`.
4. Change-Set Commit
   - The agent MUST commit each completed change-set as exactly one commit.
   - The agent MUST NOT mark the change-set as completed before this commit exists.
5. Change-Set Closure Report
   - The agent MUST provide a short closure report for each completed change-set that includes:
     - change-set identifier,
     - commit hash,
     - mandatory verification command results for that change-set,
     - documentation-skill decisions (`UPDATE_REQUIRED` or `NO_UPDATE_REQUIRED`) with touched files.

### Completion

Completion MUST use `update_plan`, update statuses after each workflow step in this phase, and keep exactly one step `in_progress` at a time.

1. Full-Plan Completion Verification
   - The agent MUST verify that all approved change-sets from the plan were implemented, or that any approved deviation is explicitly documented.
   - The agent MUST verify `one change-set = one commit` across the full plan and MUST explicitly list this mapping check result.
   - The agent MUST verify that measurable success criteria are satisfied for the full planned scope.
   - The agent MUST verify that required tests were added or updated according to Engineering Guardrails, or that an exception is explicitly documented.
   - The agent MUST verify that all mandatory verification commands were completed for the full planned scope, or that a limitation is explicitly documented.
   - The agent MUST verify that mandatory tests pass, or that a limitation is explicitly documented.
   - The agent MUST verify that naming and terminology remain consistent across the full planned scope.
   - If any required verification item fails or is missing, the agent MUST stop completion and finish missing workflow steps first.
2. Final Completion Report
   - After completing the full planned scope, the agent MUST provide one final completion report.
   - The report MUST include `CHANGES MADE` as a file-level summary of what changed and why.
   - The report MUST include `RISKS / VERIFY` as potential regressions and additional checks to run, including manual tests.
   - The report MUST include `CHANGE-SET COMMITS` with each completed change-set mapped to exactly one commit hash.
   - The report MUST include all accepted local exceptions (for example linter or security suppressions) with concrete rationale.
   - The report SHOULD stay short and concrete so a junior engineer can quickly review and validate the result.
   - The agent MUST NOT publish a final completion response until all mandatory sections in this phase are present.

## 4. Engineering Guardrails

### 4.1 Dependencies and Toolchain

- Dependency/toolchain baseline MUST be taken from:
  - `docs/technical-documentation.md#9-technology-stack-and-versions`
  - `go.mod`
- Adding third-party dependencies MUST have explicit approval.

### 4.2 Architecture

The agent MUST use `docs/technical-documentation.md#3-architecture-and-boundaries` as the primary architecture guide.
For non-trivial architecture work, the agent SHOULD consult `docs/clean-architecture-ddd.md`, especially for boundary changes, dependency-direction decisions, and new ports/adapters.

Non-negotiable summary:

- Dependencies MUST point inward.
- Domain MUST stay isolated from outer layers.
- TUI MUST remain an adapter (no direct database/business rule implementation).
- Infrastructure MUST implement ports and MUST NOT drive use case logic.
- The implementation MUST respect architecture boundaries and dependency direction.
- The implementation SHOULD prefer interface-driven changes through application ports.
- Interface adapters MUST NOT bypass use cases.

### 4.2.1 Architecture Rules for New Features

When adding functionality that changes behavior, the agent MUST follow this order:

1. Start from domain model/service changes if behavior changes domain rules.
2. Add/update use case orchestration.
3. Extend port interfaces only when a new boundary is required.
4. Implement infrastructure adapters for new port behavior.
5. Connect UI adapter to use case, not to infrastructure.

For adapter-only or infrastructure-only changes that do not change domain behavior, steps `1` and `2` MAY be no-op, but dependency direction and architecture boundaries MUST still be preserved.

### 4.3 Development Standards

#### 4.3.1 General Development Rules

- The agent MUST use English for identifiers and internal technical documentation.
- The implementation SHOULD prefer the simplest solution that satisfies requirements.
- The agent MUST NOT add speculative abstractions, configurability, or extensibility that were not requested.
- Changes MUST stay minimal and scoped to task intent.
- Changes MUST stay surgical; every changed line MUST map directly to task intent.
- The agent MUST NOT refactor adjacent or orthogonal code unless explicitly requested.
- If unrelated issues are discovered, the agent MUST report them separately instead of changing them.

Quick examples:

- Good: edit one use case and its tests for one behavior change.
- Bad: adding new generic helper layers "for future reuse" when only one call site exists.

#### 4.3.2 TDD Rules

- Before starting unit-test work, the agent MUST invoke skill `create-unit-tests`.
- This prerequisite MUST apply to adding, editing, fixing, refactoring, reviewing, and designing unit tests.
- The agent MUST apply this prerequisite independent of the chosen language, test framework, or test workflow (`TDD`, `BDD`, or test-after).
- TDD MUST be applied for every feature change, bug fix, and behavior-impacting refactor.
- The agent MUST treat `docs/test-driven-development.md` as the normative TDD reference and SHOULD consult it for behavior-impacting implementation, test strategy updates, and non-trivial Red-Green-Refactor decisions.
- For bug fixes, the agent MUST add a regression unit test that reproduces the bug before applying the fix.
- The agent MUST NOT weaken assertions only to make failing behavior pass.
- The agent MUST NOT skip the `Red` step unless technically impossible; if impossible, the agent MUST document the reason and treat test-after as an explicit exception.

#### 4.3.3 Go-Specific Rules

- The agent MUST write idiomatic Go.
- The agent SHOULD keep functions focused and explicit in error handling.
- To format changed Go files, the agent MUST run `gofmt -w <changed-go-files>`.
- To lint full repository scope, the agent MUST run `golangci-lint run ./...`.
- To test full repository scope, the agent MUST run `go test ./...`.
- The agent MUST NOT introduce new lint violations.
- The agent MUST NOT use unchecked `defer x.Close()` in production code.
- The agent MUST handle `Close()` errors explicitly or justify a deliberate ignore.
- The agent MUST NOT build runtime SQL using unvalidated string interpolation.
- The agent MUST use placeholders for values.
- For dynamic identifiers (table/column names), the agent MUST use strict allowlist and/or safe identifier quoting.
- The agent MUST NOT disable security linters globally to silence findings.
- If an exception is required, the agent MUST apply it locally (`#nosec` / `nolint`) with a concrete inline justification.
- Every local linter exception MUST have explicit user approval each time (no blanket pre-approval).

## 5. Documentation Policy

For tasks that directly create or modify documentation files, the agent MUST invoke the matching skill:

- For `docs/product-documentation.md`, the agent MUST invoke `authoring-product-documentation`.
- For `docs/technical-documentation.md`, the agent MUST invoke `authoring-technical-documentation`.
- For `README.md`, the agent MUST invoke `authoring-readme-file`.

If multiple documentation perspectives are affected, the agent MUST invoke all applicable skills independently and apply each skill decision.

### 5.1 Product Documentation Policy

- Product documentation policy is governed exclusively by skill `authoring-product-documentation`; `AGENTS.md` MUST NOT define additional or duplicate product-documentation authoring/decision rules.
- The agent MUST accept the invoked skill decision (`UPDATE_REQUIRED` or `NO_UPDATE_REQUIRED`) and proceed accordingly.

### 5.2 Technical Documentation Policy

- Technical documentation policy is governed exclusively by skill `authoring-technical-documentation`; `AGENTS.md` MUST NOT define additional or duplicate technical-documentation authoring/decision rules.
- The agent MUST accept the invoked skill decision (`UPDATE_REQUIRED` or `NO_UPDATE_REQUIRED`) and proceed accordingly.

### 5.3 README Documentation Policy

- README policy is governed exclusively by skill `authoring-readme-file`; `AGENTS.md` MUST NOT define additional or duplicate README authoring/decision rules.
- The agent MUST accept the invoked skill decision (`UPDATE_REQUIRED` or `NO_UPDATE_REQUIRED`) and proceed accordingly.

## 6. Cross-Cutting Operational Rules

- For commit-message creation, validation, classification, or commit requests without an explicit message, the agent MUST invoke skill `write-commit-messages`.
- For manual `TC-*` execution and reporting (`single test case` and `full test case suite`), the agent MUST use `docs/test-case-execution-reporting-specification.md`.
- Whenever the agent asks the user a question, it MUST present exactly four numbered response options:
  - Options `1`, `2`, and `3` MUST be predefined choices.
  - Option `4` MUST allow the user to provide a custom response.
- Whenever any file is renamed or moved, the agent MUST update inbound references to that file across the repository in the same change-set; exclude completed PRD and TASK artifacts.
- Whenever Markdown headings are changed (title or numeric prefix), the agent MUST update inbound heading references across the repository in the same change-set.

## 7. Quick Reference

- Product source of truth: `docs/product-documentation.md`
- Technical source of truth: `docs/technical-documentation.md`
- Architecture deep dive: `docs/clean-architecture-ddd.md`
- TDD deep dive: `docs/test-driven-development.md`
- Run/setup basics: `README.md`

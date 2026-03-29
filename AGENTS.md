# AGENTS

## Global Normative and Language Rules

- The keywords `MUST`, `MUST NOT`, `SHOULD`, `SHOULD NOT`, and `MAY` in this file MUST be interpreted as described in RFC 2119.
- The agent MUST communicate with the user in Polish by default.
- The agent MUST use English for identifiers, code, plans, internal technical documentation, and operational artifacts.
- Exception: The agent MAY use another language for direct user communication when the user explicitly requests it.
- If documentation that describes current state conflicts with current code behavior, the agent MUST treat current code as the factual current state.
- If current code conflicts with canonical rules in `AGENTS.md` or `docs/clean-architecture-ddd.md`, the agent MUST treat the code as factual current state but MUST NOT use that conflict to justify future changes.
- When the task touches a material conflict between current code and canonical rules, the agent MUST report that drift explicitly.

## Source Documents

- `docs/technical-documentation.md` MUST be used as the primary source for current implementation state, current code structure, current technical contracts and mechanisms, current technical constraints, and known drift.
- `docs/product-documentation.md` MUST be used as the primary source for current user-visible product state, current workflows, current interaction rules, current non-goals, and user-visible constraints.
- `README.md` SHOULD be used for repository entry context, including product summary, supported environments, installation prerequisites, and basic run commands.
- `docs/clean-architecture-ddd.md` MUST be used as the canonical architecture source for boundary placement, dependency direction, logic placement, ports/adapters decisions, and application vs adapter responsibilities.
- `docs/test-driven-development.md` MUST be used as the normative TDD reference for behavior-impacting changes, test strategy updates, and Red-Green-Refactor decisions.

## Engineering Guardrails

### Dependencies and Toolchain

- The agent MUST take the dependency and toolchain baseline from `go.mod` and `docs/technical-documentation.md`.
- Adding third-party dependencies MUST have explicit approval.

### Architecture

#### Architecture Authority and Interpretation

- The agent MUST use `docs/clean-architecture-ddd.md` as the canonical architecture guide for all future code changes.
- The agent MUST treat `docs/technical-documentation.md` as a description of current implementation state, not as the canonical source of future architecture rules.
- The agent MUST treat `docs/product-documentation.md` as a description of current product behavior, not as the canonical source of internal implementation rules.
- If current code conflicts with `docs/clean-architecture-ddd.md`, the agent MUST treat the code as factual current state but MUST NOT extend, normalize, or justify the conflict for future changes.

#### Dependency Boundaries

- Allowed dependency direction MUST remain `interfaces -> application -> domain` and `infrastructure -> application/domain`.
- `internal/domain/**` MUST NOT import `internal/application/**`, `internal/interfaces/**`, or `internal/infrastructure/**`.
- `internal/application/**` MUST NOT import `internal/interfaces/**` or `internal/infrastructure/**`.
- Repository interfaces MUST belong to the inner layers, and repository implementations MUST live in infrastructure adapters.
- Interface adapters MUST NOT import infrastructure adapters directly.

#### Logic Placement

- `internal/interfaces/**` MUST be limited to input handling, presentation, interaction-local state, DTO mapping, and use-case invocation.
- `internal/interfaces/**` MUST NOT own business rules, decision policies, workflow orchestration, identity derivation, persistence semantics, or state-transition policy.
- `internal/interfaces/tui/**` MUST remain an interface adapter and MUST NOT absorb business behavior that belongs to the domain or use cases.
- Use cases MUST own application workflow orchestration and cross-component decision flow, but they MUST NOT absorb domain invariants or domain rules that belong in domain models or domain services.

#### Change Placement Rules

- When a new behavior is added, the agent MUST classify it as domain, application, interface adapter, or infrastructure before implementation.
- Minimal or surgical change scope MUST NOT be used as justification for placing logic in the wrong architectural layer.
- When adding functionality that changes behavior, the agent MUST prefer implementation flow from inner layers outward: domain, use case, port, infrastructure adapter, then UI adapter.
- For adapter-only or infrastructure-only changes, the inner-layer steps MAY be no-op only when the change does not introduce application logic, business rules, or workflow decisions; architecture boundaries and dependency direction MUST still be preserved.

#### Architecture Maintainability Preferences

- The agent MUST treat human and AI discoverability as first-class quality concerns.
- The agent SHOULD prefer structures where the likely change location is predictable from naming, boundaries, and module ownership.
- The agent SHOULD prefer interface-driven changes through application ports instead of adapter-to-adapter coupling.
- The agent SHOULD prefer finer-grained files, packages, and modules when the current code mixes separable architectural responsibilities or crosses stable change seams.
- The agent MUST NOT split files, packages, or modules only to reduce size, only to reduce token usage, or only to satisfy a generic granularity preference.
- The agent MAY treat lower token consumption, smaller review surface, and easier navigation as secondary benefits when a decomposition is already justified by cohesion, boundary clarity, testability, or reduced change blast radius.
- When proposing or applying decomposition, the agent MUST be able to name the architectural seam or responsibility split that justifies it.

### Development Standards

#### General Development Rules

- The agent MUST use English for identifiers and internal technical documentation.
- The implementation SHOULD prefer the simplest solution that satisfies requirements.
- For feature changes, bug fixes, and behavior-impacting refactors, the agent MUST follow `TDD Rules`.
- Before finalizing a non-documentation code change, the agent MUST run the applicable formatter, tests, and linter for the affected stack.
- The agent MUST NOT add speculative abstractions, configurability, or extensibility that were not requested.
- Changes MUST stay minimal and scoped to task intent.
- Changes MUST stay surgical; every changed line MUST map directly to task intent.
- The agent MUST NOT refactor adjacent or orthogonal code unless explicitly requested.
- The agent SHOULD prefer decomposition or simplification that removes mixed responsibilities, duplicated orchestration, unstable change coupling, unnecessary complexity, unnecessary nesting, redundancy, or over-abstraction, but it MUST keep cohesive workflows together when splitting would add indirection without architectural gain.
- The agent MUST NOT simplify in a way that merges distinct concerns, weakens separation of responsibilities, or makes debugging harder.
- If unrelated issues are discovered, the agent MUST report them separately instead of changing them.

Quick examples:

- Good: edit one use case and its tests for one behavior change.
- Bad: adding new generic helper layers "for future reuse" when only one call site exists.

#### TDD Rules

- This section MUST be applied whenever `General Development Rules` routes work to `TDD Rules`.
- Before starting unit-test work for such changes, the agent MUST invoke skill `create-unit-tests`.
- The `create-unit-tests` prerequisite MUST apply to adding, editing, fixing, refactoring, reviewing, and designing unit tests, independent of the chosen language, test framework, or test workflow (`TDD`, `BDD`, or test-after).
- The agent MUST treat `docs/test-driven-development.md` as the normative TDD reference and SHOULD consult it for behavior-impacting implementation, test strategy updates, and non-trivial Red-Green-Refactor decisions.
- For bug fixes, the agent MUST add a regression unit test that reproduces the bug before applying the fix.
- The agent MUST NOT weaken assertions only to make failing behavior pass.
- The agent MUST NOT skip the `Red` step unless technically impossible; if impossible, the agent MUST document the reason and treat test-after as an explicit exception.

#### Go-Specific Rules

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

## Cross-Cutting Operational Rules

### General Rules

- For commit-message creation, validation, classification, or commit requests without an explicit message, the agent MUST invoke skill `write-commit-messages`.
- For manual `TC-*` execution and reporting (`single test case` and `full test case suite`), the agent MUST use `docs/test-case-execution-reporting-specification.md`.

### Plan Mode Rules

These rules apply only in Plan mode.
- For coding-related plans, the plan MUST explicitly account for the applicable rules defined in `Development Standards`. If the plan requires a justified deviation (for example for a large refactoring), that deviation MUST be explicitly labeled with its scope and rationale.
- For non-trivial coding plans, the agent MUST inspect the current implementation at the most likely change seams before finalizing the plan.
- The agent MUST distinguish explicitly between confirmed facts, assumptions, and open questions.
- If a decision-critical open question remains unresolved, the agent MUST NOT present the plan as execution-ready and MUST call out the gap explicitly.
- When planning a change to existing behavior or functionality, the agent MUST keep repository planning artifacts focused on the intended resulting state and MUST NOT leave information about the previous state unless the user explicitly requests that historical context.
- When persisting a generated plan to the repository, the agent MUST save that plan under `.plans/` with a short descriptive kebab-case filename that communicates the task intent.
- When the user asks to save a generated plan after a `<proposed_plan>` was already presented, the agent MUST treat that as a persistence request rather than a request to regenerate, restyle, summarize, expand, or otherwise rewrite the plan content. The persisted file MUST be textually identical to the most recently presented `<proposed_plan>` content unless the user explicitly requests edits before saving.
- A generated plan MUST include references to the documentation files that SHOULD be consulted to enrich task context for that specific task, with a short note describing why each file is relevant.

### Repository Consistency Rules

- Whenever any file is renamed or moved, the agent MUST update inbound references to that file across the repository in the same change-set; exclude completed PRD and TASK artifacts.
- Whenever Markdown headings are changed (title or numeric prefix), the agent MUST update inbound heading references across the repository in the same change-set.

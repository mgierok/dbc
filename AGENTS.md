# AGENTS

## Global Normative and Language Rules

- The keywords `MUST`, `MUST NOT`, `SHOULD`, `SHOULD NOT`, and `MAY` in this file MUST be interpreted as described in RFC 2119.
- The agent MUST use English regardless of the language used in user instructions.
- If any documentation conflicts with current code behavior, the agent MUST treat current code as factual state.

## Source Documents

- `docs/technical-documentation.md` MUST be used as the primary source for architecture boundaries, dependency direction, component responsibilities, technical contracts, supported stack versions, and operational constraints.
- `docs/product-documentation.md` MUST be used as the primary source for user-visible behavior, workflows, interaction rules, product constraints, and current non-goals.
- `README.md` SHOULD be used for repository entry context, including product summary, supported environments, installation prerequisites, and basic run commands.
- `docs/clean-architecture-ddd.md` SHOULD be used for non-trivial architecture work, especially boundary changes, dependency-direction decisions, and new ports/adapters.
- `docs/test-driven-development.md` MUST be used as the normative TDD reference for behavior-impacting changes, test strategy updates, and Red-Green-Refactor decisions.

## Engineering Guardrails

### Dependencies and Toolchain

- The agent MUST take the dependency and toolchain baseline from `go.mod` and `docs/technical-documentation.md`.
- Adding third-party dependencies MUST have explicit approval.

### Architecture

The agent MUST use `docs/technical-documentation.md#architecture-and-boundaries` as the primary architecture guide.
For non-trivial architecture work, the agent SHOULD consult `docs/clean-architecture-ddd.md`, especially for boundary changes, dependency-direction decisions, and new ports/adapters.
- The agent MUST preserve the architecture boundaries and dependency direction defined in `docs/technical-documentation.md`.
- The agent MUST treat human and AI discoverability as first-class quality concerns.
- The agent SHOULD prefer structures where the likely change location is predictable from naming, boundaries, and module ownership.
- The agent SHOULD prefer interface-driven changes through application ports instead of adapter-to-adapter coupling.
- The agent SHOULD prefer finer-grained files, packages, and modules when the current code mixes separable architectural responsibilities or crosses stable change seams.
- The agent MUST NOT split files, packages, or modules only to reduce size, only to reduce token usage, or only to satisfy a generic granularity preference.
- The agent MAY treat lower token consumption, smaller review surface, and easier navigation as secondary benefits when a decomposition is already justified by cohesion, boundary clarity, testability, or reduced change blast radius.
- When proposing or applying decomposition, the agent MUST be able to name the architectural seam or responsibility split that justifies it.

#### Architecture Rules for New Features

- When adding functionality that changes behavior, the agent MUST prefer implementation flow from inner layers outward: domain, use case, port, infrastructure adapter, then UI adapter.
- For adapter-only or infrastructure-only changes that do not change domain behavior, the inner-layer steps MAY be no-op, but architecture boundaries and dependency direction MUST still be preserved.

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

- For commit-message creation, validation, classification, or commit requests without an explicit message, the agent MUST invoke skill `write-commit-messages`.
- For manual `TC-*` execution and reporting (`single test case` and `full test case suite`), the agent MUST use `docs/test-case-execution-reporting-specification.md`.
- Whenever the agent asks the user a question, it MUST present exactly four numbered response options:
  - Options `1`, `2`, and `3` MUST be predefined choices.
  - Option `4` MUST allow the user to provide a custom response.
- Whenever any file is renamed or moved, the agent MUST update inbound references to that file across the repository in the same change-set; exclude completed PRD and TASK artifacts.
- Whenever Markdown headings are changed (title or numeric prefix), the agent MUST update inbound heading references across the repository in the same change-set.

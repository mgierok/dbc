# AGENTS

## Global Normative and Language Rules

- The keywords `MUST`, `MUST NOT`, `SHOULD`, `SHOULD NOT`, and `MAY` in this file MUST be interpreted as described in RFC 2119.
- The agent MUST use English regardless of the language used in user instructions.

## Engineering Guardrails

### Dependencies and Toolchain

- Dependency/toolchain baseline MUST be taken from:
  - `docs/technical-documentation.md#technology-stack-and-versions`
  - `go.mod`
- Adding third-party dependencies MUST have explicit approval.

### Architecture

The agent MUST use `docs/technical-documentation.md#architecture-and-boundaries` as the primary architecture guide.
For non-trivial architecture work, the agent SHOULD consult `docs/clean-architecture-ddd.md`, especially for boundary changes, dependency-direction decisions, and new ports/adapters.

Non-negotiable summary:

- Dependencies MUST point inward.
- Domain MUST stay isolated from outer layers.
- TUI MUST remain an adapter (no direct database/business rule implementation).
- Infrastructure MUST implement ports and MUST NOT drive use case logic.
- The implementation MUST respect architecture boundaries and dependency direction.
- The implementation SHOULD prefer interface-driven changes through application ports.
- Interface adapters MUST NOT bypass use cases.

#### Architecture Rules for New Features

When adding functionality that changes behavior, the agent MUST follow this order:

1. Start from domain model/service changes if behavior changes domain rules.
2. Add/update use case orchestration.
3. Extend port interfaces only when a new boundary is required.
4. Implement infrastructure adapters for new port behavior.
5. Connect UI adapter to use case, not to infrastructure.

For adapter-only or infrastructure-only changes that do not change domain behavior, steps `1` and `2` MAY be no-op, but dependency direction and architecture boundaries MUST still be preserved.

### Development Standards

#### General Development Rules

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

#### TDD Rules

- Before starting unit-test work, the agent MUST invoke skill `create-unit-tests`.
- This prerequisite MUST apply to adding, editing, fixing, refactoring, reviewing, and designing unit tests.
- The agent MUST apply this prerequisite independent of the chosen language, test framework, or test workflow (`TDD`, `BDD`, or test-after).
- TDD MUST be applied for every feature change, bug fix, and behavior-impacting refactor.
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

## Documentation

- If any documentation conflicts with current code behavior, the agent MUST treat current code as factual state.
- For tasks that directly create or modify documentation files, the agent MUST invoke the matching skill:
  - For `docs/product-documentation.md`, the agent MUST invoke `authoring-product-documentation`.
  - For `docs/technical-documentation.md`, the agent MUST invoke `authoring-technical-documentation`.
  - For `README.md`, the agent MUST invoke `authoring-readme-file`.

### Product Documentation Policy

- Product documentation policy is governed exclusively by skill `authoring-product-documentation`; `AGENTS.md` MUST NOT define additional or duplicate product-documentation authoring/decision rules.
- The agent MUST accept the invoked skill decision (`UPDATE_REQUIRED` or `NO_UPDATE_REQUIRED`) and proceed accordingly.

### Technical Documentation Policy

- Technical documentation policy is governed exclusively by skill `authoring-technical-documentation`; `AGENTS.md` MUST NOT define additional or duplicate technical-documentation authoring/decision rules.
- The agent MUST accept the invoked skill decision (`UPDATE_REQUIRED` or `NO_UPDATE_REQUIRED`) and proceed accordingly.

### README Documentation Policy

- README policy is governed exclusively by skill `authoring-readme-file`; `AGENTS.md` MUST NOT define additional or duplicate README authoring/decision rules.
- The agent MUST accept the invoked skill decision (`UPDATE_REQUIRED` or `NO_UPDATE_REQUIRED`) and proceed accordingly.

## Cross-Cutting Operational Rules

- For commit-message creation, validation, classification, or commit requests without an explicit message, the agent MUST invoke skill `write-commit-messages`.
- For manual `TC-*` execution and reporting (`single test case` and `full test case suite`), the agent MUST use `docs/test-case-execution-reporting-specification.md`.
- For structured multi-phase delivery workflow execution, the agent MUST invoke skill `delivery-workflow` only when explicitly requested by the user.
- Whenever the agent asks the user a question, it MUST present exactly four numbered response options:
  - Options `1`, `2`, and `3` MUST be predefined choices.
  - Option `4` MUST allow the user to provide a custom response.
- Whenever any file is renamed or moved, the agent MUST update inbound references to that file across the repository in the same change-set; exclude completed PRD and TASK artifacts.
- Whenever Markdown headings are changed (title or numeric prefix), the agent MUST update inbound heading references across the repository in the same change-set.

## Quick Reference

- Product source of truth: `docs/product-documentation.md`
- Technical source of truth: `docs/technical-documentation.md`
- Architecture deep dive: `docs/clean-architecture-ddd.md`
- TDD deep dive: `docs/test-driven-development.md`
- Run/setup basics: `README.md`

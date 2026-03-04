---
name: architecture-audit
description: Audit architectural conformance for a full repository or a user-specified directory and produce actionable findings for architecture-safe optimization, quality improvement, and controlled rule-exception candidates. Use only when the user explicitly requests this skill by name; this skill MUST NOT auto-trigger.
---

# Architecture Audit

## Purpose

This skill performs an analysis-only, periodic architecture audit to:
- verify conformance with project-defined architectural rules across repository scope or selected module scope,
- identify confirmed and suspected architectural violations with concrete evidence,
- propose prioritized, implementable change recommendations that optimize code quality and maintainability while preserving the globally defined architecture,
- identify controlled rule-exception candidates and explain when violating a specific rule may provide net benefit,
- identify optimization opportunities through design patterns that may be missing due to iterative project growth.

The skill is technology agnostic and project agnostic. It MUST treat project-defined architecture rules and current code behavior as the source of truth.

## Scope

- In scope:
  - auditing architecture conformance in repository scope or a user-selected subdirectory,
  - identifying boundary, dependency-direction, and layering violations,
  - proposing remediation options with risk notes and expected impact,
  - proposing architecture-safe optimization and quality-improvement actions,
  - proposing design-pattern adoption opportunities with concrete fit rationale,
  - proposing controlled, explicit rule-exception candidates with trade-off analysis.
- Out of scope:
  - editing production code automatically,
  - executing architecture refactors automatically,
  - enforcing one universal architecture style across all projects.

## Explicit Invocation Policy

- This skill MUST run only when the user explicitly requests this skill by name.
- This skill MUST NOT auto-trigger.
- If the user request does not explicitly name this skill, the skill MUST ask the user whether they want to invoke `architecture-audit` by name.
- The skill MUST use `update_plan` from workflow start through completion, updating statuses after each workflow step and finishing with all steps marked as `completed`.

## Inputs

- `scope`:
  - optional path to analyze,
  - default is repository root when no path is provided.
- `project architecture rules`:
  - architecture sections from governance and technical docs,
  - architecture decision records, module conventions, and other local policies if present.

## Architecture Review Areas (Relevance-Gated)

- The skill MUST evaluate the following 15 areas when they are relevant to the analyzed project or selected scope.
- For each area, the skill MUST classify status as `APPLICABLE` or `NOT_APPLICABLE`.
- Every `NOT_APPLICABLE` classification MUST include a concrete one-line reason.
- Every finding and recommendation SHOULD reference one or more areas by name to keep traceability explicit.

1. Boundaries and module decomposition
- The audit MUST verify that layer/module/bounded-context boundaries are clear and consistently applied.
- The audit MUST verify dependency direction (for example domain not depending on infrastructure).
- The audit SHOULD flag god modules and overgrown shared utility modules.
- The audit SHOULD assess whether module public APIs are small and stable.

2. Dependencies and import rules
- The audit MUST inspect dependency graph risks (cycles, spaghetti imports, extreme fan-in/fan-out).
- The audit SHOULD verify whether architecture dependency rules exist (for example ArchUnit, Deptrac, ESLint boundaries) and whether CI enforces them.
- The audit SHOULD verify control of external dependencies (lockfiles, versions, licenses).

3. Data flow and contracts
- The audit MUST trace main data flows (DTO/Command/Query boundaries, mappings, validation points).
- The audit SHOULD verify that integration contracts are explicit (types, schemas, OpenAPI/AsyncAPI, protobuf, equivalent artifacts).
- The audit MUST flag leaky abstractions where persistence or transport internals leak across boundaries.

4. Domain model and business logic placement
- The audit MUST verify that business logic stays in domain/application core rather than being spread across controllers, ORMs, or infrastructure handlers.
- The audit SHOULD assess naming coherence, aggregate boundaries, invariants, and transaction semantics.
- The audit SHOULD verify that use cases remain testable and framework-independent.

5. Integrations and communication (sync/async)
- The audit SHOULD assess whether communication style choices (REST, gRPC, events) appear intentional and consistent with constraints.
- The audit SHOULD verify idempotency, retry policies, timeouts, circuit breaking, and backpressure where applicable.
- The audit SHOULD assess error semantics (codes, taxonomy, dead-letter handling when messaging exists).

6. Persistence and transactions
- The audit MUST assess transaction boundaries, data consistency risks, and race-condition exposure.
- The audit SHOULD flag N+1 and uncontrolled eager/lazy loading patterns when applicable.
- The audit SHOULD evaluate migration discipline, schema versioning, and backward compatibility expectations.
- The audit MUST flag repository/DAO leakage of database details into domain-facing contracts.

7. Configuration and environments
- The audit SHOULD assess 12-factor style configuration (config outside code, sensible defaults).
- The audit MUST flag obvious secret-management risks (secrets in repo, weak secret boundaries).
- The audit SHOULD assess feature-flag strategy, environment profiling, and dev/stage/prod parity.

8. Security in architecture
- The audit MUST assess authentication and authorization placement and consistency.
- The audit MUST assess trust-boundary handling and input validation at boundary edges.
- The audit SHOULD evaluate permission modeling and multi-tenant isolation when relevant.
- The audit SHOULD verify dependency security posture checks (SCA and update policy signals) when such workflows exist.

9. Observability and diagnostics
- The audit SHOULD assess structured logging, request correlation, and trace/span propagation where applicable.
- The audit SHOULD assess metrics and service-level indicators relevant to the architecture.
- The audit SHOULD assess distributed tracing and error-logging conventions when supported.
- The audit SHOULD verify whether a practical incident-debugging golden path exists.

10. Reliability and resilience
- The audit SHOULD verify timeout usage and safe retry boundaries.
- The audit SHOULD assess graceful shutdown and liveness/readiness patterns where runtime services exist.
- The audit SHOULD assess resource protection strategies (thread pools, DB pools, queues, caches).
- The audit SHOULD assess degradation strategies (fallbacks, cache-first, read-only mode) when architecture suggests such needs.

11. Performance and scalability
- The audit SHOULD assess hot paths, serialization overhead, allocation/GC pressure, or equivalent runtime costs where relevant.
- The audit SHOULD evaluate cache placement, invalidation, and stampede protection when caching exists.
- The audit SHOULD assess whether current architecture supports expected horizontal or vertical scaling model.

12. Testability and maintainability
- The audit SHOULD assess test pyramid balance (unit/integration/contract/e2e) for architecture-critical paths.
- The audit SHOULD verify contract-testing patterns for integrations when interfaces are owned or versioned.
- The audit SHOULD flag nondeterministic test risks (time/random/external state without control).
- The audit SHOULD assess modularity for testing (DI, ports/adapters, seam quality).

13. Conventions, standards, and in-code documentation
- The audit SHOULD assess ADR/module-README/diagram availability where architectural decisions are non-trivial.
- The audit SHOULD verify lint/format/review-checklist enforcement signals.
- The audit SHOULD assess consistency of error handling, validation, and transaction patterns.

14. CI/CD and architecture rule enforcement
- The audit SHOULD assess whether pipeline stages cover build, test, scans, and quality gates.
- The audit SHOULD verify whether architecture-regression detection exists for dependency rules.
- The audit SHOULD assess API versioning and compatibility discipline (for example SemVer and deprecation policy) when externally consumed contracts exist.

15. Complexity and simplification pressure
- The audit MUST identify accidental-complexity hotspots (for example over-abstraction, unnecessary indirection, dead layers, and speculative extension points).
- The audit SHOULD flag duplicated orchestration paths and overlapping module responsibilities.
- The audit MUST propose simplification actions that preserve architecture boundaries and dependency direction.
- The audit SHOULD include measurable simplification outcomes (for example fewer dependency edges, smaller change blast radius, and reduced cognitive load in core flows).

## Audit Workflow

1. Confirm scope and inventory architecture units
- Resolve audit scope (`repo root` or provided directory).
- Inventory relevant units (packages, modules, layers, adapters, services, boundaries).

2. Load architecture expectations
- Discover and load project-specific architecture rules before applying generic heuristics.
- If architecture rules are missing or ambiguous, explicitly state that and continue with heuristic fallback.

3. Classify review-area applicability
- Evaluate each architecture review area defined in `Architecture Review Areas (Relevance-Gated)` and set it to `APPLICABLE` or `NOT_APPLICABLE`.
- If an area is `NOT_APPLICABLE`, record one concrete reason tied to project/scope characteristics.
- Use the classified applicability set to control audit depth for downstream steps.

4. Build dependency and boundary map
- Map inbound and outbound dependencies between units.
- Map boundary crossings and integration points.
- Identify allowed and forbidden dependency directions for the selected scope.

5. Identify conformance findings
- Use two finding classes:
  - `ARCH_VIOLATION_CONFIRMED`: evidence clearly shows a rule violation.
  - `REVIEW_REQUIRED`: potential violation that needs manual validation.
- Every finding MUST include:
  - architecture review area name(s),
  - file path and unit identifier,
  - violated rule or expected boundary,
  - concrete evidence,
  - risk and impact note,
  - recommended remediation option.

6. Identify architecture-safe optimization and quality opportunities
- Capture opportunities that improve readability, maintainability, cohesion, coupling, and runtime behavior while preserving architecture rules.
- Every opportunity MUST include:
  - architecture review area name(s),
  - target files or units,
  - expected quality or performance gain,
  - why the recommendation remains architecture-compliant,
  - implementation effort estimate.

7. Identify pattern-based optimization opportunities
- Evaluate whether missing design patterns are causing duplicated logic, rigid coupling, unstable change surfaces, or orchestration leakage.
- Pattern recommendations MUST be problem-driven and MUST NOT introduce speculative abstractions.
- Every pattern recommendation MUST include:
  - architecture review area name(s),
  - current problem signal,
  - proposed pattern,
  - expected gain,
  - over-engineering risk note.

8. Identify controlled rule-exception candidates
- The skill MAY propose a `RULE_EXCEPTION_CANDIDATE` only when the expected benefit is concrete and measurable.
- Every candidate MUST include:
  - architecture review area name(s),
  - exact rule that would be bent or broken,
  - explicit reason why strict compliance is costly or harmful in this context,
  - expected benefit of the exception,
  - risk, blast radius, and reversibility note,
  - required safeguards and verification steps.

9. Prioritize recommendations
- Group recommendations into:
  - `High impact`,
  - `Medium impact`,
  - `Low impact`.
- Prioritization SHOULD consider correctness risk, maintainability gain, migration effort, and blast radius.

10. Produce final report
- Generate a Markdown report using `references/report-template.md` as a baseline structure.
- The report MUST be extended with any required sections from `Output Requirements` that are missing in the template.
- Follow validation points from `references/analysis-checklist.md`.

## Decision Rules

- The skill MUST prefer project-native architecture rules over generic assumptions.
- The skill SHOULD avoid classifying findings as confirmed violations when evidence is weak.
- The skill MUST separate evidence from assumptions.
- The skill MUST flag uncertainty and request manual confirmation where needed.
- The skill MUST keep recommendations actionable, scoped, and file-specific.
- The skill MUST prioritize architecture-compliant recommendations before exception candidates.
- The skill MUST explain why a rule-exception candidate could be beneficial and what conditions make it acceptable.
- The skill MUST mark exception candidates as `OPTIONAL` and `REQUIRES_HUMAN_APPROVAL`.

## Output Requirements

The final output MUST be a Markdown report with these sections:
- `Scope and Inputs`
- `Architecture Rules Baseline`
- `Architecture Review Area Coverage`
- `Confirmed Violations`
- `Review Required Findings`
- `Compliant Optimization Opportunities`
- `Design Pattern Opportunities`
- `Rule-Exception Candidates`
- `Recommended Change Plan`
- `Confidence and Risk`
- `Evidence and Limitations`

The report MUST include concrete file references and architecture unit identifiers wherever possible.
`Architecture Review Area Coverage` MUST reflect the applicability classification produced in `Audit Workflow` step `3`, including one-line rationale for each `NOT_APPLICABLE` area.
The skill MUST write the report file to `<project-root>/architecture-audit.md`.

## Safety and Boundaries

- This skill MUST remain analysis-only for source code and configuration.
- This skill MUST NOT edit repository files except writing the required report artifact defined in `Output Requirements`.
- This skill MUST NOT infer authoritative architecture policy from code patterns alone when explicit project rules exist.
- This skill MAY provide optional follow-up commands for validation, but it MUST label them as optional.
- This skill MUST NOT treat rule-exception candidates as default recommendations.

## References

- `references/analysis-checklist.md`
- `references/report-template.md`

---
name: engineering-quality-audit
description: Audit engineering quality for a full repository or a user-selected directory by combining architecture conformance, unit-test quality, stale-test analysis, and coverage risk into one evidence-based report. Use only when the user explicitly requests this skill by name; this skill MUST NOT auto-trigger.
---

# Engineering Quality Audit

## Purpose

This skill performs an analysis-only engineering quality audit to:
- verify conformance with project-defined architecture and testing rules across repository scope or selected module scope,
- identify confirmed and suspected structural and unit-test quality problems with concrete evidence,
- identify stale or redundant unit tests and missing high-value coverage,
- propose prioritized, implementable changes that improve maintainability, test value, system safety, and human or AI operability without weakening architecture boundaries.

The skill is technology agnostic and project agnostic. It MUST treat project-defined architecture rules, project-defined testing rules, and current code behavior as the source of truth.

## Scope

- In scope:
  - auditing architecture conformance in repository scope or a user-selected subdirectory,
  - auditing navigability, change locality, and bounded file responsibility for both humans and AI coding agents,
  - auditing unit-test quality and testability in the same scope,
  - auditing test topology and whether tests are easy to locate from the affected production code,
  - identifying stale or redundant unit tests with evidence and risk notes,
  - identifying coverage gaps for critical behavior and boundary contracts,
  - proposing architecture-safe optimization opportunities and controlled rule-exception candidates.
- Out of scope:
  - editing production code or tests automatically,
  - deleting tests automatically,
  - enforcing one universal architecture style or testing style across all projects.

## Explicit Invocation Policy

- This skill MUST run only when the user explicitly requests this skill by name.
- This skill MUST NOT auto-trigger.
- If the user request does not explicitly name this skill, the skill MUST ask the user whether they want to invoke `engineering-quality-audit` by name.
- The skill MUST use `update_plan` from workflow start through completion, updating statuses after each workflow step and finishing with all steps marked as `completed`.

## Inputs

- `scope`:
  - optional path to analyze,
  - default is repository root when no path is provided.
- `project architecture rules`:
  - architecture sections from governance and technical docs,
  - architecture decision records, module conventions, and other local policies if present.
- `project testing rules`:
  - testing sections from governance and technical docs,
  - local testing conventions, test-layout policies, and other local policies if present.
- `quality-signal sources`:
  - lint, test, and coverage settings,
  - flaky-test signals and equivalent local evidence if present.

## Engineering Quality Review Areas (Relevance-Gated)

- The skill MUST evaluate the following 12 areas when they are relevant to the analyzed project or selected scope.
- For each area, the skill MUST classify status as `APPLICABLE` or `NOT_APPLICABLE`.
- Every `NOT_APPLICABLE` classification MUST include a concrete one-line reason.
- Every finding and recommendation SHOULD reference one or more areas by name to keep traceability explicit.

1. Navigability, change locality, and file responsibility
- The audit MUST assess whether the correct file, package, or module for a likely change can be predicted from naming, directory structure, and boundary placement.
- The audit MUST flag files, packages, or modules that mix multiple workflows or responsibilities and therefore require broad surrounding context to edit safely.
- The audit SHOULD identify high-search-cost and high-context-cost hotspots where a small change requires reading too many unrelated files or too much unrelated code.
- The audit SHOULD treat reduced token consumption, smaller review surface, and easier navigation as valid supporting outcomes when they result from better cohesion and clearer seams.

2. Boundaries, module decomposition, and simplification pressure
- The audit MUST verify that layer, module, or bounded-context boundaries are clear and consistently applied.
- The audit MUST verify dependency direction.
- The audit SHOULD flag god modules, dead layers, duplicated orchestration paths, and overlapping responsibilities.
- The audit SHOULD flag unnecessary complexity, deep nesting, redundancy, and over-abstraction when they increase cognitive load without protecting a valid architectural seam.
- The audit SHOULD assess whether decomposition opportunities follow stable seams rather than size alone.

3. Dependencies, data flow, and contracts
- The audit MUST inspect dependency graph risks such as cycles, spaghetti imports, and extreme fan-in or fan-out.
- The audit MUST trace main data flows, validation points, and mapping boundaries.
- The audit MUST flag leaky abstractions where persistence or transport details cross boundaries.
- The audit SHOULD verify that owned integration contracts are explicit when such contracts exist.

4. Business logic placement and persistence boundaries
- The audit MUST verify that business logic stays in the domain or application core rather than leaking into controllers, ORMs, or infrastructure handlers.
- The audit MUST assess transaction boundaries, data consistency risks, and repository or DAO leakage into domain-facing contracts.
- The audit SHOULD assess naming coherence, invariant placement, and aggregate or workflow ownership.

5. Integrations, configuration, and trust boundaries
- The audit SHOULD assess whether communication style choices appear intentional and consistent with constraints.
- The audit SHOULD verify idempotency, retry policies, timeouts, and backpressure where applicable.
- The audit MUST flag obvious secret-management risks and weak trust-boundary handling.
- The audit SHOULD assess feature-flag and environment-configuration discipline when relevant.

6. Reliability, observability, and performance signals
- The audit SHOULD verify timeout usage, graceful shutdown signals, and resource-protection patterns where runtime services exist.
- The audit SHOULD assess logging, metrics, tracing, and practical incident-debugging paths where such mechanisms are supported.
- The audit SHOULD assess hot paths, cache placement, and scalability signals where relevant.

7. Testability and seam quality
- The audit SHOULD assess whether use cases and core workflows remain testable and framework-independent.
- The audit SHOULD assess dependency injection quality, ports and adapters, and other seam-enabling patterns.
- The audit SHOULD assess test-pyramid balance and contract-testing patterns for architecture-critical boundaries.

8. Unit-test boundaries and isolation
- The audit MUST verify whether unit tests target one unit in isolation.
- The audit MUST verify that unit tests do not touch DB, file system, network, or real clock without stubs or test doubles.
- The audit MUST verify that boundaries between `unit`, `integration`, `contract`, and `e2e` tests are explicit.

9. Unit-test value and assertion quality
- The audit MUST verify that assertions are specific and behavior-focused.
- The audit SHOULD verify behavior-level checks instead of implementation-detail checks where feasible.
- The audit SHOULD flag mock overuse, weak interaction-only tests, and mocking of the subject under test.

10. Determinism and stability
- The audit MUST evaluate resilience to time dependencies, randomness, execution order, parallel execution, and global or singleton state.
- The audit MUST flag `sleep()` used as synchronization when stronger synchronization is feasible.
- The audit SHOULD assess flaky-test reporting or rerun mechanisms when available.
- The audit MUST flag cases where flakes appear masked instead of removed.

11. Coverage, regression protection, and false-green risk
- The audit MUST verify coverage of critical happy-path and unhappy-path business rules.
- The audit MUST verify coverage of important edge, error, state-transition, and boundary-contract scenarios when they are relevant.
- The audit SHOULD verify the presence of regression tests for known bugs.
- The audit SHOULD assess mutation-testing signals when available.
- The audit MUST NOT treat coverage percentage alone as evidence of sufficient protection.

12. Test topology, maintainability, and stale-test pressure
- The audit SHOULD verify that test data is readable, minimal, and local where practical.
- The audit SHOULD assess fixture, factory, or builder usage for hidden defaults that obscure test intent.
- The audit SHOULD verify that test names and organization make tests easy to find and understand.
- The audit SHOULD verify that test organization mirrors production structure closely enough that the affected tests are easy to locate from the changed production code.
- The audit SHOULD flag unusually high test concentration around a narrow production surface as a signal of mixed responsibilities, unstable seams, or high coordination cost.
- The audit SHOULD flag test helpers or fixture layers that force broad context loading to understand a localized change.
- The audit MUST identify tests that appear redundant, obsolete, or misaligned with the current contract, but MUST attach confidence and risk notes before recommending removal.

## Audit Workflow

1. Confirm scope and inventory quality units
- Resolve audit scope (`repo root` or provided directory).
- Inventory relevant code units, test files, configuration files, boundaries, and integration points.

2. Load project expectations
- Discover and load project-specific architecture rules and testing rules before applying generic heuristics.
- If either rule set is missing or ambiguous, explicitly state that and continue with heuristic fallback.

3. Classify review-area applicability
- Evaluate each engineering quality review area defined in `Engineering Quality Review Areas (Relevance-Gated)` and set it to `APPLICABLE` or `NOT_APPLICABLE`.
- If an area is `NOT_APPLICABLE`, record one concrete reason tied to project or scope characteristics.
- Use the classified applicability set to control audit depth for downstream steps.

4. Build structure, behavior, and change-locality maps
- Map inbound and outbound dependencies between relevant units.
- Map boundary crossings, integration points, and data or contract transitions.
- Map key behavior contracts to available tests and test types.
- Map likely change entry points for important workflows and assess whether naming and placement make them discoverable.
- Identify duplicated behavior checks, missing checks, risky blind spots, and high-search-cost or high-context-cost hotspots.

5. Identify confirmed and suspected quality findings
- Use two primary finding classes:
  - `QUALITY_VIOLATION_CONFIRMED`: evidence clearly shows a violated rule or unhealthy engineering-quality property.
  - `REVIEW_REQUIRED`: potential quality issue that needs manual validation.
- Every finding MUST include:
  - review area name(s),
  - file path and unit identifier or test identifier,
  - violated rule or expected property,
  - concrete evidence,
  - risk and impact note,
  - recommended remediation option.

6. Identify quality-preserving optimization opportunities
- Capture opportunities that improve readability, maintainability, cohesion, coupling, runtime behavior, or test value while preserving architecture rules.
- The skill MUST classify each such opportunity as `OPTIMIZATION_OPPORTUNITY`.
- Every opportunity MUST include:
  - review area name(s),
  - target files, units, or tests,
  - expected quality or performance gain,
  - why the recommendation remains compliant with project rules,
  - implementation effort estimate.
- If the opportunity recommends finer-grained decomposition, it MUST also include:
  - the responsibility split being proposed,
  - the architectural seam that makes the split valid,
  - why the split is better than keeping the current unit intact,
  - why the split is not merely a size-reduction exercise.
- If the opportunity recommends consolidating duplicated behavior checks or duplicated same-layer implementations, it MUST also include:
  - the shared behavior contract being duplicated,
  - the current drift risk or inconsistency signal,
  - why consolidation is justified without introducing speculative abstraction.

7. Identify coverage gaps and test removal candidates
- The skill MUST use two additional recommendation kinds:
  - `COVERAGE_GAP`
  - `TEST_REMOVAL_CANDIDATE`
- Every `COVERAGE_GAP` MUST include:
  - review area name(s),
  - affected module or contract,
  - missing scenario type,
  - why the gap matters,
  - recommended test addition,
  - priority.
- Every `TEST_REMOVAL_CANDIDATE` MUST include:
  - review area name(s),
  - file path and test identifier,
  - reason (`redundant`, `obsolete`, `misaligned with current contract`, or similar),
  - concrete evidence,
  - risk note,
  - validation step before deletion,
  - confidence level (`High` or `Medium`).

8. Identify controlled rule-exception candidates
- The skill MAY propose a `RULE_EXCEPTION_CANDIDATE` only when the expected benefit is concrete and measurable.
- Every candidate MUST include:
  - review area name(s),
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
- Prioritization SHOULD consider correctness risk, maintainability gain, test-value gain, migration effort, and blast radius.

10. Produce final report
- Generate a Markdown report using `.agents/skills/engineering-quality-audit/references/report-template.md` as a baseline structure.
- The report MUST be extended with any required sections from `Output Requirements` that are missing in the template.
- The report MUST be concise by default.
- The report MUST keep supporting sections short and MUST place most explanatory detail in `Recommended Change Plan`.
- The report SHOULD collapse detailed inventories into a compact summary when separate long sections would only repeat the same information.
- Follow validation points from `.agents/skills/engineering-quality-audit/references/analysis-checklist.md`.

## Decision Rules

- The skill MUST prefer project-native rules over generic assumptions.
- The skill MUST treat human and AI discoverability as first-class quality concerns.
- The skill SHOULD avoid classifying findings as confirmed when evidence is weak.
- The skill MUST separate evidence from assumptions.
- The skill MUST flag uncertainty and request manual confirmation where needed.
- The skill MUST keep recommendations actionable, scoped, and file-specific.
- The skill MUST prioritize architecture-compliant and test-value-positive recommendations before rule-exception candidates.
- The skill MUST NOT recommend test deletion when evidence of redundancy or obsolescence is weak.
- The skill MUST NOT use coverage percentage alone as proof of quality or as the sole justification for a coverage-gap finding.
- The skill MUST prefer structures where the likely change location is predictable from naming, boundaries, and module ownership.
- The skill SHOULD prefer finer-grained files, packages, or modules when the code already contains separable responsibilities or stable seams.
- The skill SHOULD treat reduced token consumption, lower search effort, and smaller review surface as valid supporting benefits when they result from better cohesion and clearer seams.
- The skill MUST NOT recommend decomposition that mainly adds indirection, fragments a cohesive workflow, or weakens discoverability of a single responsibility.
- The skill SHOULD surface simplification opportunities that reduce unnecessary complexity, nesting, redundancy, or over-abstraction when those changes improve local reasoning and maintainability without changing responsibilities.
- The skill MUST NOT recommend a simplification that merges distinct concerns, weakens separation of responsibilities, or makes debugging harder.
- The skill MUST NOT dismiss duplicated same-layer implementations or duplicated tests solely because they still pass or because dependency boundaries are technically correct.
- The skill SHOULD treat unusually high unit-test count around a narrow production surface as a signal that the production unit may be doing too much or that the test topology is poorly organized.
- The skill MUST NOT treat a larger number of tests as inherently better when that volume mainly compensates for poor cohesion, unstable seams, or unclear boundaries.
- When a shared helper, renderer, orchestration primitive, or test helper already exists and a sibling path duplicates the same behavior contract, the skill SHOULD surface at least a low-priority optimization opportunity unless the divergence is explicitly justified by materially different behavior or constraints.

## Output Requirements

The final output MUST be a Markdown report with these sections:
- `Scope and Inputs`
- `Key Findings Summary`
- `Recommended Change Plan`
- `Confidence and Risk`
- `Evidence and Limitations`

The report MUST include concrete file references and architecture unit identifiers or test identifiers wherever possible.
- `Scope and Inputs` MUST stay compact and SHOULD use short bullets only.
- `Scope and Inputs` MUST list both architecture-rule inputs and testing-rule inputs.
- `Scope and Inputs` SHOULD mention navigability, structure, or test-topology evidence sources when those signals materially influenced the audit.
- `Key Findings Summary` MUST stay compact and SHOULD fit each finding or opportunity into one short table row or one short bullet.
- `Key Findings Summary` MUST preserve the recommendation kind (`QUALITY_VIOLATION_CONFIRMED`, `REVIEW_REQUIRED`, `OPTIMIZATION_OPPORTUNITY`, `COVERAGE_GAP`, `TEST_REMOVAL_CANDIDATE`, or `RULE_EXCEPTION_CANDIDATE`) without expanding each kind into its own long section.
- `Key Findings Summary` MUST surface high-search-cost, high-context-cost, or high-test-concentration hotspots when such hotspots materially affect changeability.
- `Recommended Change Plan` MUST be the primary section and MUST be more detailed than the rest of the report.
- Each item in `Recommended Change Plan` MUST explain:
  - the essence of the problem,
  - why solving it is valuable,
  - the general implementation direction,
  - the expected benefit,
  - the main sequencing or dependency considerations when relevant.
- When relevant, each item in `Recommended Change Plan` SHOULD explain how the change improves discoverability, change locality, or local context cost for humans or AI agents.
- If a plan item recommends test deletion, it MUST include the validation step before deletion.
- If a plan item recommends test additions, it MUST include where and what to test.
- `Confidence and Risk` and `Evidence and Limitations` MUST stay brief and MUST NOT restate the full plan.
The skill MUST write the report file to `<project-root>/engineering-quality-audit.md`.

## Safety and Boundaries

- This skill MUST remain analysis-only for source code, tests, and configuration.
- This skill MUST NOT edit repository files except writing the required report artifact defined in `Output Requirements`.
- This skill MUST NOT infer authoritative policy from code patterns alone when explicit project rules exist.
- This skill MAY provide optional follow-up commands for validation, but it MUST label them as optional.
- This skill MUST NOT treat rule-exception candidates or test-removal candidates as default recommendations.

## References

- `.agents/skills/engineering-quality-audit/references/analysis-checklist.md`
- `.agents/skills/engineering-quality-audit/references/report-template.md`

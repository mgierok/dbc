---
name: architecture-audit
description: Audit architectural conformance for a full repository or a user-specified directory and produce actionable findings for architecture-safe optimization, quality improvement, and controlled rule-exception candidates. Use only when the user explicitly asks for an architecture audit, boundary-compliance review, dependency-direction review, layering audit, module-level architecture health check, or pattern-based architecture optimization analysis. This skill MUST NOT auto-trigger in standard implementation workflows.
---

# Architecture Audit

## Purpose

This skill performs a read-only, periodic architecture audit to:
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

- This skill MUST run only on explicit user request.
- This skill MUST NOT be auto-invoked from standard coding or review workflows.
- If invocation intent is not explicit, the skill MUST ask the user to confirm whether the audit should run.
- The skill MUST use `update_plan` from workflow start through completion, updating statuses after each workflow step and finishing with all steps marked as `completed`.

## Inputs

- `scope`:
  - optional path to analyze,
  - default is repository root when no path is provided.
- `project architecture rules`:
  - architecture sections from governance and technical docs,
  - architecture decision records, module conventions, and other local policies if present.

## Audit Workflow

1. Confirm scope and inventory architecture units
- Resolve audit scope (`repo root` or provided directory).
- Inventory relevant units (packages, modules, layers, adapters, services, boundaries).

2. Load architecture expectations
- Discover and load project-specific architecture rules before applying generic heuristics.
- If architecture rules are missing or ambiguous, explicitly state that and continue with heuristic fallback.

3. Build dependency and boundary map
- Map inbound and outbound dependencies between units.
- Map boundary crossings and integration points.
- Identify allowed and forbidden dependency directions for the selected scope.

4. Identify conformance findings
- Use two finding classes:
  - `ARCH_VIOLATION_CONFIRMED`: evidence clearly shows a rule violation.
  - `REVIEW_REQUIRED`: potential violation that needs manual validation.
- Every finding MUST include:
  - file path and unit identifier,
  - violated rule or expected boundary,
  - concrete evidence,
  - risk and impact note,
  - recommended remediation option.

5. Identify architecture-safe optimization and quality opportunities
- Capture opportunities that improve readability, maintainability, cohesion, coupling, and runtime behavior while preserving architecture rules.
- Every opportunity MUST include:
  - target files or units,
  - expected quality or performance gain,
  - why the recommendation remains architecture-compliant,
  - implementation effort estimate.

6. Identify pattern-based optimization opportunities
- Evaluate whether missing design patterns are causing duplicated logic, rigid coupling, unstable change surfaces, or orchestration leakage.
- Pattern recommendations MUST be problem-driven and MUST NOT introduce speculative abstractions.
- Every pattern recommendation MUST include:
  - current problem signal,
  - proposed pattern,
  - expected gain,
  - over-engineering risk note.

7. Identify controlled rule-exception candidates
- The skill MAY propose a `RULE_EXCEPTION_CANDIDATE` only when the expected benefit is concrete and measurable.
- Every candidate MUST include:
  - exact rule that would be bent or broken,
  - explicit reason why strict compliance is costly or harmful in this context,
  - expected benefit of the exception,
  - risk, blast radius, and reversibility note,
  - required safeguards and verification steps.

8. Prioritize recommendations
- Group recommendations into:
  - `High impact`,
  - `Medium impact`,
  - `Low impact`.
- Prioritization SHOULD consider correctness risk, maintainability gain, migration effort, and blast radius.

9. Produce final report
- Generate a Markdown report using `references/report-template.md`.
- Save the generated report to `architecture-audit.md` in the project root directory.
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
- `Confirmed Violations`
- `Review Required Findings`
- `Compliant Optimization Opportunities`
- `Design Pattern Opportunities`
- `Rule-Exception Candidates`
- `Recommended Change Plan`
- `Confidence and Risk`
- `Evidence and Limitations`

The report MUST include concrete file references and architecture unit identifiers wherever possible.
The skill MUST write the report file to `<project-root>/architecture-audit.md`.

## Safety and Boundaries

- This skill MUST remain analysis-only for source code and configuration.
- This skill MUST NOT edit repository files except writing `<project-root>/architecture-audit.md`.
- This skill MUST NOT infer authoritative architecture policy from code patterns alone when explicit project rules exist.
- This skill MAY provide optional follow-up commands for validation, but it MUST label them as optional.
- This skill MUST NOT treat rule-exception candidates as default recommendations.

## References

- `references/analysis-checklist.md`
- `references/report-template.md`

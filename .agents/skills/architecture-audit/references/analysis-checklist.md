# Architecture Audit Checklist

Use this checklist to keep audit output consistent, evidence-based, and actionable.

## 1. Scope and Context

- [ ] Scope is explicit (`repo root` or exact directory path).
- [ ] Architecture unit inventory is complete for selected scope.
- [ ] Relevant boundaries and integration points are identified.

## 2. Project Rules First

- [ ] Project-defined architecture rules were discovered and applied first.
- [ ] If project rules were missing, heuristic fallback was explicitly declared.
- [ ] Any assumptions were listed as assumptions, not facts.

## 3. Dependency and Boundary Mapping

- [ ] Dependency directions are mapped for key architecture units.
- [ ] Boundary crossings are identified and categorized.
- [ ] Allowed versus forbidden dependency flows are explicit.
- [ ] When UI or adapter packages are in scope, sibling implementations of the same conceptual component or workflow were checked for duplicate rendering or orchestration paths.

## 4. Finding Quality Review

For each architecture finding:

- [ ] File path and architecture unit identifier are included.
- [ ] Finding class is set (`ARCH_VIOLATION_CONFIRMED` or `REVIEW_REQUIRED`).
- [ ] Violated rule or boundary expectation is explicit.
- [ ] Evidence is concrete (imports, call paths, ownership crossing, layer leakage).
- [ ] Risk/impact note is included.
- [ ] At least one concrete remediation option is provided.

## 5. Compliant Optimization Opportunity Review

For each compliant optimization opportunity:

- [ ] Opportunity scope is clear.
- [ ] Expected value is stated.
- [ ] Architecture-compliance rationale is explicit.
- [ ] Effort estimate is stated (`Low`, `Medium`, `High`).
- [ ] Suggested sequencing is included.
- [ ] For same-layer duplication findings, the shared behavior contract and current drift risk are explicit.

## 6. Design Pattern Opportunity Review

- [ ] Pattern recommendation is tied to an observed problem.
- [ ] Proposed pattern and expected gain are explicit.
- [ ] Over-engineering risk is explicitly assessed.

## 7. Rule-Exception Candidate Review

For each `RULE_EXCEPTION_CANDIDATE`:

- [ ] Exact rule to be bent or broken is identified.
- [ ] Why the exception may be beneficial is explained concretely.
- [ ] Benefit, risk, and blast radius are explicit.
- [ ] Safeguards and verification steps are included.
- [ ] Candidate is labeled as optional and requiring human approval.

## 8. Prioritization Quality

- [ ] Findings are prioritized (`High`, `Medium`, `Low` impact).
- [ ] Prioritization rationale is consistent and concise.
- [ ] Quick wins are clearly separated from high-effort structural work.

## 9. Output Quality Gate

- [ ] Report follows required section structure.
- [ ] Report is saved to `<project-root>/architecture-audit.md`.
- [ ] Supporting sections stay concise and do not turn into long inventories.
- [ ] `Recommended Change Plan` is the most detailed section in the report.
- [ ] Each plan item explains the problem, expected value, and general implementation direction.
- [ ] Evidence is separated from assumptions and open questions.
- [ ] Limitations are declared (for example partial scope, missing architecture rules).
- [ ] Recommendations are actionable and specific, not generic.
- [ ] Local duplication was not dismissed solely because architecture boundaries remained valid.

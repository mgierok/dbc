# Engineering Quality Audit Checklist

Use this checklist to keep audit output consistent, evidence-based, and actionable.

## 1. Scope and Context

- [ ] Scope is explicit (`repo root` or exact directory path).
- [ ] Code, test, configuration, and CI inventories are complete for the selected scope.
- [ ] Relevant boundaries, integration points, and behavior contracts are identified.
- [ ] Likely change entry points and file-discovery paths are identified for key workflows when relevant.

## 2. Project Rules First

- [ ] Project-defined architecture rules were discovered and applied first.
- [ ] Project-defined testing rules were discovered and applied first.
- [ ] If any rule set was missing, heuristic fallback was explicitly declared.
- [ ] Any assumptions were listed as assumptions, not facts.

## 3. Applicability and Mapping

- [ ] Relevant review areas were classified as `APPLICABLE` or `NOT_APPLICABLE`.
- [ ] Each `NOT_APPLICABLE` area includes a concrete reason.
- [ ] Dependency directions are mapped for key units.
- [ ] Boundary crossings and integration points are identified.
- [ ] Key behaviors are mapped to available tests and test types.
- [ ] Same-layer duplicated behavior paths and duplicated tests were checked when relevant.
- [ ] High-search-cost and high-context-cost hotspots were checked when relevant.
- [ ] Production-to-test mirroring and excessive test concentration were checked when relevant.

## 4. Finding Quality Review

For each `QUALITY_VIOLATION_CONFIRMED` or `REVIEW_REQUIRED` item:

- [ ] File path and architecture unit identifier or test identifier are included.
- [ ] Violated rule or expected property is explicit.
- [ ] Evidence is concrete.
- [ ] Risk or impact note is included.
- [ ] At least one concrete remediation option is provided.

## 5. Optimization Opportunity Review

For each `OPTIMIZATION_OPPORTUNITY`:

- [ ] Opportunity scope is clear.
- [ ] Expected value is stated.
- [ ] Compliance rationale is explicit.
- [ ] Effort estimate is stated (`Low`, `Medium`, `High`).
- [ ] Suggested sequencing is included.
- [ ] For decomposition findings, the seam and non-size justification are explicit.
- [ ] For decomposition findings, the discoverability or context-cost benefit is explicit when relevant.
- [ ] For consolidation findings, the shared behavior contract and drift risk are explicit.

## 6. Coverage Gap and Test Removal Review

For each `COVERAGE_GAP`:

- [ ] Affected module or contract is identified.
- [ ] Missing scenario type is specified.
- [ ] Why the gap matters is explained.
- [ ] Recommendation includes where and what to test.

For each `TEST_REMOVAL_CANDIDATE`:

- [ ] File path and test identifier are included.
- [ ] Reason category is explicit.
- [ ] Evidence is concrete.
- [ ] Risk note is included.
- [ ] Validation before deletion is provided.
- [ ] Confidence level is stated (`High` or `Medium`).

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

- [ ] Report follows the required section structure.
- [ ] Report is saved to `<project-root>/engineering-quality-audit.md`.
- [ ] Supporting sections stay concise and do not turn into long inventories.
- [ ] `Recommended Change Plan` is the most detailed section in the report.
- [ ] Each plan item explains the problem, expected value, and general implementation direction.
- [ ] High-search-cost, high-context-cost, and high-test-concentration hotspots are surfaced when they materially affect changeability.
- [ ] Evidence is separated from assumptions and open questions.
- [ ] Limitations are declared.
- [ ] Recommendations are actionable and specific, not generic.

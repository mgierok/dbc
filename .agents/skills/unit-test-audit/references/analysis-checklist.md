# Unit Test Audit Checklist

Use this checklist to keep audit output consistent, evidence-based, and actionable.

## 1. Scope and Context

- [ ] Scope is explicit (`repo root` or exact directory path).
- [ ] Unit-test file inventory is complete for selected scope.
- [ ] Relevant production modules for audited tests are identified.

## 2. Project Rules First

- [ ] Project-defined testing rules were discovered and applied first.
- [ ] If project rules were missing, heuristic fallback was explicitly declared.
- [ ] Any assumptions were listed as assumptions, not facts.

## 3. Removal Candidate Review

For each removal candidate:

- [ ] File path and test identifier are included.
- [ ] Candidate class is set (`HIGH_CONFIDENCE_REMOVE` or `REVIEW_REQUIRED`).
- [ ] Reason category is explicit (redundant, obsolete, or misaligned with current contract).
- [ ] Evidence is concrete (overlap, dead behavior, outdated assertion logic).
- [ ] Risk note is included.
- [ ] A validation step before deletion is provided.

## 4. Coverage Gap Review

For each coverage gap:

- [ ] Affected module or behavior contract is identified.
- [ ] Missing scenario type is specified (happy, edge, error, transition).
- [ ] Why the gap matters is explained.
- [ ] Recommendation includes where and what to test.

## 5. Prioritization Quality

- [ ] Findings are prioritized (`High`, `Medium`, `Low` impact).
- [ ] Prioritization rationale is consistent and concise.
- [ ] Quick wins are clearly separated from high-effort items.

## 6. Output Quality Gate

- [ ] Report follows required section structure.
- [ ] Report is saved to `<project-root>/unit-test-audit.md`.
- [ ] Evidence is separated from assumptions and open questions.
- [ ] Limitations are declared (for example partial scope, missing project rules).
- [ ] Recommendations are actionable and specific, not generic.

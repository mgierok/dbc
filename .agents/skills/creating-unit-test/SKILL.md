---
name: creating-unit-test
description: Design, create, and modify unit tests as black-box behavioral checks using AAA and FIRST principles with test doubles for external dependencies. Use when asked to add new unit tests, update existing unit tests after production-code changes, refactor brittle or flaky unit tests, improve coverage, fix failing unit tests, review unit-test quality, or explain unit-testing strategy in any language/framework and regardless of methodology (TDD, BDD, or test-after).
---

# Creating Unit Tests

## Goal

Produce deterministic, maintainable unit tests that verify one observable behavior at a time while keeping the unit isolated from external systems.

## Operating Rules

- Treat the unit under test as a black box; validate contracts via inputs, outputs, observable state, and externally visible interactions.
- Follow `FIRST`: keep tests fast, independent, repeatable, self-validating, and timely.
- Structure every test in `Arrange -> Act -> Assert`, with one business action in `Act`.
- Validate one behavior per test; split scenarios instead of combining many assertions for unrelated behaviors.
- Use test doubles (`stub`, `mock`, `fake`, `spy`) for process-external dependencies (database, network, filesystem, clock, randomness, queues, environment).
- Avoid coupling to implementation details (private methods, internal call order unless order is contractually required, incidental data structures).

## Workflow

1. Define behavior contract before writing assertions.
2. Build a scenario matrix: happy path, boundary values, invalid input, dependency failure, and idempotency/ordering when relevant.
3. Prepare minimal fixture and isolate dependencies with test doubles.
4. Implement tests with explicit naming (`should_<expected_behavior>_when_<condition>` or equivalent project convention).
5. Run tests and harden against flakiness (remove time/order/random dependencies).
6. Refactor test structure only after tests are green and readable.

## Pseudo-Code Template

```text
test "should_return_error_when_dependency_times_out":
  Arrange:
    dependency_stub := timeout_error
    unit := create_unit(dependency=dependency_stub)
    input := valid_request()
  Act:
    result := unit.execute(input)
  Assert:
    expect(result.error_code).to_equal("DEPENDENCY_TIMEOUT")
    expect(result.data).to_be_empty()
```

## Methodology Neutrality

- Apply the same rules for any delivery style: TDD, BDD, or test-after implementation.
- If project methodology is unspecified, default to behavior-first test design and explicit scenario coverage.

## References

- Read `references/unit-testing-guide.md` for detailed heuristics, anti-patterns, pseudo-code patterns, and review checklists.

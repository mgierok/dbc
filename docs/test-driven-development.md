# Test-Driven Development (TDD) in Modern Software Development

## Overview

Test-Driven Development (TDD) is a workflow where tests are written before production
code. It guides design, improves confidence, and reduces defects by forcing code to
become testable, modular, and focused on behavior.

At a high level, TDD follows a short feedback loop:
write a failing test, make it pass with minimal code, then refactor.

## Core Principles

1. **Red-Green-Refactor**
   - **Red**: write a test that fails for a specific behavior.
   - **Green**: write the minimal code to make it pass.
   - **Refactor**: improve design without changing behavior.

2. **Test behavior, not implementation**
   - Tests define observable outcomes and contracts.
   - Implementation can change as long as behavior remains correct.

3. **Small, focused steps**
   - Write the smallest test and minimal code to satisfy it.
   - Prefer incremental progress over large jumps.

4. **Fast feedback**
   - Tests should run quickly to encourage frequent execution.

5. **Design through tests**
   - Tests drive API shape and boundaries.
   - Difficult-to-test code reveals coupling and poor design.

## Benefits

- **Better design**: pushes toward smaller components and clear interfaces.
- **Lower defect rate**: errors caught immediately.
- **Safer refactoring**: tests act as a safety net.
- **Documentation**: tests show how code is intended to be used.

## Best Practices

### Keep Tests Readable

- Name tests by behavior, not by function name.
- Use the **AAA pattern**: Arrange, Act, Assert.
- Keep each test focused on one behavior.

### Favor Determinism

- Avoid time-dependent or random behavior in tests.
- Use fixed inputs and predictable outputs.

### Avoid Over-Mocking

- Mock only external boundaries (network, DB, file system).
- Do not mock internal collaborators if it obscures behavior.

### Test at the Right Level

- Most tests should be **unit tests**.
- Use **integration tests** for critical boundaries.
- Use **end-to-end tests** sparingly due to cost.

### Refactor Continuously

- Refactor code and tests after each green stage.
- Remove duplication in both tests and implementation.

### Keep Tests Fast

- Isolate slow dependencies.
- Use in-memory or stubbed alternatives for external systems.

## Example Workflows

### Example 1: Simple Calculator (Generic)

**Goal**: Implement `add(a, b)` with TDD.

1. **Red**
   - Write test: `add(2, 3) == 5`
   - Test fails because `add` does not exist.
2. **Green**
   - Implement minimal `add` returning `a + b`.
3. **Refactor**
   - No changes needed; test remains green.

### Example 2: Business Rule with Validation

**Goal**: Place an order only if it has items.

1. **Red**
   - Test: `placeOrder([])` returns error `ErrEmptyOrder`.
2. **Green**
   - Implement minimal validation to return error when items list is empty.
3. **Refactor**
   - Extract validation into a helper or entity constructor.

### Example 3: External Dependency Boundary

**Goal**: Save user data to a repository.

1. **Red**
   - Test: `CreateUser` calls `UserRepository.Save` and returns created ID.
   - Use a mock repository interface in the test.
2. **Green**
   - Implement `CreateUser` to call the repository and return ID.
3. **Refactor**
   - Simplify or rename components for clarity.

## Common Patterns

### Arrange, Act, Assert (AAA)

```
Arrange: set up inputs and dependencies
Act:     execute the system under test
Assert:  verify the expected outcome
```

### Given, When, Then (BDD Style)

```
Given: a precondition
When:  an action is performed
Then:  the outcome is verified
```

## Practical Tips for Teams

- Agree on naming conventions and test structure.
- Review tests in code review with the same rigor as production code.
- Use CI to run the test suite on every change.
- Keep test data minimal and easy to reason about.

## Common Pitfalls

- **Writing too many tests at once**: keep cycles short.
- **Testing private methods**: test through public behavior.
- **Flaky tests**: avoid timeouts, randomness, and shared state.
- **Skipping refactor**: leaves duplication and weak design.

## TDD vs Traditional Testing

- Traditional testing often happens after code is written.
- TDD makes tests the driver for design and behavior.
- TDD reduces the cost of change by providing early and frequent feedback.

## Summary

TDD is a disciplined workflow that emphasizes small steps, clear behavior, and rapid
feedback. It improves design quality, reduces defects, and makes refactoring safer.
Applied consistently, it becomes a powerful tool for building reliable software.

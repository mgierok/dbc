# Project Rules

## Project Overview

- DBC is a terminal application for browsing and managing databases.
- First supported engine: SQLite; additional engines planned later.
- UI should feel like Midnight Commander with vim-like shortcuts.
- Go is the implementation language.

## Running the App

- Read README.md to understand how to run the app in development environment.

## Language

- Use English for identifiers (variables, functions, types, packages, etc.) and internal documentation.

## Architecture (Clean Architecture + DDD)

This document is authoritative for architecture in this repository. All new code MUST follow Clean Architecture with Domain-Driven Design (DDD).
See `docs/CLEAN_ARCHITECTURE_DDD.md` for full details and examples.

### Architecture Rules (Non-Negotiable)

- The canonical layers are Domain, Application, Interfaces (TUI adapters), and Infrastructure.
- Source code dependencies MUST point inward: Interfaces and Infrastructure depend on Application, Application depends on Domain, Domain depends on nothing.
- Domain MUST NOT import packages from other internal layers or external frameworks.
- Application MAY depend on Domain only. It MUST NOT depend on Infrastructure or Interfaces.
- Interfaces MUST depend on Application (and Domain types) only. It MUST NOT depend on Infrastructure.
- Infrastructure MAY depend on Application and Domain to implement ports. It MUST NOT depend on Interfaces.
- The TUI is an adapter: it handles input/output only and MUST NOT contain business rules or database access logic.
- Database engine interfaces live in Domain or Application; implementations live in Infrastructure.

### Preferred Directory Layout (for new code)

```
cmd/
  dbc/
    main.go
internal/
  domain/
    model/
    service/
    repository/
    event/
    engine/         # engine interfaces
  application/
    usecase/
    port/
    dto/
  interfaces/
    tui/
  infrastructure/
    engine/
    config/
```

### Placement Rules

- Entities, value objects, and aggregates live in `internal/domain/model`.
- Domain services live in `internal/domain/service`.
- Repository interfaces live in `internal/domain/repository`.
- Domain events live in `internal/domain/event`.
- Use cases live in `internal/application/usecase`.
- DTOs live in `internal/application/dto` or under Interfaces if TUI-specific.
- Ports live in `internal/application/port` and must be implemented in Infrastructure.
- TUI adapters (input handling, rendering) live in `internal/interfaces/tui`.
- Database engine implementations live in `internal/infrastructure/engine`.

### DDD Glossary

- Bounded context: A clearly defined domain boundary with its own model, language, and business rules.
- Entity: A domain object defined by identity, not by its attributes. It can change state over time while keeping the same identity.
- Value object: An immutable domain object defined by its attributes. Equality is based on value, not identity.
- Aggregate: A cluster of entities and value objects that must remain consistent. The aggregate root is the only entry point for changes.
- Ubiquitous language: Shared terminology used by the team and reflected in code and docs within a bounded context.
- Repository: An interface that provides access to aggregates. It belongs in the domain or application layer; implementations live in infrastructure.
- Domain service: Stateless domain logic that does not naturally fit in an entity or value object.
- Application service (use case): Orchestrates domain behavior to fulfill a business capability.
- Port: An interface that defines a boundary between the application/domain and infrastructure.
- Adapter: A concrete implementation that translates between external systems and ports.
- Infrastructure: Implementation details such as databases and frameworks.
- DTO: A structure used to move data across boundaries; DTOs must not be used as domain models.

## Database Engine Rules

- SQLite is the first supported engine.
- Adding a new engine must not require changes to TUI code paths beyond wiring.
- Default mode is read-only; write features must be explicit and safe.

## UI/UX Principles

- Keyboard-first navigation; vim-like motions for common actions.
- Consistent panel-based layout inspired by Midnight Commander.
- Clear mode indicators (read-only vs write).

## Development Guidelines

- Write idiomatic Go code following standard conventions and patterns.
- Use interface-driven development with explicit dependency injection.
- Write short, focused functions with single responsibility.
- Handle errors explicitly; avoid global state.
- Adding third-party dependencies requires explicit approval.
- Follow TDD practices described in `docs/TEST_DRIVEN_DEVELOPMENT.md`.

## Approved Libraries (Stage 1)

- `github.com/charmbracelet/bubbletea`
- `modernc.org/sqlite`
- `github.com/pelletier/go-toml/v2`

## Toolchain

- Go 1.25.5

## Performance and Concurrency

- Use goroutines safely with proper synchronization mechanisms.
- Implement goroutine cancellation using context propagation.
- Minimize allocations and profile before optimizing.
- Use benchmarks to track performance regressions.
- Guard shared state with channels or sync primitives.

## Documentation

- Keep business and roadmap docs in `docs/`.
- Update `docs/BRD.md` when scope or roadmap changes.

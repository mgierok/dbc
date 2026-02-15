# Clean Architecture and Domain-Driven Design (DDD)

## Overview

Clean Architecture and Domain-Driven Design (DDD) are complementary approaches for
building maintainable software. Clean Architecture focuses on *how code is structured*
and how dependencies flow. DDD focuses on *what the software models* and how to
capture complex business rules.

Used together, Clean Architecture provides the scaffolding that keeps the domain
model stable, while DDD provides the language and patterns that make the domain model
accurate and expressive.

## Clean Architecture in Brief

Clean Architecture organizes code into layers with a strict dependency rule:
dependencies always point inward. Outer layers can depend on inner layers, but not
the other way around.

### Core Principles

- **Independence of frameworks**: frameworks are tools, not foundations.
- **Testability**: core business rules run without UI, DB, or network dependencies.
- **Independence of UI**: UI changes do not affect business rules.
- **Independence of database**: persistence is a replaceable detail.
- **Independence of external services**: APIs and queues are adapters.

### The Layers

```
-----------------------------------------
| Frameworks and Drivers (outermost)    |
| - Web frameworks, DBs, UI, queues     |
|                                       |
| Interface Adapters                    |
| - Controllers, presenters, gateways  |
|                                       |
| Application Business Rules            |
| - Use cases, application services     |
|                                       |
| Enterprise Business Rules (core)      |
| - Entities, domain models             |
-----------------------------------------
```

### Dependency Rule

```
Outer -> Inner is allowed
Inner -> Outer is forbidden
```

## Domain-Driven Design in Brief

DDD is a modeling approach for complex domains. It emphasizes creating a shared
language between domain experts and developers, and shaping code around that language.

### Strategic DDD (Big-Picture Design)

- **Ubiquitous Language**: shared vocabulary used in code and conversations.
- **Bounded Contexts**: explicit boundaries within which a model is consistent.
- **Context Mapping**: relationships between contexts (e.g., upstream/downstream).
- **Anti-Corruption Layer (ACL)**: translation layer protecting a context from
  external model leakage.

### Tactical DDD (Modeling Patterns)

- **Entity**: identity-based object with lifecycle and invariants.
- **Value Object**: immutable object defined by its attributes.
- **Aggregate**: cluster of entities and value objects with a single root.
- **Repository**: abstraction for loading and persisting aggregates.
- **Domain Service**: domain logic that does not fit a single entity.
- **Factory**: encapsulates complex creation logic.
- **Domain Event**: captures meaningful state changes in the domain.
- **Specification**: reusable, composable business rule.

## How Clean Architecture and DDD Work Together

Clean Architecture provides the structural boundaries that protect the DDD model.
DDD provides the meaning and correctness of the inner layers.

### Mapping DDD Concepts to Clean Architecture

| Clean Architecture Layer         | DDD Placement                                                 |
|----------------------------------|---------------------------------------------------------------|
| Enterprise Business Rules        | Entities, Value Objects, Aggregates, Events, Repositories     |
| Application Business Rules       | Application Services, Use Cases                               |
| Interface Adapters               | Controllers, Presenters, DTO Mappers, ACL                     |
| Frameworks and Drivers           | Web frameworks, DBs, queues, external SDKs                    |

### Complementary Roles

- **Clean Architecture** ensures that the domain model (DDD) remains independent
  of infrastructure and framework choices.
- **DDD** gives the inner layers meaningful structure and behavior, preventing
  anemic models and service-heavy logic.
- Repository interfaces belong to the domain; implementations live in infrastructure.

### Example Interaction Flow

1. A controller (adapter) receives a request.
2. The controller builds an input model and calls a use case.
3. The use case orchestrates a domain operation on aggregates.
4. Repositories load and persist aggregates via interfaces.
5. A presenter maps output to HTTP response or UI model.

## Practical Integration Steps

1. **Define bounded contexts first**
   - Identify core domain language and boundaries.
   - Decide which contexts deserve their own Clean Architecture structure.

2. **Model the domain core**
   - Create entities, value objects, aggregates, and domain events.
   - Keep this layer independent of DB and frameworks.

3. **Define use cases**
   - Create application services that orchestrate domain behavior.
   - Use repository interfaces and domain services as needed.

4. **Build adapters and infrastructure**
   - Implement repositories, controllers, and external API gateways.
   - Use mappers and ACLs to prevent model leakage.

5. **Compose at the edges**
   - Wire dependencies in a composition root (main entry point).

## Example Folder Structure

```
src/
  domain/
    entities/
    value-objects/
    aggregates/
    events/
    services/
    repositories/
  application/
    use-cases/
    ports/
    dto/
  adapters/
    controllers/
    presenters/
    mappers/
    acl/
  infrastructure/
    db/
    http/
    messaging/
    config/
  main/
    composition-root/
```

## Benefits of Using Both

- **Clear separation of concerns**: domain stays stable, infrastructure stays
  replaceable.
- **Testable business rules**: domain and use cases are easy to test in isolation.
- **Model fidelity**: DDD keeps the model aligned with real business behavior.
- **Easier evolution**: adapters and frameworks can be replaced with minimal impact.

## Common Pitfalls

- **Anemic domain**: logic scattered in services instead of entities/aggregates.
- **Leaky abstractions**: ORM objects passed into domain or use cases.
- **Over-layering**: too many abstractions for a simple domain.
- **Ignoring bounded contexts**: forcing a single model across different business
  subdomains.

## When This Combination Fits Best

Use Clean Architecture + DDD when:

- The domain is complex or likely to evolve.
- Multiple interfaces or delivery channels are expected.
- Long-term maintainability is a priority.

It may be too heavy for:

- Small CRUD apps with minimal business logic.
- Short-lived prototypes.

## Summary

Clean Architecture defines *how* dependencies and layers are structured. DDD defines
*what* the core of the system represents. Together they create systems where the
domain model is both expressive and protected from external change.

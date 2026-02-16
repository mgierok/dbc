# DBC Technical Documentation

## Document Control

| Field | Value |
| --- | --- |
| Document Name | DBC Technical Documentation |
| Audience | Junior Software Engineer (primary), all contributors |
| Purpose | Describe how the project is built, structured, tested, and extended |
| Status | Active |
| Last Updated | 2026-02-16 |
| Source of Truth Scope | Current technical state of the codebase |

## Table of Contents

1. [Technical Overview](#1-technical-overview)
2. [Quick Start for Contributors](#2-quick-start-for-contributors)
3. [Project Structure](#3-project-structure)
4. [Architecture Guidelines](#4-architecture-guidelines)
5. [Runtime Flow](#5-runtime-flow)
6. [Technical Decisions](#6-technical-decisions)
7. [Technology Stack and Versions](#7-technology-stack-and-versions)
8. [Testing Strategy and Workflow](#8-testing-strategy-and-workflow)
9. [Feature Delivery Guide](#9-feature-delivery-guide)
10. [Common Technical Constraints](#10-common-technical-constraints)
11. [Reference Documents](#11-reference-documents)
12. [Maintenance Policy](#12-maintenance-policy)

## 1. Technical Overview

DBC is a terminal application written in Go. It currently supports SQLite and follows Clean Architecture with DDD-style boundaries.

Core technical characteristics:

- Terminal UI adapter built on Bubble Tea.
- Layered architecture with inward dependency flow.
- Engine abstraction (`internal/application/port/engine.go`) to support future database engines.
- SQLite implementation in infrastructure layer.
- Staged table changes (insert/update/delete) saved in one transaction.

## 2. Quick Start for Contributors

### 2.1 Prerequisites

- Go toolchain compatible with this repository:
  - `go` directive: `1.25.0`
  - preferred toolchain: `go1.25.5`

### 2.2 Local Setup

1. Create local config directory:
   - macOS/Linux:
   ```bash
   mkdir -p ~/.config/dbc
   ```
2. Copy example config:
   - macOS/Linux:
   ```bash
   cp docs/config.example.toml ~/.config/dbc/config.toml
   ```
3. Edit config file and define at least one `[[databases]]` entry with:
   - `name`
   - `db_path`
   - Default paths:
     - macOS/Linux: `~/.config/dbc/config.toml`
     - Windows: `%APPDATA%\dbc\config.toml`

### 2.3 Run the App

```bash
go run ./cmd/dbc
```

### 2.4 Run Tests

```bash
go test ./...
```

### 2.5 Run Linter

```bash
golangci-lint run ./...
```

## 3. Project Structure

Current source layout:

```text
cmd/
  dbc/
    main.go
internal/
  domain/
    model/
    service/
  application/
    dto/
    port/
    usecase/
  interfaces/
    tui/
  infrastructure/
    config/
    engine/
```

Package responsibilities:

- `cmd/dbc`: composition root and runtime wiring.
- `internal/domain/model`: domain entities/value structures and domain errors.
- `internal/domain/service`: domain-level helper logic (for example value parsing and table sorting).
- `internal/application/usecase`: use case orchestration.
- `internal/application/port`: interfaces (technical boundaries) that infrastructure implements.
- `internal/application/dto`: data structures exchanged with interface adapters.
- `internal/interfaces/tui`: terminal adapter (input, state, rendering).
- `internal/infrastructure/config`: config file loading and validation.
- `internal/infrastructure/engine`: SQLite adapter implementation.

## 4. Architecture Guidelines

This project follows the architecture rules defined in:

- `docs/clean-architecture-ddd.md`
- `AGENTS.md` (Project Rules and layer constraints)

### 4.1 Dependency Direction

Allowed direction:

- `interfaces` -> `application` -> `domain`
- `infrastructure` -> `application` and `domain`

Not allowed:

- `domain` importing `application`, `interfaces`, or `infrastructure`
- `application` importing `interfaces` or `infrastructure`
- `interfaces` importing `infrastructure`

### 4.2 Key Technical Boundaries

- Database access is behind `application/port.Engine`.
- Use cases depend on port interfaces, not concrete database code.
- TUI does not access SQLite directly; it calls use cases.
- Infrastructure provides concrete adapters (`SQLiteEngine`, config loader).

### 4.3 Architecture Rule for New Features

When adding functionality:

1. Start from domain model/service changes if behavior changes domain rules.
2. Add/update use case orchestration.
3. Extend port interfaces only when a new boundary is required.
4. Implement infrastructure adapters for new port behavior.
5. Connect UI adapter to use case, not to infrastructure.

## 5. Runtime Flow

### 5.1 Startup Flow

1. `cmd/dbc/main.go` resolves config path using OS-specific defaults:
   - macOS/Linux: `~/.config/dbc/config.toml`
   - Windows: `%APPDATA%\dbc\config.toml` (fallback `%USERPROFILE%\AppData\Roaming\dbc\config.toml` when `APPDATA` is unset).
2. Startup selector is created with config-management use cases:
   - list configured databases,
   - create/update/delete configured database entry,
   - resolve active config path.
3. Selector UI supports in-session config management (add/edit/delete with delete confirmation) and refreshes entries from config store after each successful mutation.
   Add/edit submit path performs use-case validation in this order: required fields -> SQLite connection check -> config store mutation.
   Active add/edit text input renders a caret (`|`) in the currently focused field.
4. When config has zero entries, selector starts in mandatory first-entry setup:
   - first valid add is required before continue,
   - users can optionally add more entries in the same setup context,
   - normal browsing cannot start until at least one entry exists.
5. User confirms selected database from refreshed selector list.
6. Selected SQLite database is opened and pinged.
7. SQLite engine and runtime table/record use cases are created.
8. Bubble Tea application loop starts (`tui.Run`).
9. If runtime exits with `ErrOpenConfigSelector` (triggered by `:config`), the DB connection is closed and startup selector flow runs again without restarting process.

### 5.2 Main Read Flow

1. TUI model initializes by loading tables.
2. Selected table schema is loaded.
3. Records are loaded in pages (`offset`, `limit`) with optional filter.
4. Additional records load when selection approaches loaded tail.
5. Command entry (`:`) is handled inside TUI model.
6. `:config` routing behavior:
   - if no staged changes: set selector-return signal and exit runtime loop,
   - if staged changes exist: open dirty-state decision popup with `save`, `discard`, `cancel`,
   - `save` executes save flow first and exits to selector only after successful save,
   - `discard` clears staged state and exits to selector immediately,
   - `cancel` keeps runtime session active with staged state unchanged.

### 5.3 Write Flow

1. User stages insert/edit/delete operations in TUI state.
2. On save confirmation, TUI builds `model.TableChanges`.
3. Use case validates payload (`SaveTableChanges`).
4. Engine applies all changes in one transaction:
   - inserts
   - updates (skipping rows also marked for delete)
   - deletes
5. On success:
   - default save action: staged state is cleared and records are reloaded,
   - save triggered from dirty `:config` decision: staged state is cleared and runtime exits to selector instead of reloading records.
6. On failure: rollback occurs and staged state remains.

## 6. Technical Decisions

### 6.1 SQLite-First with Engine Port

Decision:

- Keep current implementation focused on SQLite while coding to an engine interface.

Why:

- Delivers usable value quickly.
- Preserves extension path to future engines with minimal TUI changes.

Where:

- Port: `internal/application/port/engine.go`
- SQLite adapter: `internal/infrastructure/engine/sqlite_engine.go`

### 6.2 Staged Changes Before Save

Decision:

- Do not write immediately on each edit; stage and save explicitly.

Why:

- Safer user workflow.
- Enables undo/redo before persistence.
- Supports single-transaction commit behavior.

Where:

- TUI staged state: `internal/interfaces/tui/model.go`
- Save use case: `internal/application/usecase/save_table_changes.go`

### 6.3 Transactional Save per Table

Decision:

- Apply inserts/updates/deletes for selected table inside one transaction.

Why:

- Prevent partial writes.
- Keep table changes consistent after failures.

Where:

- `internal/infrastructure/engine/sqlite_update.go`

### 6.4 Strict Config Contract

Decision:

- Require at least one configured database, each with `name` and `db_path`.
- Require successful SQLite connection validation before persisting add/edit changes.

Why:

- Prevent ambiguous startup behavior.
- Keep startup errors explicit and actionable.
- Prevent saving unreachable or non-existent database targets in selector configuration.

Where:

- `internal/infrastructure/config/config.go`
- `internal/infrastructure/config/config_test.go`
- `internal/application/port/database_connection_checker.go`
- `internal/application/usecase/config_management.go`
- `internal/infrastructure/engine/sqlite_connection_checker.go`

### 6.5 Operator Allowlist for Filters

Decision:

- Filter operators are limited to a known, validated set.

Why:

- Reduces unsafe SQL clause composition risk.
- Keeps filtering behavior predictable.

Where:

- Operators: `internal/infrastructure/engine/sqlite_operator.go`
- Clause builder: `internal/infrastructure/engine/sqlite_filter.go`

## 7. Technology Stack and Versions

Version source: `go.mod`.

### 7.1 Language and Toolchain

- Go language version: `1.25.0`
- Go toolchain: `go1.25.5`

### 7.2 Direct Dependencies

| Dependency | Version | Purpose |
| --- | --- | --- |
| `github.com/charmbracelet/bubbletea` | `v1.3.10` | Terminal UI framework |
| `modernc.org/sqlite` | `v1.42.2` | SQLite driver |
| `github.com/pelletier/go-toml/v2` | `v2.2.4` | TOML config parsing |

### 7.3 Notes on Indirect Dependencies

- Additional packages in `go.mod` are transitive dependencies and should not be edited manually unless intentionally upgrading dependencies.

### 7.4 Linting Tooling

- Static analysis is configured in `.golangci.yml`.
- The project uses deterministic `golangci-lint` configuration (`linters.default: standard` with explicitly enabled additional linters).
- `revive` comment-enforcement rules are intentionally disabled:
  - `package-comments`
  - `exported`
- `//nolint` usage is restricted:
  - explanation is required
  - specific linter name is required

## 8. Testing Strategy and Workflow

This repository follows TDD expectations documented in:

- `docs/test-driven-development.md`

### 8.1 Current Test Layers

- Domain service tests:
  - `internal/domain/service/*_test.go`
- Application use case tests:
  - `internal/application/usecase/*_test.go`
- Infrastructure adapter tests:
  - `internal/infrastructure/config/*_test.go`
  - `internal/infrastructure/engine/*_test.go`
- TUI behavior tests:
  - `internal/interfaces/tui/*_test.go`

### 8.2 Practical Test Workflow

1. Write/adjust failing test for behavior change.
   - If the target function/type does not exist yet, you may assume `Red` without running the test at this step.
2. Implement minimal code to pass.
3. Refactor safely while tests remain green.
4. Run full suite:
   ```bash
   go test ./...
   ```
5. Run static checks:
   ```bash
   golangci-lint run ./...
   ```

### 8.3 Current Conventions Seen in Tests

- Arrange/Act/Assert structure is used in test bodies.
- Tests target behavior contracts, not private implementation details.
- Integration-like SQLite tests use in-memory databases.

## 9. Feature Delivery Guide

This is a practical checklist for adding a feature correctly.

### 9.1 Step-by-Step

1. Confirm product behavior in `docs/product-documentation.md`.
2. Define technical boundary impact:
   - domain model/service?
   - use case?
   - port/interface?
   - infrastructure adapter?
   - TUI adapter?
3. Add or update tests first (TDD cycle).
4. Implement changes by layer, respecting dependency direction.
5. Run `go test ./...`.
6. Update documentation:
   - `docs/product-documentation.md` for product behavior changes.
   - `docs/technical-documentation.md` for technical changes.
7. Verify naming consistency and terminology.

### 9.2 Typical Change Patterns

- New read capability:
  - extend engine port if needed
  - implement adapter in SQLite engine
  - add/update use case and DTO mapping
  - expose via TUI model/view

- New write capability:
  - update domain change models if needed
  - validate in use case
  - apply in transactional SQLite update path
  - reflect staged state behavior in TUI

## 10. Common Technical Constraints

- Only SQLite engine is implemented today.
- Table edit/delete for persisted rows depends on primary key identity.
- Filter supports one active condition at a time in TUI state.
- Records view currently has no direct shortcut to return to schema view.
- Application behavior assumes config file exists and is valid at startup.

## 11. Reference Documents

- Architecture and DDD details:
  - `docs/clean-architecture-ddd.md`
- TDD principles and workflow:
  - `docs/test-driven-development.md`
- Product behavior source of truth:
  - `docs/product-documentation.md`
- Contributor rules:
  - `AGENTS.md`

## 12. Maintenance Policy

- This document is a source of truth for the technical state of the repository.
- Update this document with every codebase change that affects:
  - architecture
  - boundaries/interfaces
  - dependency versions
  - runtime flow
  - testing workflow
  - development conventions
- Keep wording understandable for a Junior Software Engineer.
- Prefer links to deep-dive documents instead of duplicating long conceptual content.

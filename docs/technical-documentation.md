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
| Related product doc | `docs/product-documentation.md` |

## Table of Contents

1. [Technical Overview](#1-technical-overview)
2. [Runtime Entry Points and Environment Contracts](#2-runtime-entry-points-and-environment-contracts)
3. [Project Structure](#3-project-structure)
4. [Architecture Guidelines](#4-architecture-guidelines)
5. [Runtime Flow](#5-runtime-flow)
6. [Technical Decisions](#6-technical-decisions)
7. [Technology Stack and Versions](#7-technology-stack-and-versions)
8. [Testing Strategy and Coverage](#8-testing-strategy-and-coverage)
9. [Technical Interaction Patterns](#9-technical-interaction-patterns)
10. [Common Technical Constraints](#10-common-technical-constraints)
11. [Reference Documents](#11-reference-documents)
12. [Maintenance Policy](#12-maintenance-policy)
13. [Cross-References to Product Documentation](#13-cross-references-to-product-documentation)

## 1. Technical Overview

DBC is a terminal application written in Go. It currently supports SQLite and follows Clean Architecture with DDD-style boundaries.

Canonical ownership note:

- This document is canonical for implementation details, architecture, runtime, and technical constraints.
- Product behavior intent and user-facing scope are canonical in `docs/product-documentation.md`.

Core technical characteristics:

- Terminal UI adapter built on Bubble Tea.
- Layered architecture with inward dependency flow.
- Engine abstraction (`internal/application/port/engine.go`) to support future database engines.
- SQLite implementation in infrastructure layer.
- Staged table changes (insert/update/delete) saved in one transaction.

## 2. Runtime Entry Points and Environment Contracts

### 2.1 Runtime Entry Point

- Process entry point is `cmd/dbc/main.go`.
- `main.go` is the composition root for:
  - startup CLI argument parsing (`-d` / `--database`),
  - config path resolution,
  - selector/config-management use case wiring,
  - database engine creation,
  - TUI runtime startup.

### 2.2 Runtime Configuration Contract

- Default config paths:
  - macOS/Linux: `~/.config/dbc/config.toml`
  - Windows: `%APPDATA%\\dbc\\config.toml`
- Config contract requires `[[databases]]` entries with:
  - `name`
  - `db_path`
- Empty config states (`missing file`, `empty file`, `databases = []`) are mapped to mandatory first-entry setup.
- Malformed config content is treated as startup error and blocks runtime initialization.

### 2.3 Runtime Lifecycle Boundaries

- Startup supports two paths:
  - default selector-first path (no direct-launch argument),
  - direct-launch path using `-d` / `--database` that validates connectivity before runtime startup.
- Direct-launch startup resolves target identity against configured entries using normalized SQLite path comparison and reuses the first configured match in config order.
- Direct-launch validation failure prints actionable error output and exits non-zero without selector fallback.
- Runtime can return to selector without process restart via `ErrOpenConfigSelector`.
- Active DB connection is explicitly closed before selector re-entry.

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

## 5. Runtime Flow

### 5.1 Startup Flow

1. `cmd/dbc/main.go` parses startup CLI arguments.
   - Supported direct-launch aliases: `-d <db_path>` and `--database <db_path>`.
   - Invalid startup arguments return clear error output and terminate startup.
2. `cmd/dbc/main.go` resolves config path using OS-specific defaults:
   - macOS/Linux: `~/.config/dbc/config.toml`
   - Windows: `%APPDATA%\dbc\config.toml`
3. Startup selector dependencies are created with config-management use cases:
   - list configured databases,
   - create/update/delete configured database entry,
   - resolve active config path.
4. If direct-launch argument is provided, startup attempts direct SQLite open/ping first (before selector UI):
   - startup first resolves direct-launch target against configured entries using normalized SQLite path identity,
   - when normalized match exists, startup reuses that configured entry identity (deterministic first match by config order),
   - success: runtime starts immediately and selector is skipped for initial startup,
   - failure: startup prints actionable failure message and exits non-zero (no selector fallback).
5. For selector-first path (no direct-launch argument), selector UI supports in-session config management (add/edit/delete with delete confirmation) and refreshes entries from config store after each successful mutation.
   Add/edit submit path performs use-case validation in this order: required fields -> SQLite connection check -> config store mutation.
   Active add/edit text input renders a caret (`|`) in the currently focused field.
6. Empty config state (`missing file`, `empty file`, or `databases = []`) starts selector in mandatory first-entry setup:
   - first valid add is required before continue,
   - users can optionally add more entries in the same setup context,
   - normal browsing cannot start until at least one entry exists.
   Pressing `Esc` in this setup cancels startup and exits application loop before DB open.
   Malformed config content (for example malformed TOML, unknown key shape, or invalid entry structure) is still treated as startup error.
7. User confirms selected database from refreshed selector list.
8. Selected SQLite database is opened and pinged.
   Startup DB open and config-entry connection checks reuse the same infrastructure connection-open helper (`internal/infrastructure/engine.OpenSQLiteDatabase`) to keep validation behavior and errors consistent.
   If open/ping fails, startup loop reopens selector with status error and preferred selection set to failed connection string.
   Runtime does not start until user selects a reachable entry or updates configuration.
9. SQLite engine and runtime table/record use cases are created.
10. Bubble Tea application loop starts (`tui.Run`).
11. If runtime exits with `ErrOpenConfigSelector` (triggered by `:config`), the DB connection is closed and startup selector flow runs again without restarting process.

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

## 8. Testing Strategy and Coverage

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

### 8.2 Coverage Boundaries

- Domain tests validate domain-level rules and helper behavior.
- Application tests validate use-case orchestration and port contracts.
- Infrastructure tests validate adapter behavior for config and SQLite integration.
- TUI tests validate user-visible state transitions and input handling.

### 8.3 Current Conventions Seen in Tests

- Arrange/Act/Assert structure is used in test bodies.
- Tests target behavior contracts, not private implementation details.
- Integration-like SQLite tests use in-memory databases.

## 9. Technical Interaction Patterns

### 9.1 Read Interaction Pattern

- TUI requests table/schema/records through application use cases.
- Use cases call `application/port.Engine`.
- Infrastructure engine adapter maps calls to SQLite queries.
- Retrieved records are adapted into DTOs consumed by TUI state.

### 9.2 Write Interaction Pattern

- TUI accumulates staged changes in session state.
- Save path builds `model.TableChanges`.
- `SaveTableChanges` validates payload and delegates persistence to engine port.
- SQLite adapter executes inserts/updates/deletes in one transaction.

## 10. Common Technical Constraints

This section captures implementation constraints. For user-facing wording of constraints/non-goals, see `docs/product-documentation.md#10-known-constraints-and-non-goals`.

- Only SQLite engine is implemented today.
- Table edit/delete for persisted rows depends on primary key identity.
- Filter pipeline currently applies one active predicate for the selected table at a time.
- Missing or empty config file is tolerated at startup and mapped to mandatory first-entry setup.
- Invalid config content still stops startup with explicit error.

## 11. Reference Documents

- Architecture and DDD details:
  - `docs/clean-architecture-ddd.md`
- TDD principles:
  - `docs/test-driven-development.md`
- Product behavior source of truth:
  - `docs/product-documentation.md`

## 12. Maintenance Policy

- This document is a source of truth for the technical state of the repository.
- Keep technical statements aligned with current code paths and runtime behavior.
- Keep wording understandable for a Junior Software Engineer.
- Prefer links to deep-dive documents instead of duplicating long conceptual content.
- Keep product intent in product documentation and reference it instead of restating it here.

## 13. Cross-References to Product Documentation

- Product scope and non-goals: `docs/product-documentation.md#4-current-product-scope`, `docs/product-documentation.md#10-known-constraints-and-non-goals`
- User journey and behavior intent: `docs/product-documentation.md#6-end-to-end-user-journey`
- Capability-level behavior: `docs/product-documentation.md#7-functional-specification-current-state`
- User-visible interaction contract: `docs/product-documentation.md#8-keyboard-interaction-model`
- Safety and governance intent: `docs/product-documentation.md#9-data-safety-and-change-governance`

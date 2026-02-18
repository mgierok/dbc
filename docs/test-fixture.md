# Test Fixture

## Test Database

- Fixture source file: `scripts/test.db`.
- Scope: local/manual test runs for startup and navigation scenarios.
- Domain modeled by fixture: `customers`, `categories`, `products`, `orders`, `order_items`.

## Database Structure

### Tables

- `customers`
  - `id` (PK), `email` (UNIQUE, NOT NULL), `full_name` (NOT NULL, CHECK), `phone` (NULLable), `loyalty_points` (DEFAULT `0`), `is_active` (DEFAULT `1`, CHECK), `created_at` (DEFAULT timestamp).
- `categories`
  - `id` (PK), `name` (UNIQUE, NOT NULL), `description` (NOT NULL, DEFAULT `''`, CHECK length).
- `products`
  - `id` (PK), `sku` (UNIQUE, NOT NULL), `name` (NOT NULL), `category_id` (FK), `price_cents` (CHECK), `notes` (NULLable), `weight_kg` (`REAL`), `metadata` (`BLOB`), `is_discontinued` (DEFAULT `0`, CHECK).
- `orders`
  - `id` (PK), `customer_id` (FK), `status` (NOT NULL, enum-like CHECK), `placed_at` (NULLable), `note` (NOT NULL, DEFAULT `''`), `total_cents` (DEFAULT `0`, CHECK).
- `order_items`
  - `id` (PK), `order_id` (FK with `ON DELETE CASCADE`), `product_id` (FK), `quantity` (CHECK), `unit_price_cents` (CHECK), `discount_percent` (`REAL`), UNIQUE(`order_id`, `product_id`).

## Table Relationships

- `products.category_id` -> `categories.id`
- `orders.customer_id` -> `customers.id`
- `order_items.order_id` -> `orders.id` (`ON DELETE CASCADE`)
- `order_items.product_id` -> `products.id`

## Edge-Case Coverage Contract

| Category | Fixture Evidence |
| --- | --- |
| `null` | `customers.phone` (`id=1`), `products.notes` (`id=3`), `orders.placed_at` (`id=2`) |
| `default` | `orders.note` default `''`, `customers.loyalty_points` default `0` |
| `not-null` | `customers.email`, `products.name`, `orders.status`, `order_items.quantity` |
| `unique` | `customers.email`, `categories.name`, `products.sku`, `order_items(order_id, product_id)` |
| `foreign-key` | `products.category_id`, `orders.customer_id`, `order_items.order_id`, `order_items.product_id` |
| `check` | non-negative points/cents, `orders.status` enum-like check, boolean-like `is_active`/`is_discontinued` |
| `empty values` | `customers.phone` (`id=2`), `categories.description` (`id=2`) |
| `long values` | long text in `products.notes` (`id=5`) |
| `varied SQLite types` | `INTEGER`, `REAL`, `TEXT`, `BLOB`, `NULL` |

## Startup Scripts

Run from repository root.

| Script | Run | Use When |
| --- | --- | --- |
| [`scripts/start-direct-launch.sh`](../scripts/start-direct-launch.sh) | `bash scripts/start-direct-launch.sh` | You want immediate runtime start with `-d` and no selector step. |
| [`scripts/start-selector-from-config.sh`](../scripts/start-selector-from-config.sh) | `bash scripts/start-selector-from-config.sh` | You want selector startup backed by a valid temporary config file. |
| [`scripts/start-without-database.sh`](../scripts/start-without-database.sh) | `bash scripts/start-without-database.sh` | You want startup without `-d` and without configured databases (mandatory first-entry setup). |

### Output and Cleanup

- Every startup script prints `TMP_ROOT=...`.
- `scripts/start-without-database.sh` additionally prints `TMP_DB=...` for follow-up tests.
- Cleanup command:
  - Script: [`scripts/cleanup-temp-environment.sh`](../scripts/cleanup-temp-environment.sh)
  - Run: `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`

## Example Test Scenario

Goal: verify direct-launch startup against fixture database.

1. Run `bash scripts/start-direct-launch.sh`.
2. Confirm app opens directly in main view (no selector).
3. Confirm table list contains: `categories`, `customers`, `order_items`, `orders`, `products`.
4. Press `j`, then `Enter` to open `customers` records.
5. Confirm rows include: `alice@example.com`, `bob@example.com`, `charlie@example.com`.
6. Press `q` to exit.
7. Run cleanup using printed `TMP_ROOT`:
   - `bash scripts/cleanup-temp-environment.sh <TMP_ROOT>`

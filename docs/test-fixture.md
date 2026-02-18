# Test Fixture Contract

## Canonical Fixture

- Canonical local fixture database path: `docs/test.db`.
- This fixture is the default dataset for local/manual validation and agent-assisted checks in PRD-4 workflows.

## Relation Map

The fixture models a compact order domain with explicit foreign keys:

- `customers` (`id` PK)
- `categories` (`id` PK)
- `products` (`id` PK, `category_id` FK -> `categories.id`)
- `orders` (`id` PK, `customer_id` FK -> `customers.id`)
- `order_items` (`id` PK, `order_id` FK -> `orders.id`, `product_id` FK -> `products.id`, unique pair `order_id + product_id`)

## Edge-Case Coverage Contract

Required categories and where they are represented:

| Category | Fixture Evidence |
| --- | --- |
| `null` | `customers.phone` (`id=1`), `products.notes` (`id=3`), `orders.placed_at` (`id=2`) |
| `default` | `orders.note` default `''` (omitted on inserts), `customers.loyalty_points` default `0` |
| `not-null` | `customers.email`, `products.name`, `orders.status`, `order_items.quantity` |
| `unique` | `customers.email`, `categories.name`, `products.sku`, `order_items(order_id, product_id)` |
| `foreign-key` | `products.category_id`, `orders.customer_id`, `order_items.order_id`, `order_items.product_id` |
| `check` | non-negative cents/points, enum-like `orders.status`, boolean-like flags (`is_active`, `is_discontinued`) |
| `empty values` | empty string in `customers.phone` (`id=2`) and `categories.description` (`id=2`) |
| `long values` | long text in `products.notes` (`id=5`) |
| `varied SQLite types` | `INTEGER`, `REAL`, `TEXT`, `BLOB`, and `NULL` values across tables |

## Small-Fixture Thresholds

PRD-4 limits for this fixture:

- File size: `<= 1 MiB`
- Total rows across all tables: `<= 300`
- Max rows per table: `<= 120`

## Verification Commands

```bash
# FK integrity
sqlite3 docs/test.db "PRAGMA foreign_keys=ON; PRAGMA foreign_key_check;"

# Per-table row counts
sqlite3 docs/test.db "
SELECT 'customers', COUNT(*) FROM customers
UNION ALL SELECT 'categories', COUNT(*) FROM categories
UNION ALL SELECT 'products', COUNT(*) FROM products
UNION ALL SELECT 'orders', COUNT(*) FROM orders
UNION ALL SELECT 'order_items', COUNT(*) FROM order_items;
"

# Fixture file size (bytes)
wc -c docs/test.db
```

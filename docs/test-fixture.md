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

## Tmp Startup Playbooks

Run all commands from repository root.

### Shared Tmp Bootstrap

```bash
TMP_ROOT="$(mktemp -d)"
TMP_HOME="$TMP_ROOT/home"
TMP_DB="$TMP_ROOT/test.db"
DBC_BIN="$TMP_ROOT/dbc"

mkdir -p "$TMP_HOME/.config/dbc"
cp docs/test.db "$TMP_DB"
go build -o "$DBC_BIN" ./cmd/dbc
```

### Variant 1: Direct Launch via `-d`

Specific behavior:

- Startup validates the provided SQLite path before runtime starts.
- On success, runtime opens directly and startup selector is bypassed.
- On failure, startup exits non-zero with direct-launch guidance (no selector fallback).

When to use:

- Fast local smoke checks against a known fixture path.
- Agent-driven runs where selector interaction is intentionally skipped.

Executable commands:

```bash
# Happy path: direct launch into fixture, then press q in the app
HOME="$TMP_HOME" "$DBC_BIN" -d "$TMP_DB"

# Negative path: invalid direct-launch target should fail with startup guidance
HOME="$TMP_HOME" "$DBC_BIN" -d "$TMP_ROOT/missing.db"
```

### Variant 2: Startup via Config File

Specific behavior:

- Startup reads config-backed entries from `$HOME/.config/dbc/config.toml`.
- Without `-d`, selector-first startup is used and config entries are available for selection.
- Malformed config content is treated as startup error and blocks runtime initialization.

When to use:

- Validation of standard operator flow where startup is driven by persisted config.
- Reproducing selector-based startup behavior with deterministic tmp inputs.

Executable commands:

```bash
# Happy path: valid tmp config + selector-first startup, then press q in selector
cat > "$TMP_HOME/.config/dbc/config.toml" <<EOF
[[databases]]
name = "fixture"
db_path = "$TMP_DB"
EOF
HOME="$TMP_HOME" "$DBC_BIN"

# Negative path: malformed config should fail during startup
cat > "$TMP_HOME/.config/dbc/config.toml" <<'EOF'
[[databases]]
name = "fixture
db_path = "/tmp/fixture.db"
EOF
HOME="$TMP_HOME" "$DBC_BIN"
```

### Variant 3: Startup Without Database Parameter

Specific behavior:

- Startup is selector-first whenever no `-d`/`--database` argument is provided.
- With valid tmp config, selector can open configured entries and continue to runtime.
- With missing config, startup enters mandatory first-entry setup; user must add an entry or cancel (`Esc`) to recover.

When to use:

- Validating default startup behavior expected by first-time and selector-driven sessions.
- Reproducing no-parameter startup recovery path for empty/missing config environments.

Executable commands:

```bash
# Happy path: no parameter with valid config present, then Enter + q
cat > "$TMP_HOME/.config/dbc/config.toml" <<EOF
[[databases]]
name = "fixture"
db_path = "$TMP_DB"
EOF
HOME="$TMP_HOME" "$DBC_BIN"

# Negative path: intentionally missing config triggers mandatory first-entry setup; Esc cancels startup
rm -f "$TMP_HOME/.config/dbc/config.toml"
HOME="$TMP_HOME" "$DBC_BIN"
```

## Manual Scenario (PRD-4 Task 3)

Startup method binding:

- This scenario uses `Variant 1: Direct Launch via -d`.

Preconditions:

- Run `Shared Tmp Bootstrap` once from this document.

Execution steps and expected observations:

1. Start DBC with direct launch:
   - Command: `HOME="$TMP_HOME" "$DBC_BIN" -d "$TMP_DB"`
   - Expected observation:
     - Startup goes directly to main view (no selector-first screen).
     - Left panel table list includes: `categories`, `customers`, `order_items`, `orders`, `products`.
2. Navigate to `customers` records:
   - Keys: `j`, then `Enter`
   - Expected observation:
     - Right panel switches to `Records`.
     - `customers` rows include emails:
       - `alice@example.com`
       - `bob@example.com`
       - `charlie@example.com`
3. Exit:
   - Key: `q`
   - Expected observation: process exits cleanly (status `0`) and terminal returns to shell prompt.

Pass/fail criteria:

- PASS: every expected observation in steps 1-3 is satisfied in one continuous run.
- FAIL: any expected observation is missing or mismatched.

Failure reporting notes:

- Record:
  - executed startup command,
  - failing step number,
  - observed output/screen state,
  - expected output/screen state,
  - whether rerun reproduced the same mismatch.

Rerun notes:

- Re-run `Shared Tmp Bootstrap` before each retry.
- Keep startup method fixed (`-d`) and fixture source fixed (`docs/test.db`).
- Use `Tmp Cleanup` after each run to avoid cross-run residue.

### Tmp Cleanup

```bash
rm -rf "$TMP_ROOT"
```

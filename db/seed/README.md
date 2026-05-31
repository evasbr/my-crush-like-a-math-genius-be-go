# Database SQL Seeder Mechanism

This package implements a programmatic SQL-based seeding mechanism for populating the database with initial/dummy data.

It works similarly to a database migration tool, with built-in version tracking to prevent duplicate execution of seeds in production environments.

---

## How It Works

1.  All SQL seed files are stored in the `db/seed/` folder.
2.  Seed files use the format:
    *   `*.up.sql` (Applies the seed data).
    *   `*.down.sql` (Rolls back / removes the seed data).
3.  The seeder creates a tracking table named `seeder_history` in your database.
4.  Before executing a seed, it checks `seeder_history`. If the seed's version (prefix) has already been applied, it skips it.
5.  All seed files are embedded into the compiled Go binary using Go's `//go:embed` package, meaning you do not need to copy physical `.sql` files into production containers.

---

## How to Create a New Seeder

1.  Run the CLI generator command in the root directory:
    ```bash
    go run main.go -seed-create="seed_products"
    ```
    This will automatically generate a timestamp-prefixed pair of files inside `db/seed/` (e.g. `20260531145812_seed_products.up.sql` and `20260531145812_seed_products.down.sql`).
2.  Write your SQL DML statement in `*.up.sql`:
    ```sql
    INSERT INTO tb_product (id, name, price, quantity) VALUES 
    ('p001', 'Milo Premium', 15000, 100)
    ON CONFLICT (id) DO NOTHING;
    ```
3.  Write the corresponding reversal SQL query in `*.down.sql`:
    ```sql
    DELETE FROM tb_product WHERE id = 'p001';
    ```

---

## Usage

You can trigger the seeder using the command-line flags in the root of the project:

### 1. Apply All Pending Seeders
```bash
go run main.go -seed
```

### 2. Rollback the Last Seeder
```bash
go run main.go -seed-rollback
```

### 3. Generate New Seeder Files (with Timestamp)
```bash
go run main.go -seed-create="nama_seeder_anda"
```

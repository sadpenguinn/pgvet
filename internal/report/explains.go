package report

// explains holds detailed explanations for danger rules.
var explains = map[string]string{
	"create-index-no-concurrently": `CREATE INDEX acquires a ShareLock that blocks all writes (INSERT/UPDATE/DELETE)
  for the entire duration of the build. On large tables this causes downtime.

  Fix: use CREATE INDEX CONCURRENTLY — builds without a write lock, requires
  two table scans. Cannot run inside a transaction block.`,

	"drop-index-no-concurrently": `DROP INDEX acquires an ACCESS EXCLUSIVE lock, blocking all reads and writes
  until the operation completes.

  Fix: use DROP INDEX CONCURRENTLY. Cannot run inside a transaction block.`,

	"add-column-not-null": `On PostgreSQL < 11 adding a NOT NULL column without DEFAULT triggers a full
  table rewrite with ACCESS EXCLUSIVE lock held for the duration.
  On PostgreSQL >= 11 this is safe only with a non-volatile DEFAULT.

  Fix (safe for all versions):
    1. ALTER TABLE t ADD COLUMN col TYPE DEFAULT <value>;
    2. UPDATE t SET col = <value> WHERE col IS NULL;  -- backfill in batches
    3. ALTER TABLE t ALTER COLUMN col SET NOT NULL;`,

	"set-not-null": `SET NOT NULL performs a full sequential scan to verify no existing rows
  contain NULL. Holds ACCESS EXCLUSIVE lock for the duration.

  Fix: use a CHECK constraint instead:
    1. ALTER TABLE t ADD CONSTRAINT chk_col_not_null CHECK (col IS NOT NULL) NOT VALID;
    2. ALTER TABLE t VALIDATE CONSTRAINT chk_col_not_null;  -- ShareUpdateExclusiveLock`,

	"add-foreign-key-no-valid": `Without NOT VALID, the constraint scans the entire table to validate all
  existing rows, holding ACCESS EXCLUSIVE lock for the duration.

  Fix:
    1. ALTER TABLE t ADD CONSTRAINT fk ... FOREIGN KEY (...) REFERENCES ... NOT VALID;
    2. ALTER TABLE t VALIDATE CONSTRAINT fk;  -- ShareRowExclusiveLock, allows reads`,

	"drop-table": `DROP TABLE is irreversible — all rows, indexes, and constraints are
  permanently deleted. There is no undo.

  Fix: ensure a recent backup exists before running. Consider renaming the
  table first as a safety step, then dropping after confirming no impact.`,

	"drop-column": `DROP COLUMN permanently destroys the column and all its data. On older
  PostgreSQL versions it also triggers a table rewrite.

  Fix: if unsure, mark the column as unused at the application level first.
  Ensure backups exist before dropping.`,

	"truncate": `TRUNCATE removes all rows irreversibly and holds ACCESS EXCLUSIVE lock
  for the duration, blocking all reads and writes.

  Fix: if incremental cleanup is needed, use batched DELETE instead.
  Ensure backups exist before running in production.`,

	"lock-table": `Explicit LOCK TABLE acquires a lock that queues all conflicting transactions,
  potentially causing a cascade of blocked queries and downtime.

  Fix: rely on implicit locking from normal DML operations. If a lock is
  truly needed, use the least restrictive mode and keep the transaction short.`,

	"rename": `RENAME requires ACCESS EXCLUSIVE lock and immediately breaks any code,
  queries, migrations, or external tools that reference the old name.

  Fix:
    1. Add the new name (new column or view with new name).
    2. Migrate all references in code and dependent services.
    3. Remove the old name only after all consumers are updated.`,

	"change-column-type": `Changing a column's type rewrites the entire table with ACCESS EXCLUSIVE
  lock held for the full duration. On large tables this causes downtime.

  Fix:
    1. ADD COLUMN new_col new_type;
    2. Backfill: UPDATE t SET new_col = old_col::new_type;  -- in batches
    3. Swap application to use new_col.
    4. DROP COLUMN old_col;`,

	"redundant-index": `A redundant index duplicates the prefix of another index on the same table.
  PostgreSQL never uses the narrower index when a wider one already covers
  the same leading columns, so the redundant index wastes disk space and
  slows down every write (INSERT/UPDATE/DELETE must maintain all indexes).

  Common cases:
  • INDEX(a)       is redundant when INDEX(a, b) exists (btree prefix rule)
  • INDEX(a)       is redundant when UNIQUE(a) exists (unique is also scannable)
  • INDEX(a, b, c) is redundant when INDEX(a, b, c, d) exists

  Fix: DROP the redundant index.
  Note: UNIQUE indexes are never redundant — they enforce a constraint.
  Partial indexes (WHERE ...) are only redundant when the superseder covers
  the same or broader predicate.`,
}

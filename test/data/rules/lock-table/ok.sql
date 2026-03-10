BEGIN;
UPDATE orders SET status = 'processed' WHERE id = 1;
COMMIT;

CREATE TABLE orders (id SERIAL PRIMARY KEY, user_id INTEGER NOT NULL);

-- sqlint:disable create-index-no-concurrently
CREATE INDEX idx_orders_user_id ON orders (user_id);
-- sqlint:enable create-index-no-concurrently

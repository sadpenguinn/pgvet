CREATE TABLE orders (id SERIAL PRIMARY KEY, user_id INTEGER NOT NULL);

CREATE INDEX idx_orders_user_id ON orders (user_id); -- nolint:create-index-no-concurrently

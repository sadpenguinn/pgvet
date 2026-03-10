CREATE TABLE orders (
    id         SERIAL  PRIMARY KEY,
    user_id    INTEGER NOT NULL,
    status     TEXT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id     ON orders (user_id); -- nolint:redundant-index
CREATE INDEX idx_orders_user_status ON orders (user_id, status);

CREATE TABLE orders (
    id         SERIAL  PRIMARY KEY,
    user_id    INTEGER NOT NULL,
    status     TEXT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_status ON orders (user_id, status);
CREATE INDEX idx_orders_created_at  ON orders (created_at);

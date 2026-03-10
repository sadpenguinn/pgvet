-- +goose Up
-- +goose StatementBegin

CREATE TABLE users (
    id         SERIAL                   PRIMARY KEY,
    email      CHARACTER VARYING(255)   NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
    id         SERIAL                   PRIMARY KEY,
    user_id    INTEGER                  NOT NULL,
    total      NUMERIC(12, 2)           NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders (user_id);

ALTER TABLE orders
    ADD CONSTRAINT fk_orders_user
    FOREIGN KEY (user_id) REFERENCES users (id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE orders;
DROP TABLE users;

-- +goose StatementEnd

CREATE TABLE products (
    id          SERIAL                   PRIMARY KEY,
    name        CHARACTER VARYING(255)   NOT NULL,
    price       NUMERIC(10, 2)           NOT NULL,
    category_id INTEGER,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_category ON products (category_id);

ALTER TABLE products ADD COLUMN stock_count INTEGER NOT NULL;

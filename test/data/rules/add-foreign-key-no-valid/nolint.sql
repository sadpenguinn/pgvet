ALTER TABLE orders -- nolint:add-foreign-key-no-valid
    ADD CONSTRAINT fk_orders_user
    FOREIGN KEY (user_id) REFERENCES users (id);

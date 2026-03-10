ALTER TABLE orders
    ADD CONSTRAINT fk_orders_user
    FOREIGN KEY (user_id) REFERENCES users (id); -- sqlint:disable add-foreign-key-no-valid

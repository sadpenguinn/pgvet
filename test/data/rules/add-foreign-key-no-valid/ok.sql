ALTER TABLE orders
    ADD CONSTRAINT fk_orders_user
    FOREIGN KEY (user_id) REFERENCES users (id) NOT VALID;

ALTER TABLE orders VALIDATE CONSTRAINT fk_orders_user;

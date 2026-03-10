ALTER TABLE users ADD CONSTRAINT chk_users_email_not_null
    CHECK (email IS NOT NULL) NOT VALID;

ALTER TABLE users VALIDATE CONSTRAINT chk_users_email_not_null;

-- safe pattern: add nullable → fill → set not null
ALTER TABLE tag ADD COLUMN kind BIGINT;

UPDATE tag SET kind = 1;

ALTER TABLE tag ALTER COLUMN kind SET NOT NULL;

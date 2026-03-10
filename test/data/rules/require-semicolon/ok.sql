CREATE TABLE users (
    id    INTEGER NOT NULL,
    email TEXT    NOT NULL
);

-- +goose StatementBegin
INSERT INTO tag (text, uses)
VALUES
    ('foo', 10),
    ('bar', 10)
-- +goose StatementEnd

CREATE TABLE users (
    id    SERIAL  PRIMARY KEY,
    email TEXT    NOT NULL
);

CREATE INDEX CONCURRENTLY idx_users_email ON users (email);

-- text as a column name — must not trigger keyword-case
SELECT id AS text FROM users;

UPDATE users SET text = 'hello' WHERE text = 'world';

SELECT t.text FROM users t;

CREATE TABLE tag (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text TEXT UNIQUE NOT NULL
);

CREATE INDEX CONCURRENTLY tag_text_pattern_idx ON tag(text text_pattern_ops);
CREATE INDEX CONCURRENTLY tag_text_uses_idx ON tag(text, uses DESC);

DELETE FROM tag WHERE text IN ('a', 'b');

SELECT id FROM tag WHERE text LIKE '%foo%';

SELECT id FROM tag WHERE text IS NULL;

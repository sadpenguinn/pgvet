create table users ( -- nolint:keyword-case
    id    serial  primary key, -- nolint:keyword-case
    email text    not null -- nolint:keyword-case
);

create index idx_users_email on users (email); -- nolint:keyword-case

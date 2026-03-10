-- sqlint:disable keyword-case
create table users (
    id    serial  primary key,
    email text    not null
);

create index idx_users_email on users (email);
-- sqlint:enable keyword-case

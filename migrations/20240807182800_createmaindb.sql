-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table if not exists users (
    id bigserial primary key,
    username varchar(255) unique not null,
    email varchar(255) unique not null,
    password varchar(255) not null,
    created_at timestamptz not null,
    updated_at timestamptz
);

create index users_username_idx on users(LOWER(username));
create index users_email_idx on users(LOWER(email));

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop index if exists users_username_idx;
drop index if exists users_email_idx;

drop table if exists users;
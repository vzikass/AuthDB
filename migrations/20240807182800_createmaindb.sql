-- +goose Up
-- +goose StatementBegin
create table if not exists users (
    id bigserial primary key,
    username varchar(255) unique not null,
    email varchar(255) unique not null,
    password varchar(255) not null,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
);
-- +goose StatementEnd

create index users_username_idx on users(LOWER(username));
create index users_email_idx on users(LOWER(email));

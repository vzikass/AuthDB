-- +goose Up
-- +goose StatementBegin
create table if not exists users (
    id bigserial primary key,
    username varchar(255) unique not null,
    email varchar(255) unique not null,
    password varchar(255) not null,
    role varchar(50) not null default 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd
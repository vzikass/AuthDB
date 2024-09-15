-- +goose Up
-- +goose StatementBegin
DROP DATABASE IF EXISTS maindb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE DATABASE maindb;
-- +goose StatementEnd

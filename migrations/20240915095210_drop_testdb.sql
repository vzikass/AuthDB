-- +goose Up
-- +goose StatementBegin
DROP DATABASE IF EXISTS testdb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE DATABASE testdb;
-- +goose StatementEnd

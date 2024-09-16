-- +goose Up
-- +goose StatementBegin
DROP DATABASE IF EXISTS AuthDB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE DATABASE AuthDB;
-- +goose StatementEnd

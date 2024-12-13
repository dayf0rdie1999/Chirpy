-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN is_chirpy_red bool DEFAULT FALSE NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN is_chirpy_red;
-- +goose StatementEnd

-- +goose Up
ALTER TABLE transactions ADD COLUMN fee INTEGER DEFAULT (0);

-- +goose Down
ALTER TABLE transactions DROP COLUMN fee;

-- +goose Up
-- +goose StatementBegin
ALTER TABLE
    "transaction"
ADD
    spender_id int NULL;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE
    "transaction" DROP COLUMN spender_id;

-- +goose StatementEnd

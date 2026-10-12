-- +goose Up
-- +goose StatementBegin
-- 已有食物和未指定收藏状态的新食物都默认为未收藏。
ALTER TABLE foods
    ADD COLUMN favorite BOOLEAN NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE foods
    DROP COLUMN favorite;
-- +goose StatementEnd

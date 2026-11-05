-- +goose Up
CREATE TABLE chat_requests (
    id BIGSERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    reply TEXT NOT NULL DEFAULT '',
    tool_results JSONB NOT NULL DEFAULT '[]',
    status TEXT NOT NULL,
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE chat_requests;

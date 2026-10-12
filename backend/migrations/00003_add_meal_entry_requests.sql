-- +goose Up
CREATE TABLE meal_entry_requests (
    key TEXT PRIMARY KEY CHECK (key ~ '^[A-Za-z0-9._:-]{1,128}$'),
    request_hash TEXT NOT NULL CHECK (request_hash ~ '^[0-9a-f]{64}$'),
    response JSONB NOT NULL CHECK (jsonb_typeof(response) = 'object'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE meal_entry_requests;

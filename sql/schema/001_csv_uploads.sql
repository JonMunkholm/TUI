-- +goose Up
CREATE TABLE csv_uploads (
    name TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('upload', 'update', 'delete')),
    uploaded_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT csv_uploads_name_action_unique UNIQUE (name, action)
);

-- +goose Down
DROP TABLE IF EXISTS csv_uploads;

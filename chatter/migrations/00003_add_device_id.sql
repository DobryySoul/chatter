-- +goose Up
ALTER TABLE refresh_tokens ADD COLUMN device_id VARCHAR(255) NOT NULL DEFAULT 'unknown';
CREATE INDEX idx_refresh_tokens_device_id ON refresh_tokens (device_id);

-- +goose Down
ALTER TABLE refresh_tokens DROP COLUMN device_id;


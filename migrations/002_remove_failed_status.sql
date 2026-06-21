-- +goose Up
ALTER TABLE jobs
DROP CONSTRAINT jobs_status_check;

ALTER TABLE jobs
ADD CONSTRAINT jobs_status_check CHECK (
    status IN ('pending', 'processing', 'succeeded', 'cancelled', 'dead')
);

-- +goose Down
ALTER TABLE jobs
DROP CONSTRAINT jobs_status_check;

ALTER TABLE jobs
ADD CONSTRAINT jobs_status_check CHECK (
    status IN ('pending', 'processing', 'succeeded', 'failed', 'cancelled', 'dead')
);
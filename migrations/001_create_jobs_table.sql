-- +goose Up
CREATE TABLE jobs (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL,
    attempt INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    run_at TIMESTAMPTZ NOT NULL,
    locked_at TIMESTAMPTZ,
    locked_by TEXT,
    last_error TEXT,
    idempotency_key TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT jobs_status_check CHECK (
        status IN ('pending', 'processing', 'succeeded', 'failed', 'cancelled', 'dead')
    ),
    CONSTRAINT jobs_attempt_check CHECK (attempt >= 0),
    CONSTRAINT jobs_max_attempts_check CHECK (max_attempts > 0)
);

CREATE INDEX idx_jobs_pending_run_at_created_at
ON jobs (run_at, created_at)
WHERE status = 'pending';

CREATE INDEX idx_jobs_status
ON jobs (status);

-- +goose Down
DROP TABLE IF EXISTS jobs;
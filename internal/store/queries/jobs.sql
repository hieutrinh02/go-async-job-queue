-- name: CreateJob :one
INSERT INTO jobs (
    id,
    type,
    payload,
    status,
    max_attempts,
    run_at,
    idempotency_key
)
VALUES (
    $1,
    $2,
    $3,
    'pending',
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetJob :one
SELECT *
FROM jobs
WHERE id = $1;

-- name: GetJobByIdempotencyKey :one
SELECT *
FROM jobs
WHERE idempotency_key = $1;

-- name: CancelPendingJob :one
UPDATE jobs
SET
    status = 'cancelled',
    updated_at = NOW()
WHERE id = $1
    AND status = 'pending'
RETURNING *;

-- name: ClaimJobs :many
WITH selected_jobs AS (
    SELECT id
    FROM jobs
    WHERE status = 'pending'
        AND run_at <= NOW()
    ORDER BY created_at
    FOR UPDATE SKIP LOCKED
    LIMIT $1
)
UPDATE jobs
SET
    status = 'processing',
    locked_at = NOW(),
    locked_by = $2,
    updated_at = NOW()
FROM selected_jobs
WHERE jobs.id = selected_jobs.id
RETURNING jobs.*;

-- name: MarkJobSucceeded :one
UPDATE jobs
SET
    status = 'succeeded',
    locked_at = NULL,
    locked_by = NULL,
    last_error = NULL,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkJobFailed :one
UPDATE jobs
SET
    attempt = attempt + 1,
    status = CASE
        WHEN attempt + 1 >= max_attempts THEN 'dead'
        ELSE 'pending'
    END,
    run_at = CASE
        WHEN attempt + 1 >= max_attempts THEN run_at
        ELSE $2
    END,
    locked_at = NULL,
    locked_by = NULL,
    last_error = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ReleaseProcessingJobsByWorker :exec
UPDATE jobs
SET
    status = 'pending',
    locked_at = NULL,
    locked_by = NULL,
    updated_at = NOW()
WHERE status = 'processing'
  AND locked_by = $1;
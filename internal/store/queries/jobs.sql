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
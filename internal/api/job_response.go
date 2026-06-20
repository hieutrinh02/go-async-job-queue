package api

import (
	"encoding/json"
	"time"

	"github.com/hieutrinh02/go-async-job-queue/internal/store/sqlc"
)

type jobResponse struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	Status         string          `json:"status"`
	Attempt        int32           `json:"attempt"`
	MaxAttempts    int32           `json:"max_attempts"`
	RunAt          time.Time       `json:"run_at"`
	LockedAt       *time.Time      `json:"locked_at,omitempty"`
	LockedBy       *string         `json:"locked_by,omitempty"`
	LastError      *string         `json:"last_error,omitempty"`
	IdempotencyKey string          `json:"idempotency_key"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func newJobResponse(job sqlc.Job) jobResponse {
	var lockedAt *time.Time
	if job.LockedAt.Valid {
		lockedAt = &job.LockedAt.Time
	}

	var lockedBy *string
	if job.LockedBy.Valid {
		lockedBy = &job.LockedBy.String
	}

	var lastError *string
	if job.LastError.Valid {
		lastError = &job.LastError.String
	}

	idempotencyKey := ""
	if job.IdempotencyKey.Valid {
		idempotencyKey = job.IdempotencyKey.String
	}

	return jobResponse{
		ID:             job.ID.String(),
		Type:           job.Type,
		Payload:        json.RawMessage(job.Payload),
		Status:         job.Status,
		Attempt:        job.Attempt,
		MaxAttempts:    job.MaxAttempts,
		RunAt:          job.RunAt.Time,
		LockedAt:       lockedAt,
		LockedBy:       lockedBy,
		LastError:      lastError,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      job.CreatedAt.Time,
		UpdatedAt:      job.UpdatedAt.Time,
	}
}

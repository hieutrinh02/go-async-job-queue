package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hieutrinh02/go-async-job-queue/internal/store/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	queries *sqlc.Queries
}

type CreateJobParams struct {
	ID             string
	Type           string
	Payload        json.RawMessage
	MaxAttempts    int32
	RunAt          time.Time
	IdempotencyKey string
}

type MarkJobFailedParams struct {
	ID        string
	NextRunAt time.Time
	LastError string
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{
		queries: sqlc.New(pool),
	}
}

func (s *Store) CreateJob(ctx context.Context, params CreateJobParams) (sqlc.Job, error) {
	id := pgtype.UUID{}
	if err := id.Scan(params.ID); err != nil {
		return sqlc.Job{}, err
	}

	runAt := pgtype.Timestamptz{
		Time:  params.RunAt,
		Valid: true,
	}

	idempotencyKey := pgtype.Text{
		String: params.IdempotencyKey,
		Valid:  params.IdempotencyKey != "",
	}

	return s.queries.CreateJob(ctx, sqlc.CreateJobParams{
		ID:             id,
		Type:           params.Type,
		Payload:        params.Payload,
		MaxAttempts:    params.MaxAttempts,
		RunAt:          runAt,
		IdempotencyKey: idempotencyKey,
	})
}

func (s *Store) GetJob(ctx context.Context, id string) (sqlc.Job, error) {
	jobID := pgtype.UUID{}
	if err := jobID.Scan(id); err != nil {
		return sqlc.Job{}, err
	}

	return s.queries.GetJob(ctx, jobID)
}

func (s *Store) GetJobByIdempotencyKey(ctx context.Context, idempotencyKey string) (sqlc.Job, error) {
	key := pgtype.Text{
		String: idempotencyKey,
		Valid:  idempotencyKey != "",
	}

	return s.queries.GetJobByIdempotencyKey(ctx, key)
}

func (s *Store) CancelPendingJob(ctx context.Context, id string) (sqlc.Job, error) {
	jobID := pgtype.UUID{}
	if err := jobID.Scan(id); err != nil {
		return sqlc.Job{}, err
	}

	return s.queries.CancelPendingJob(ctx, jobID)
}

func (s *Store) ClaimJobs(ctx context.Context, limit int32, workerID string) ([]sqlc.Job, error) {
	return s.queries.ClaimJobs(ctx, sqlc.ClaimJobsParams{
		Limit:    limit,
		LockedBy: pgtype.Text{String: workerID, Valid: workerID != ""},
	})
}

func (s *Store) MarkJobSucceeded(ctx context.Context, id string) (sqlc.Job, error) {
	jobID := pgtype.UUID{}
	if err := jobID.Scan(id); err != nil {
		return sqlc.Job{}, err
	}

	return s.queries.MarkJobSucceeded(ctx, jobID)
}

func (s *Store) MarkJobFailed(ctx context.Context, params MarkJobFailedParams) (sqlc.Job, error) {
	jobID := pgtype.UUID{}
	if err := jobID.Scan(params.ID); err != nil {
		return sqlc.Job{}, err
	}

	nextRunAt := pgtype.Timestamptz{
		Time:  params.NextRunAt,
		Valid: true,
	}

	lastError := pgtype.Text{
		String: params.LastError,
		Valid:  params.LastError != "",
	}

	return s.queries.MarkJobFailed(ctx, sqlc.MarkJobFailedParams{
		ID:        jobID,
		RunAt:     nextRunAt,
		LastError: lastError,
	})
}

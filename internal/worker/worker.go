package worker

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/hieutrinh02/go-async-job-queue/internal/store"
)

type Worker struct {
	id           string
	store        *store.Store
	batchSize    int32
	pollInterval time.Duration
}

type Config struct {
	ID           string
	BatchSize    int32
	PollInterval time.Duration
}

func New(jobStore *store.Store, cfg Config) *Worker {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}

	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 2 * time.Second
	}

	return &Worker{
		id:           cfg.ID,
		store:        jobStore,
		batchSize:    cfg.BatchSize,
		pollInterval: cfg.PollInterval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	log.Printf("worker %s started", w.id)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		// Wait for either shutdown or the next polling tick.
		select {
		case <-ctx.Done():
			log.Printf("worker %s stopped", w.id)
			return
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

func (w *Worker) poll(ctx context.Context) {
	jobs, err := w.store.ClaimJobs(ctx, w.batchSize, w.id)
	if err != nil {
		log.Printf("worker %s failed to claim jobs: %v", w.id, err)
		return
	}

	if len(jobs) == 0 {
		return
	}

	log.Printf("worker %s claimed %d job(s)", w.id, len(jobs))

	for _, job := range jobs {
		log.Printf("worker %s claimed job %s type=%s", w.id, job.ID.String(), job.Type)
		w.processJob(ctx, job.ID.String(), job.Type, job.Attempt)
	}
}

func (w *Worker) processJob(ctx context.Context, jobID string, jobType string, attempt int32) {
	log.Printf("worker %s processing job %s type=%s", w.id, jobID, jobType)

	if err := executeMockJob(ctx, jobType); err != nil {
		nextRunAt := time.Now().UTC().Add(retryBackoff(attempt))

		job, markErr := w.store.MarkJobFailed(ctx, store.MarkJobFailedParams{
			ID:        jobID,
			NextRunAt: nextRunAt,
			LastError: err.Error(),
		})
		if markErr != nil {
			log.Printf("worker %s failed to mark job %s failed: %v", w.id, jobID, markErr)
			return
		}

		log.Printf("worker %s failed job %s status=%s attempt=%d error=%q", w.id, jobID, job.Status, job.Attempt, err.Error())
		return
	}

	if _, err := w.store.MarkJobSucceeded(ctx, jobID); err != nil {
		log.Printf("worker %s failed to mark job %s succeeded: %v", w.id, jobID, err)
		return
	}

	log.Printf("worker %s completed job %s", w.id, jobID)
}

func retryBackoff(attempt int32) time.Duration {
	switch attempt {
	case 0:
		return 5 * time.Second
	case 1:
		return 30 * time.Second
	case 2:
		return 2 * time.Minute
	default:
		return 10 * time.Minute
	}
}

func executeMockJob(ctx context.Context, jobType string) error {
	switch jobType {
	case "send_email":
		if err := sleepWithContext(ctx, 300*time.Millisecond); err != nil {
			return err
		}
	case "generate_report":
		if err := sleepWithContext(ctx, 700*time.Millisecond); err != nil {
			return err
		}
	case "webhook_delivery":
		if err := sleepWithContext(ctx, 500*time.Millisecond); err != nil {
			return err
		}
	case "image_resize":
		if err := sleepWithContext(ctx, 800*time.Millisecond); err != nil {
			return err
		}
	case "always_fail":
		if err := sleepWithContext(ctx, 300*time.Millisecond); err != nil {
			return err
		}
		return errors.New("mock handler failed")
	default:
		if err := sleepWithContext(ctx, 200*time.Millisecond); err != nil {
			return err
		}
		return errors.New("unknown job type")
	}

	if rand.Intn(100) < 30 {
		return errors.New("mock transient failure")
	}

	return nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

package worker

import (
	"context"
	"log"
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
		// Wait for many channels at the same time
		// We have 2 channels
		// ctx.Done()
		// ticker.C
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
		w.processJob(ctx, job.ID.String(), job.Type)
	}
}

func (w *Worker) processJob(ctx context.Context, jobID string, jobType string) {
	log.Printf("worker %s processing job %s type=%s", w.id, jobID, jobType)

	time.Sleep(500 * time.Millisecond)

	if _, err := w.store.MarkJobSucceeded(ctx, jobID); err != nil {
		log.Printf("worker %s failed to mark job %s succeeded: %v", w.id, jobID, err)
		return
	}

	log.Printf("worker %s completed job %s", w.id, jobID)
}

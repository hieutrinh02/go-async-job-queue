package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hieutrinh02/go-async-job-queue/internal/api"
	"github.com/hieutrinh02/go-async-job-queue/internal/config"
	"github.com/hieutrinh02/go-async-job-queue/internal/db"
	"github.com/hieutrinh02/go-async-job-queue/internal/store"
	"github.com/hieutrinh02/go-async-job-queue/internal/worker"
)

func main() {
	// Create config
	cfg := config.Load()

	// Database pool
	ctx := context.Background()
	dbPool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()
	log.Println("connected to database")

	// Create store
	jobStore := store.New(dbPool)

	// Create and start worker
	workerCtx, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()

	var workerWG sync.WaitGroup

	workerIDs := make([]string, 0, cfg.WorkerCount)

	for i := 1; i <= cfg.WorkerCount; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		workerIDs = append(workerIDs, workerID)

		jobWorker := worker.New(jobStore, worker.Config{
			ID:           workerID,
			BatchSize:    cfg.WorkerBatchSize,
			PollInterval: cfg.WorkerPollInterval,
		})

		workerWG.Add(1)
		go func(jobWorker *worker.Worker) {
			defer workerWG.Done()
			jobWorker.Run(workerCtx)
		}(jobWorker)
	}

	log.Printf("started %d worker(s)", cfg.WorkerCount)

	// Create router and address
	router := api.NewRouter(jobStore)
	addr := ":" + cfg.Port

	// Create HTTP server
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Listen and serve in Goroutine
	go func() {
		log.Println("server listening on " + addr)

		// If the server's errored, and the error is not a normal "server closed"
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Create a channel to receive signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal from channel
	<-quit
	log.Println("shutting down server...")

	// Stop accepting new HTTP requests
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	} else {
		log.Println("server stopped")
	}

	// Stop workers and wait for them to exit
	stopWorker()

	workersDone := make(chan struct{})
	go func() {
		workerWG.Wait()
		close(workersDone)
	}()

	select {
	case <-workersDone:
		log.Println("workers stopped")
	case <-time.After(cfg.WorkerShutdownTimeout):
		log.Printf("worker shutdown timeout exceeded after %s", cfg.WorkerShutdownTimeout)

		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, workerID := range workerIDs {
			if err := jobStore.ReleaseProcessingJobsByWorker(releaseCtx, workerID); err != nil {
				log.Printf("failed to release processing jobs for %s: %v", workerID, err)
			}
		}
	}
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
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

	jobWorker := worker.New(jobStore, worker.Config{
		ID:           "worker-1",
		BatchSize:    10,
		PollInterval: 2 * time.Second,
	})

	go jobWorker.Run(workerCtx)

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

	// Stop worker
	stopWorker()

	// Create context timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // should be called just before main() exits

	// Shutdown server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	log.Println("server stopped")
}

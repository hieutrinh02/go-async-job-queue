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
)

func main() {
	// Create config, router and address
	cfg := config.Load()
	router := api.NewRouter()
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

	<-quit
	log.Println("shutting down server...")

	// Create context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // should be called just before main() exits

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("server stopped")
}

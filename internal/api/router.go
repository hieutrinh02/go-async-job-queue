package api

import (
	"net/http"

	"github.com/hieutrinh02/go-async-job-queue/internal/store"
)

type Server struct {
	store *store.Store
}

func NewRouter(jobStore *store.Store) http.Handler {
	server := &Server{
		store: jobStore,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", server.handleHealthz)
	mux.HandleFunc("POST /v1/jobs", server.handleCreateJob)
	mux.HandleFunc("GET /v1/jobs/{id}", server.handleGetJob)
	mux.HandleFunc("POST /v1/jobs/{id}/cancel", server.handleCancelJob)

	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

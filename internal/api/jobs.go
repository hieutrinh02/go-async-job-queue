package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hieutrinh02/go-async-job-queue/internal/store"
)

type createJobRequest struct {
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	RunAt          *time.Time      `json:"run_at"`
	IdempotencyKey string          `json:"idempotency_key"`
}

func (s *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Type == "" {
		writeJSONError(w, http.StatusBadRequest, "type is required")
		return
	}

	if req.IdempotencyKey == "" {
		writeJSONError(w, http.StatusBadRequest, "idempotency_key is required")
		return
	}

	if len(req.Payload) == 0 {
		req.Payload = json.RawMessage(`{}`)
	}

	runAt := time.Now().UTC()
	if req.RunAt != nil {
		runAt = req.RunAt.UTC()
	}

	jobID := uuid.NewString()

	job, err := s.store.CreateJob(r.Context(), store.CreateJobParams{
		ID:             jobID,
		Type:           req.Type,
		Payload:        req.Payload,
		MaxAttempts:    5,
		RunAt:          runAt,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create job")
		return
	}

	writeJSON(w, http.StatusCreated, newJobResponse(job))
}

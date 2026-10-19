package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/apiwat-tb/go-webhook/internal/domain"
)

// errorResponse standardizes the JSON structure for HTTP error payloads.
type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON sets the Content-Type header, HTTP status code, and serializes the value to JSON.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError maps known application/domain errors to proper HTTP status codes.
func writeError(w http.ResponseWriter, err error) {
	var status int
	var msg string

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
		msg = err.Error()
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		msg = err.Error()
	default:
		status = http.StatusInternalServerError
		msg = "internal server error"
	}

	writeJSON(w, status, errorResponse{Error: msg})
}

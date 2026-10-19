package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/apiwat-tb/go-webhook/internal/domain"
	"github.com/apiwat-tb/go-webhook/internal/usecase"
)

// EventHandler handles event triggering (/events) and receiving incoming webhook deliveries (/incoming).
type EventHandler struct {
	dispatchUsecase *usecase.DispatchUsecase
	verifySig       func(secret string, payload []byte, sig string) bool
	incomingSecret  string
}

// NewEventHandler creates an EventHandler with injected signature verification logic and secret.
func NewEventHandler(
	dispatchUsecase *usecase.DispatchUsecase,
	verifySig func(secret string, payload []byte, sig string) bool,
	incomingSecret string,
) *EventHandler {
	return &EventHandler{
		dispatchUsecase: dispatchUsecase,
		verifySig:       verifySig,
		incomingSecret:  incomingSecret,
	}
}

type dispatchRequest struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// Dispatch handles POST /events to trigger webhook broadcasts.
func (h *EventHandler) Dispatch(w http.ResponseWriter, r *http.Request) {
	var req dispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, fmt.Errorf("%w: malformed JSON body", domain.ErrInvalidInput))
		return
	}

	result, err := h.dispatchUsecase.Dispatch(req.Event, req.Data)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Incoming handles POST /incoming to verify and ingest delivered webhooks.
func (h *EventHandler) Incoming(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("X-Signature")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, fmt.Errorf("%w: failed to read request body", domain.ErrInvalidInput))
		return
	}

	if sig == "" || !h.verifySig(h.incomingSecret, body, sig) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid signature"})
		return
	}

	log.Printf("Received incoming webhook payload: %s", string(body))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

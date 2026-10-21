package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/apiwat-tb/go-webhook/internal/domain"
	"github.com/apiwat-tb/go-webhook/internal/usecase"
	"github.com/gofiber/fiber/v2"
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
func (h *EventHandler) Dispatch(c *fiber.Ctx) error {
	var req dispatchRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return writeError(c, fmt.Errorf("%w: malformed JSON body", domain.ErrInvalidInput))
	}

	result, err := h.dispatchUsecase.Dispatch(req.Event, req.Data)
	if err != nil {
		return writeError(c, err)
	}

	return writeJSON(c, http.StatusOK, result)
}

// Incoming handles POST /incoming to verify and ingest delivered webhooks.
func (h *EventHandler) Incoming(c *fiber.Ctx) error {
	sig := c.Get("X-Signature")
	body := c.Body()

	if sig == "" || !h.verifySig(h.incomingSecret, body, sig) {
		return writeJSON(c, http.StatusUnauthorized, errorResponse{Error: "invalid signature"})
	}

	log.Printf("Received incoming webhook payload: %s", string(body))
	return writeJSON(c, http.StatusOK, map[string]string{"status": "ok"})
}

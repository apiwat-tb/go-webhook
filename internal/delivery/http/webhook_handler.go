package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/apiwat-tb/go-webhook/internal/domain"
	"github.com/apiwat-tb/go-webhook/internal/usecase"
)

// WebhookHandler handles HTTP endpoints for webhook CRUD and delivery logs.
type WebhookHandler struct {
	webhookUsecase  *usecase.WebhookUsecase
	dispatchUsecase *usecase.DispatchUsecase
}

// NewWebhookHandler constructs a new WebhookHandler.
func NewWebhookHandler(webhookUsecase *usecase.WebhookUsecase, dispatchUsecase *usecase.DispatchUsecase) *WebhookHandler {
	return &WebhookHandler{
		webhookUsecase:  webhookUsecase,
		dispatchUsecase: dispatchUsecase,
	}
}

type createWebhookRequest struct {
	URL    string `json:"url"`
	Event  string `json:"event"`
	Secret string `json:"secret"`
}

// Create handles POST /webhooks.
func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, fmt.Errorf("%w: malformed JSON body", domain.ErrInvalidInput))
		return
	}

	created, err := h.webhookUsecase.Create(req.URL, req.Event, req.Secret)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// List handles GET /webhooks.
func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	webhooks, err := h.webhookUsecase.List()
	if err != nil {
		writeError(w, err)
		return
	}

	// Always return an empty JSON array instead of null when no webhooks exist.
	if webhooks == nil {
		webhooks = []*domain.Webhook{}
	}

	writeJSON(w, http.StatusOK, webhooks)
}

// Get handles GET /webhooks/{id}.
func (h *WebhookHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	webhook, err := h.webhookUsecase.Get(id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, webhook)
}

type updateWebhookRequest struct {
	URL    *string `json:"url"`
	Event  *string `json:"event"`
	Active *bool   `json:"active"`
}

// Update handles PUT /webhooks/{id}.
func (h *WebhookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, fmt.Errorf("%w: malformed JSON body", domain.ErrInvalidInput))
		return
	}

	updated, err := h.webhookUsecase.Update(id, req.URL, req.Event, req.Active)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /webhooks/{id}.
func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.webhookUsecase.Delete(id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListDeliveries handles GET /webhooks/{id}/deliveries.
func (h *WebhookHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	deliveries, err := h.dispatchUsecase.ListDeliveries(id)
	if err != nil {
		writeError(w, err)
		return
	}

	// Ensure empty array serialization instead of null.
	if deliveries == nil {
		deliveries = []*domain.Delivery{}
	}

	writeJSON(w, http.StatusOK, deliveries)
}

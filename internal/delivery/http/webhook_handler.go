package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/apiwat-tb/go-webhook/internal/domain"
	"github.com/apiwat-tb/go-webhook/internal/usecase"
	"github.com/gofiber/fiber/v2"
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
func (h *WebhookHandler) Create(c *fiber.Ctx) error {
	var req createWebhookRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return writeError(c, fmt.Errorf("%w: malformed JSON body", domain.ErrInvalidInput))
	}

	created, err := h.webhookUsecase.Create(req.URL, req.Event, req.Secret)
	if err != nil {
		return writeError(c, err)
	}

	return writeJSON(c, http.StatusCreated, created)
}

// List handles GET /webhooks.
func (h *WebhookHandler) List(c *fiber.Ctx) error {
	webhooks, err := h.webhookUsecase.List()
	if err != nil {
		return writeError(c, err)
	}

	// Always return an empty JSON array instead of null when no webhooks exist.
	if webhooks == nil {
		webhooks = []*domain.Webhook{}
	}

	return writeJSON(c, http.StatusOK, webhooks)
}

// Get handles GET /webhooks/:id.
func (h *WebhookHandler) Get(c *fiber.Ctx) error {
	webhook, err := h.webhookUsecase.Get(c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}

	return writeJSON(c, http.StatusOK, webhook)
}

type updateWebhookRequest struct {
	URL    *string `json:"url"`
	Event  *string `json:"event"`
	Active *bool   `json:"active"`
}

// Update handles PUT /webhooks/:id.
func (h *WebhookHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req updateWebhookRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return writeError(c, fmt.Errorf("%w: malformed JSON body", domain.ErrInvalidInput))
	}

	updated, err := h.webhookUsecase.Update(id, req.URL, req.Event, req.Active)
	if err != nil {
		return writeError(c, err)
	}

	return writeJSON(c, http.StatusOK, updated)
}

// Delete handles DELETE /webhooks/:id.
func (h *WebhookHandler) Delete(c *fiber.Ctx) error {
	if err := h.webhookUsecase.Delete(c.Params("id")); err != nil {
		return writeError(c, err)
	}

	return c.SendStatus(http.StatusNoContent)
}

// ListDeliveries handles GET /webhooks/:id/deliveries.
func (h *WebhookHandler) ListDeliveries(c *fiber.Ctx) error {
	deliveries, err := h.dispatchUsecase.ListDeliveries(c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}

	// Ensure empty array serialization instead of null.
	if deliveries == nil {
		deliveries = []*domain.Delivery{}
	}

	return writeJSON(c, http.StatusOK, deliveries)
}

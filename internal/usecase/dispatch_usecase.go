package usecase

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/apiwat-tb/go-webhook/internal/domain"
	"github.com/google/uuid"
)

// DispatchResult summarizes the outcome of dispatching an event to subscribers.
type DispatchResult struct {
	Matched   int `json:"matched"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

// eventPayload structures the outgoing JSON body delivered to subscriber endpoints.
type eventPayload struct {
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
	Timestamp string          `json:"timestamp"`
}

// DispatchUsecase coordinates event publishing, webhook dispatching, and delivery logs.
type DispatchUsecase struct {
	webhooks   WebhookRepository
	deliveries DeliveryRepository
	sender     Sender
}

// NewDispatchUsecase creates a new DispatchUsecase instance.
func NewDispatchUsecase(webhooks WebhookRepository, deliveries DeliveryRepository, sender Sender) *DispatchUsecase {
	return &DispatchUsecase{
		webhooks:   webhooks,
		deliveries: deliveries,
		sender:     sender,
	}
}

// Dispatch triggers delivery of an event to all active matching webhook subscriptions sequentially.
func (u *DispatchUsecase) Dispatch(event string, data json.RawMessage) (DispatchResult, error) {
	trimmedEvent := strings.TrimSpace(event)
	if trimmedEvent == "" {
		return DispatchResult{}, fmt.Errorf("%w: event is required", domain.ErrInvalidInput)
	}

	if len(data) == 0 || !json.Valid(data) {
		return DispatchResult{}, fmt.Errorf("%w: data must be valid JSON", domain.ErrInvalidInput)
	}

	payload := eventPayload{
		Event:     trimmedEvent,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return DispatchResult{}, fmt.Errorf("failed to marshal dispatch payload: %w", err)
	}

	matchedWebhooks, err := u.webhooks.ListByEvent(trimmedEvent)
	if err != nil {
		return DispatchResult{}, err
	}

	result := DispatchResult{
		Matched: len(matchedWebhooks),
	}

	// Dispatch to subscribers sequentially to keep logic straightforward and predictable.
	for _, w := range matchedWebhooks {
		statusCode, sendErr := u.sender.Send(w.URL, w.Secret, payloadBytes)
		success := statusCode >= 200 && statusCode < 300

		var errText string
		if sendErr != nil {
			errText = sendErr.Error()
		}

		delivery := &domain.Delivery{
			ID:         uuid.NewString(),
			WebhookID:  w.ID,
			Event:      trimmedEvent,
			Payload:    string(payloadBytes),
			StatusCode: statusCode,
			Success:    success,
			Error:      errText,
			CreatedAt:  time.Now().UTC(),
		}

		// A delivery or network failure must not stop the loop or fail the entire dispatch.
		_ = u.deliveries.Save(delivery)

		if success {
			result.Succeeded++
		} else {
			result.Failed++
		}
	}

	return result, nil
}

// ListDeliveries retrieves the delivery history for a webhook, returning domain.ErrNotFound if absent.
func (u *DispatchUsecase) ListDeliveries(webhookID string) ([]*domain.Delivery, error) {
	if _, err := u.webhooks.GetByID(webhookID); err != nil {
		return nil, err
	}

	return u.deliveries.ListByWebhook(webhookID)
}

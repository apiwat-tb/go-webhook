package usecase

import (
	"github.com/apiwat-tb/go-webhook/internal/domain"
)

// WebhookRepository defines storage operations for Webhook entities.
// Keeping this interface in the usecase layer decouples business logic from persistence.
type WebhookRepository interface {
	Create(w *domain.Webhook) error
	GetByID(id string) (*domain.Webhook, error)          // Returns domain.ErrNotFound if absent
	List() ([]*domain.Webhook, error)
	ListByEvent(event string) ([]*domain.Webhook, error) // Returns active webhooks matching the event
	Update(w *domain.Webhook) error                      // Returns domain.ErrNotFound if absent
	Delete(id string) error                              // Returns domain.ErrNotFound if absent
}

// DeliveryRepository defines storage operations for Delivery history.
type DeliveryRepository interface {
	Save(d *domain.Delivery) error
	ListByWebhook(webhookID string) ([]*domain.Delivery, error)
}

// Sender defines the contract for sending HTTP webhook requests to subscribers.
type Sender interface {
	// Send POSTs payload to url, signed with secret. Returns HTTP status code.
	Send(url, secret string, payload []byte) (statusCode int, err error)
}

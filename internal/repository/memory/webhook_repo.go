package memory

import (
	"sort"
	"sync"

	"github.com/apiwat-tb/go-webhook/internal/domain"
)

// WebhookRepo provides an in-memory, thread-safe implementation of usecase.WebhookRepository.
type WebhookRepo struct {
	mu       sync.RWMutex
	webhooks map[string]*domain.Webhook
}

// NewWebhookRepo initializes an empty in-memory webhook repository.
func NewWebhookRepo() *WebhookRepo {
	return &WebhookRepo{
		webhooks: make(map[string]*domain.Webhook),
	}
}

// Create stores a defensive copy of the webhook.
func (r *WebhookRepo) Create(w *domain.Webhook) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.webhooks[w.ID] = cloneWebhook(w)
	return nil
}

// GetByID returns a copy of the webhook or domain.ErrNotFound if absent.
func (r *WebhookRepo) GetByID(id string) (*domain.Webhook, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	w, exists := r.webhooks[id]
	if !exists {
		return nil, domain.ErrNotFound
	}

	return cloneWebhook(w), nil
}

// List returns all webhooks sorted by CreatedAt ascending.
func (r *WebhookRepo) List() ([]*domain.Webhook, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Webhook, 0, len(r.webhooks))
	for _, w := range r.webhooks {
		result = append(result, cloneWebhook(w))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

// ListByEvent returns only active webhooks matching the event, sorted by CreatedAt ascending.
func (r *WebhookRepo) ListByEvent(event string) ([]*domain.Webhook, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Webhook, 0)
	for _, w := range r.webhooks {
		if w.Active && w.Event == event {
			result = append(result, cloneWebhook(w))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

// Update updates an existing webhook or returns domain.ErrNotFound.
func (r *WebhookRepo) Update(w *domain.Webhook) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.webhooks[w.ID]; !exists {
		return domain.ErrNotFound
	}

	r.webhooks[w.ID] = cloneWebhook(w)
	return nil
}

// Delete deletes a webhook by ID or returns domain.ErrNotFound.
func (r *WebhookRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.webhooks[id]; !exists {
		return domain.ErrNotFound
	}

	delete(r.webhooks, id)
	return nil
}

// cloneWebhook returns a shallow copy of Webhook to prevent external mutations of stored state.
func cloneWebhook(w *domain.Webhook) *domain.Webhook {
	if w == nil {
		return nil
	}
	copied := *w
	return &copied
}

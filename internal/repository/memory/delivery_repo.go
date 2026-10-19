package memory

import (
	"sort"
	"sync"

	"github.com/apiwat-tb/go-webhook/internal/domain"
)

// DeliveryRepo provides an in-memory, thread-safe implementation of usecase.DeliveryRepository.
type DeliveryRepo struct {
	mu         sync.RWMutex
	deliveries []*domain.Delivery
}

// NewDeliveryRepo initializes an empty in-memory delivery repository.
func NewDeliveryRepo() *DeliveryRepo {
	return &DeliveryRepo{
		deliveries: make([]*domain.Delivery, 0),
	}
}

// Save stores a defensive copy of the delivery record.
func (r *DeliveryRepo) Save(d *domain.Delivery) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.deliveries = append(r.deliveries, cloneDelivery(d))
	return nil
}

// ListByWebhook returns deliveries for a given webhook ID, ordered newest first.
func (r *DeliveryRepo) ListByWebhook(webhookID string) ([]*domain.Delivery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Delivery, 0)
	for _, d := range r.deliveries {
		if d.WebhookID == webhookID {
			result = append(result, cloneDelivery(d))
		}
	}

	// Sort newest first (descending by CreatedAt)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}

// cloneDelivery returns a shallow copy of Delivery to avoid shared mutable state.
func cloneDelivery(d *domain.Delivery) *domain.Delivery {
	if d == nil {
		return nil
	}
	copied := *d
	return &copied
}

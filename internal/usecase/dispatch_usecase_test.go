package usecase_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/apiwat-tb/go-webhook/internal/domain"
	"github.com/apiwat-tb/go-webhook/internal/usecase"
)

type fakeDeliveryRepo struct {
	deliveries []*domain.Delivery
}

func newFakeDeliveryRepo() *fakeDeliveryRepo {
	return &fakeDeliveryRepo{deliveries: make([]*domain.Delivery, 0)}
}

func (f *fakeDeliveryRepo) Save(d *domain.Delivery) error {
	f.deliveries = append(f.deliveries, d)
	return nil
}

func (f *fakeDeliveryRepo) ListByWebhook(webhookID string) ([]*domain.Delivery, error) {
	var list []*domain.Delivery
	for _, d := range f.deliveries {
		if d.WebhookID == webhookID {
			list = append(list, d)
		}
	}
	return list, nil
}

// fakeSender simulates successful and failing HTTP dispatches based on destination URL.
type fakeSender struct{}

func (s *fakeSender) Send(targetURL, secret string, payload []byte) (int, error) {
	if targetURL == "http://fail.example.com" {
		return 500, fmt.Errorf("connection refused")
	}
	return 200, nil
}

func TestDispatchUsecase_Dispatch(t *testing.T) {
	webhookRepo := newFakeWebhookRepo()
	deliveryRepo := newFakeDeliveryRepo()
	sender := &fakeSender{}

	uc := usecase.NewDispatchUsecase(webhookRepo, deliveryRepo, sender)

	// Register two webhooks for "order.paid": one succeeding, one failing
	_ = webhookRepo.Create(&domain.Webhook{
		ID:        "wh-1",
		URL:       "http://ok.example.com",
		Event:     "order.paid",
		Secret:    "secret1",
		Active:    true,
		CreatedAt: time.Now(),
	})
	_ = webhookRepo.Create(&domain.Webhook{
		ID:        "wh-2",
		URL:       "http://fail.example.com",
		Event:     "order.paid",
		Secret:    "secret2",
		Active:    true,
		CreatedAt: time.Now(),
	})
	// Inactive webhook should NOT be matched
	_ = webhookRepo.Create(&domain.Webhook{
		ID:        "wh-3",
		URL:       "http://ok.example.com",
		Event:     "order.paid",
		Secret:    "secret3",
		Active:    false,
		CreatedAt: time.Now(),
	})

	result, err := uc.Dispatch("order.paid", json.RawMessage(`{"order_id": 42}`))
	if err != nil {
		t.Fatalf("expected dispatch to succeed, got: %v", err)
	}

	if result.Matched != 2 {
		t.Errorf("expected matched=2, got: %d", result.Matched)
	}
	if result.Succeeded != 1 {
		t.Errorf("expected succeeded=1, got: %d", result.Succeeded)
	}
	if result.Failed != 1 {
		t.Errorf("expected failed=1, got: %d", result.Failed)
	}

	// Verify deliveries were recorded in the repository
	if len(deliveryRepo.deliveries) != 2 {
		t.Fatalf("expected 2 saved deliveries, got: %d", len(deliveryRepo.deliveries))
	}
}

func TestDispatchUsecase_Validation(t *testing.T) {
	webhookRepo := newFakeWebhookRepo()
	deliveryRepo := newFakeDeliveryRepo()
	uc := usecase.NewDispatchUsecase(webhookRepo, deliveryRepo, &fakeSender{})

	// Empty event
	_, err := uc.Dispatch("  ", json.RawMessage(`{}`))
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty event, got: %v", err)
	}

	// Invalid JSON
	_, err = uc.Dispatch("order.paid", json.RawMessage(`invalid-json`))
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for malformed JSON, got: %v", err)
	}
}

func TestDispatchUsecase_ListDeliveries(t *testing.T) {
	webhookRepo := newFakeWebhookRepo()
	deliveryRepo := newFakeDeliveryRepo()
	uc := usecase.NewDispatchUsecase(webhookRepo, deliveryRepo, &fakeSender{})

	// Non-existent webhook returns ErrNotFound
	_, err := uc.ListDeliveries("unknown-id")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound for missing webhook, got: %v", err)
	}

	// Existing webhook returns empty or populated list
	_ = webhookRepo.Create(&domain.Webhook{
		ID:     "wh-1",
		URL:    "http://example.com",
		Active: true,
	})

	list, err := uc.ListDeliveries("wh-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 deliveries, got: %d", len(list))
	}
}

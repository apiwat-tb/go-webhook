package usecase_test

import (
	"errors"
	"testing"

	"github.com/apiwat-tb/go-webhook/internal/domain"
	"github.com/apiwat-tb/go-webhook/internal/usecase"
)

// fakeWebhookRepo is an in-test mock implementing usecase.WebhookRepository.
type fakeWebhookRepo struct {
	items map[string]*domain.Webhook
}

func newFakeWebhookRepo() *fakeWebhookRepo {
	return &fakeWebhookRepo{items: make(map[string]*domain.Webhook)}
}

func (f *fakeWebhookRepo) Create(w *domain.Webhook) error {
	f.items[w.ID] = w
	return nil
}

func (f *fakeWebhookRepo) GetByID(id string) (*domain.Webhook, error) {
	w, exists := f.items[id]
	if !exists {
		return nil, domain.ErrNotFound
	}
	return w, nil
}

func (f *fakeWebhookRepo) List() ([]*domain.Webhook, error) {
	var list []*domain.Webhook
	for _, w := range f.items {
		list = append(list, w)
	}
	return list, nil
}

func (f *fakeWebhookRepo) ListByEvent(event string) ([]*domain.Webhook, error) {
	var list []*domain.Webhook
	for _, w := range f.items {
		if w.Active && w.Event == event {
			list = append(list, w)
		}
	}
	return list, nil
}

func (f *fakeWebhookRepo) Update(w *domain.Webhook) error {
	if _, exists := f.items[w.ID]; !exists {
		return domain.ErrNotFound
	}
	f.items[w.ID] = w
	return nil
}

func (f *fakeWebhookRepo) Delete(id string) error {
	if _, exists := f.items[id]; !exists {
		return domain.ErrNotFound
	}
	delete(f.items, id)
	return nil
}

func TestWebhookUsecase_Create(t *testing.T) {
	repo := newFakeWebhookRepo()
	uc := usecase.NewWebhookUsecase(repo)

	// Test valid creation with generated secret
	w, err := uc.Create("https://example.com/hook", "order.created", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if w.ID == "" {
		t.Errorf("expected generated UUID, got empty")
	}
	if len(w.Secret) != 32 { // 16 bytes encoded as hex is 32 characters
		t.Errorf("expected 32 hex chars secret, got: %s", w.Secret)
	}
	if !w.Active {
		t.Errorf("expected Active=true")
	}

	// Test invalid URL validation
	_, err = uc.Create("invalid-url", "order.created", "secret")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}

	// Test missing event validation
	_, err = uc.Create("https://example.com/hook", "   ", "secret")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty event, got: %v", err)
	}
}

func TestWebhookUsecase_Get_NotFound(t *testing.T) {
	repo := newFakeWebhookRepo()
	uc := usecase.NewWebhookUsecase(repo)

	_, err := uc.Get("non-existent-id")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestWebhookUsecase_Update(t *testing.T) {
	repo := newFakeWebhookRepo()
	uc := usecase.NewWebhookUsecase(repo)

	created, err := uc.Create("https://example.com/hook", "order.created", "secret")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	newURL := "https://example.com/updated"
	newActive := false
	updated, err := uc.Update(created.ID, &newURL, nil, &newActive)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if updated.URL != newURL || updated.Active != false || updated.Event != "order.created" {
		t.Errorf("updated fields do not match expected: %+v", updated)
	}

	// Test update not found
	_, err = uc.Update("unknown", &newURL, nil, nil)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestWebhookUsecase_Delete(t *testing.T) {
	repo := newFakeWebhookRepo()
	uc := usecase.NewWebhookUsecase(repo)

	created, _ := uc.Create("https://example.com/hook", "order.created", "secret")

	if err := uc.Delete(created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Second delete should return ErrNotFound
	if err := uc.Delete(created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound on second delete, got: %v", err)
	}
}

package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/apiwat-tb/go-webhook/internal/domain"
	"github.com/google/uuid"
)

// WebhookUsecase coordinates business operations for webhook subscriptions.
type WebhookUsecase struct {
	repo WebhookRepository
}

// NewWebhookUsecase creates a new WebhookUsecase instance.
func NewWebhookUsecase(repo WebhookRepository) *WebhookUsecase {
	return &WebhookUsecase{
		repo: repo,
	}
}

// Create validates input, assigns defaults/identifiers, and saves the webhook.
func (u *WebhookUsecase) Create(rawURL, event, secret string) (*domain.Webhook, error) {
	if err := validateURL(rawURL); err != nil {
		return nil, err
	}
	if err := validateEvent(event); err != nil {
		return nil, err
	}

	trimmedEvent := strings.TrimSpace(event)

	// If no secret was provided by the caller, generate 16 cryptographically secure random bytes as hex.
	if strings.TrimSpace(secret) == "" {
		generated, err := generateSecret()
		if err != nil {
			return nil, fmt.Errorf("failed to generate webhook secret: %w", err)
		}
		secret = generated
	}

	webhook := &domain.Webhook{
		ID:        uuid.NewString(),
		URL:       rawURL,
		Event:     trimmedEvent,
		Secret:    secret,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.repo.Create(webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

// Get retrieves a webhook subscription by its ID.
func (u *WebhookUsecase) Get(id string) (*domain.Webhook, error) {
	return u.repo.GetByID(id)
}

// List returns all registered webhook subscriptions.
func (u *WebhookUsecase) List() ([]*domain.Webhook, error) {
	return u.repo.List()
}

// Update modifies an existing webhook by applying non-nil fields and re-validating.
func (u *WebhookUsecase) Update(id string, updateURL, updateEvent *string, active *bool) (*domain.Webhook, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if updateURL != nil {
		if err := validateURL(*updateURL); err != nil {
			return nil, err
		}
		existing.URL = *updateURL
	}

	if updateEvent != nil {
		if err := validateEvent(*updateEvent); err != nil {
			return nil, err
		}
		existing.Event = strings.TrimSpace(*updateEvent)
	}

	if active != nil {
		existing.Active = *active
	}

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete removes a webhook subscription by its ID.
func (u *WebhookUsecase) Delete(id string) error {
	return u.repo.Delete(id)
}

// validateURL checks that the URL is present, parses cleanly, has http/https scheme and a non-empty host.
func validateURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("%w: url is required", domain.ErrInvalidInput)
	}
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("%w: invalid url format", domain.ErrInvalidInput)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%w: url scheme must be http or https", domain.ErrInvalidInput)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%w: url host cannot be empty", domain.ErrInvalidInput)
	}
	return nil
}

// validateEvent ensures event names are non-empty after trimming whitespace.
func validateEvent(event string) error {
	if strings.TrimSpace(event) == "" {
		return fmt.Errorf("%w: event is required", domain.ErrInvalidInput)
	}
	return nil
}

// generateSecret creates a 16-byte random hex string for HMAC signing.
func generateSecret() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

package http

import (
	"net/http"

	"github.com/apiwat-tb/go-webhook/internal/usecase"
)

// NewRouter sets up standard library HTTP routing with Go 1.22 pattern matching.
func NewRouter(
	webhookUsecase *usecase.WebhookUsecase,
	dispatchUsecase *usecase.DispatchUsecase,
	verifySig func(secret string, payload []byte, sig string) bool,
	incomingSecret string,
) http.Handler {
	mux := http.NewServeMux()

	webhookHandler := NewWebhookHandler(webhookUsecase, dispatchUsecase)
	eventHandler := NewEventHandler(dispatchUsecase, verifySig, incomingSecret)

	// Webhook CRUD & Deliveries
	mux.HandleFunc("POST /webhooks", webhookHandler.Create)
	mux.HandleFunc("GET /webhooks", webhookHandler.List)
	mux.HandleFunc("GET /webhooks/{id}", webhookHandler.Get)
	mux.HandleFunc("PUT /webhooks/{id}", webhookHandler.Update)
	mux.HandleFunc("DELETE /webhooks/{id}", webhookHandler.Delete)
	mux.HandleFunc("GET /webhooks/{id}/deliveries", webhookHandler.ListDeliveries)

	// Event dispatch & incoming webhook verification
	mux.HandleFunc("POST /events", eventHandler.Dispatch)
	mux.HandleFunc("POST /incoming", eventHandler.Incoming)

	return mux
}

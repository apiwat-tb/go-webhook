package http

import (
	"github.com/apiwat-tb/go-webhook/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
)

// NewRouter sets up Fiber routing.
func NewRouter(
	webhookUsecase *usecase.WebhookUsecase,
	dispatchUsecase *usecase.DispatchUsecase,
	verifySig func(secret string, payload []byte, sig string) bool,
	incomingSecret string,
) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			"john":  "doe",
			"admin": "123456",
		},
	}))

	api := app.Group("/api")
	v1 := api.Group("/v1")
	v2 := api.Group("/v2")

	webhookHandler := NewWebhookHandler(webhookUsecase, dispatchUsecase)
	eventHandler := NewEventHandler(dispatchUsecase, verifySig, incomingSecret)

	// Webhook CRUD & Deliveries
	app.Post("/webhooks", webhookHandler.Create)
	app.Get("/webhooks", webhookHandler.List)
	app.Get("/webhooks/:id", webhookHandler.Get)
	app.Put("/webhooks/:id", webhookHandler.Update)
	app.Delete("/webhooks/:id", webhookHandler.Delete)
	app.Get("/webhooks/:id/deliveries", webhookHandler.ListDeliveries)

	// Event dispatch & incoming webhook verification
	app.Post("/events", eventHandler.Dispatch)
	app.Post("/incoming", eventHandler.Incoming)

	v1.Post("/events", eventHandler.Dispatch)
	v1.Post("/incoming", eventHandler.Incoming)

	v2.Post("/events", eventHandler.Dispatch)
	v2.Post("/incoming", eventHandler.Incoming)

	return app
}

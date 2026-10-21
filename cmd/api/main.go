package main

import (
	"log"
	"os"

	deliveryhttp "github.com/apiwat-tb/go-webhook/internal/delivery/http"
	"github.com/apiwat-tb/go-webhook/internal/infra/sender"
	"github.com/apiwat-tb/go-webhook/internal/repository/memory"
	"github.com/apiwat-tb/go-webhook/internal/usecase"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	incomingSecret := os.Getenv("INCOMING_SECRET")
	if incomingSecret == "" {
		incomingSecret = "demo-secret"
	}

	// Initialize in-memory repositories and infra sender
	webhookRepo := memory.NewWebhookRepo()
	deliveryRepo := memory.NewDeliveryRepo()
	httpSender := sender.NewHTTPSender()

	// Initialize business logic use cases
	webhookUsecase := usecase.NewWebhookUsecase(webhookRepo)
	dispatchUsecase := usecase.NewDispatchUsecase(webhookRepo, deliveryRepo, httpSender)

	// Compose router, injecting signature verification function and incoming secret
	router := deliveryhttp.NewRouter(webhookUsecase, dispatchUsecase, sender.Verify, incomingSecret)

	addr := ":" + port
	log.Printf("Server listening on http://localhost%s", addr)
	if err := router.Listen(addr); err != nil {
		log.Fatalf("Server stopped unexpectedly: %v", err)
	}
}

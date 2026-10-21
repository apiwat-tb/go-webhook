package http

import (
	"errors"
	"net/http"

	"github.com/apiwat-tb/go-webhook/internal/domain"
	"github.com/gofiber/fiber/v2"
)

// errorResponse standardizes the JSON structure for HTTP error payloads.
type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON sets the HTTP status code and serializes the value to JSON.
func writeJSON(c *fiber.Ctx, status int, v any) error {
	return c.Status(status).JSON(v)
}

// writeError maps known application/domain errors to proper HTTP status codes.
func writeError(c *fiber.Ctx, err error) error {
	var status int
	var msg string

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
		msg = err.Error()
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		msg = err.Error()
	default:
		status = http.StatusInternalServerError
		msg = "internal server error"
	}

	return writeJSON(c, status, errorResponse{Error: msg})
}

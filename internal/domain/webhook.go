package domain

import "time"

// Webhook represents a subscriber registered to receive notifications for specific events.
type Webhook struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Event     string    `json:"event"`
	Secret    string    `json:"secret"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// Delivery records an attempt to dispatch a payload to a webhook subscription.
type Delivery struct {
	ID         string    `json:"id"`
	WebhookID  string    `json:"webhook_id"`
	Event      string    `json:"event"`
	Payload    string    `json:"payload"`     // Raw JSON payload string
	StatusCode int       `json:"status_code"` // HTTP response code; 0 if the network request failed
	Success    bool      `json:"success"`     // True if response status is in the 2xx range
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

package sender

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// HTTPSender implements usecase.Sender by sending signed HTTP POST requests.
type HTTPSender struct {
	client *http.Client
}

// NewHTTPSender creates an HTTPSender with a 5-second timeout.
func NewHTTPSender() *HTTPSender {
	return &HTTPSender{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Send executes a POST request to targetURL with the payload and X-Signature header.
func (s *HTTPSender) Send(targetURL, secret string, payload []byte) (int, error) {
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", Sign(secret, payload))

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Drain any response body to allow connection reuse before returning status code.
	_, _ = io.Copy(io.Discard, resp.Body)

	return resp.StatusCode, nil
}

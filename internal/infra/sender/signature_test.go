package sender_test

import (
	"strings"
	"testing"

	"github.com/apiwat-tb/go-webhook/internal/infra/sender"
)

func TestSignature_RoundTrip(t *testing.T) {
	secret := "test-secret-key"
	payload := []byte(`{"event":"order.paid","data":{"id":123}}`)

	sig := sender.Sign(secret, payload)

	if !strings.HasPrefix(sig, "sha256=") {
		t.Fatalf("expected signature to start with 'sha256=', got: %s", sig)
	}

	if !sender.Verify(secret, payload, sig) {
		t.Errorf("expected valid signature to verify successfully")
	}
}

func TestSignature_TamperedPayload(t *testing.T) {
	secret := "test-secret-key"
	payload := []byte(`{"event":"order.paid","amount":100}`)
	tamperedPayload := []byte(`{"event":"order.paid","amount":999}`)

	sig := sender.Sign(secret, payload)

	if sender.Verify(secret, tamperedPayload, sig) {
		t.Errorf("tampered payload should not verify successfully")
	}
}

func TestSignature_WrongSecret(t *testing.T) {
	secret := "correct-secret"
	wrongSecret := "wrong-secret"
	payload := []byte(`{"message":"hello"}`)

	sig := sender.Sign(secret, payload)

	if sender.Verify(wrongSecret, payload, sig) {
		t.Errorf("verification with wrong secret should fail")
	}
}

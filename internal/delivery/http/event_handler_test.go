package http_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apiwat-tb/go-webhook/internal/infra/sender"
)

func TestEventHandler_Incoming(t *testing.T) {
	router := setupTestRouter()
	secret := "test-secret"
	payload := []byte(`{"event":"order.paid","data":{"order_id":100}}`)

	// 1. Missing signature -> 401
	reqNoSig := httptest.NewRequest(http.MethodPost, "/incoming", bytes.NewReader(payload))
	rrNoSig := httptest.NewRecorder()
	router.ServeHTTP(rrNoSig, reqNoSig)

	if rrNoSig.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing signature, got: %d", rrNoSig.Code)
	}

	// 2. Invalid signature -> 401
	reqBadSig := httptest.NewRequest(http.MethodPost, "/incoming", bytes.NewReader(payload))
	reqBadSig.Header.Set("X-Signature", "sha256=invalid")
	rrBadSig := httptest.NewRecorder()
	router.ServeHTTP(rrBadSig, reqBadSig)

	if rrBadSig.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid signature, got: %d", rrBadSig.Code)
	}

	// 3. Valid signature -> 200 {"status":"ok"}
	validSig := sender.Sign(secret, payload)
	reqValid := httptest.NewRequest(http.MethodPost, "/incoming", bytes.NewReader(payload))
	reqValid.Header.Set("X-Signature", validSig)
	rrValid := httptest.NewRecorder()
	router.ServeHTTP(rrValid, reqValid)

	if rrValid.Code != http.StatusOK {
		t.Errorf("expected 200 for valid signature, got: %d (%s)", rrValid.Code, rrValid.Body.String())
	}
}

func TestEventHandler_Dispatch(t *testing.T) {
	router := setupTestRouter()

	// Missing event -> 400
	badBody := []byte(`{"data":{"foo":"bar"}}`)
	reqBad := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(badBody))
	rrBad := httptest.NewRecorder()
	router.ServeHTTP(rrBad, reqBad)

	if rrBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing event, got: %d", rrBad.Code)
	}

	// Valid event -> 200
	goodBody := []byte(`{"event":"test.event","data":{"msg":"hello"}}`)
	reqGood := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(goodBody))
	rrGood := httptest.NewRecorder()
	router.ServeHTTP(rrGood, reqGood)

	if rrGood.Code != http.StatusOK {
		t.Errorf("expected 200 for valid dispatch, got: %d", rrGood.Code)
	}
}

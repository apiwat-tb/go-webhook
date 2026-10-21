package http_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	deliveryhttp "github.com/apiwat-tb/go-webhook/internal/delivery/http"
	"github.com/apiwat-tb/go-webhook/internal/domain"
	"github.com/apiwat-tb/go-webhook/internal/infra/sender"
	"github.com/apiwat-tb/go-webhook/internal/repository/memory"
	"github.com/apiwat-tb/go-webhook/internal/usecase"
)

// fiberHandler adapts a Fiber app to http.Handler for use with httptest.
type fiberHandler struct{ app interface {
	Test(req *http.Request, msTimeout ...int) (*http.Response, error)
} }

func (f fiberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := f.app.Test(r, -1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func setupTestRouter() http.Handler {
	webhookRepo := memory.NewWebhookRepo()
	deliveryRepo := memory.NewDeliveryRepo()
	httpSender := sender.NewHTTPSender()

	webhookUc := usecase.NewWebhookUsecase(webhookRepo)
	dispatchUc := usecase.NewDispatchUsecase(webhookRepo, deliveryRepo, httpSender)

	return fiberHandler{app: deliveryhttp.NewRouter(webhookUc, dispatchUc, sender.Verify, "test-secret")}
}

func TestWebhookHandler_CreateAndList(t *testing.T) {
	router := setupTestRouter()

	// Initial GET /webhooks should return empty array []
	reqList := httptest.NewRequest(http.MethodGet, "/webhooks", nil)
	rrList := httptest.NewRecorder()
	router.ServeHTTP(rrList, reqList)

	if rrList.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", rrList.Code)
	}
	if bytes.TrimSpace(rrList.Body.Bytes())[0] != '[' {
		t.Fatalf("expected JSON array response, got: %s", rrList.Body.String())
	}

	// POST /webhooks valid
	body := []byte(`{"url":"https://example.com/callback","event":"order.paid","secret":"mysecret"}`)
	reqCreate := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
	rrCreate := httptest.NewRecorder()
	router.ServeHTTP(rrCreate, reqCreate)

	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got: %d (%s)", rrCreate.Code, rrCreate.Body.String())
	}

	var created domain.Webhook
	if err := json.Unmarshal(rrCreate.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if created.ID == "" || created.URL != "https://example.com/callback" {
		t.Errorf("unexpected created webhook: %+v", created)
	}

	// GET /webhooks/{id}
	reqGet := httptest.NewRequest(http.MethodGet, "/webhooks/"+created.ID, nil)
	rrGet := httptest.NewRecorder()
	router.ServeHTTP(rrGet, reqGet)

	if rrGet.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", rrGet.Code)
	}

	// GET /webhooks/unknown-id -> 404
	req404 := httptest.NewRequest(http.MethodGet, "/webhooks/unknown-id", nil)
	rr404 := httptest.NewRecorder()
	router.ServeHTTP(rr404, req404)

	if rr404.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got: %d", rr404.Code)
	}
}

func TestWebhookHandler_Create_InvalidInput(t *testing.T) {
	router := setupTestRouter()

	// Missing URL
	body := []byte(`{"event":"order.paid"}`)
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing url, got: %d", rr.Code)
	}
}

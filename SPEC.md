# go-webhook – Specification (Source of Truth)

> **Audience:** AI coding models and human developers implementing this project.
> **Rule:** If code and this document conflict, this document wins. Do not add features, dependencies, or layers not described here. Keep code simple and beginner-readable.

## 1. Purpose
A small Go service for learning. It provides:
- Basic CRUD for webhook subscriptions.
- Three webhook functions: dispatch, receive (signature verification), delivery logs.
- Clean architecture with strict layer separation.

## 2. Constraints
- Go 1.22+ (uses `net/http` pattern routing, e.g. `mux.HandleFunc("GET /webhooks/{id}", ...)` and `r.PathValue("id")`).
- Module name: `github.com/apiwat-tb/go-webhook`.
- Only one external dependency allowed: `github.com/google/uuid`.
- No web framework, no ORM, no DI framework.
- Storage: in-memory, mutex-protected. Must be replaceable via interfaces.
- No auth, retries, queues, pagination, Docker, or migrations (out of scope).
- Add short comments explaining the "why" for learners. Keep functions small.

## 3. Directory Layout (exact)

```
go-webhook/
├── cmd/api/main.go
├── internal/
│   ├── domain/
│   │   ├── webhook.go          # Webhook, Delivery
│   │   └── errors.go           # ErrNotFound, ErrInvalidInput
│   ├── usecase/
│   │   ├── repository.go       # WebhookRepository, DeliveryRepository, Sender
│   │   ├── webhook_usecase.go  # CRUD
│   │   ├── dispatch_usecase.go # dispatch + list deliveries
│   │   └── *_test.go
│   ├── repository/memory/
│   │   ├── webhook_repo.go
│   │   └── delivery_repo.go
│   ├── infra/sender/
│   │   ├── http_sender.go      # implements usecase.Sender
│   │   └── signature.go        # Sign / Verify HMAC helpers
│   └── delivery/http/
│       ├── router.go
│       ├── webhook_handler.go
│       ├── event_handler.go    # /events and /incoming
│       ├── response.go         # writeJSON, writeError
│       └── *_test.go
├── go.mod
├── README.md
└── SPEC.md
```

## 4. Dependency Rules (MUST follow)

```
delivery/http  -->  usecase  -->  domain
repository/memory --> domain (implements usecase interfaces)
infra/sender      --> domain (implements usecase interfaces)
cmd/api           --> everything (composition root only)
```

- `domain` imports only the standard library.
- `usecase` imports `domain` only. It defines the interfaces it needs.
- `usecase` MUST NOT import `net/http`, `repository`, `infra`, or `delivery`.
- `delivery/http` MUST NOT import `repository` or `infra`.
- Wiring happens only in `cmd/api/main.go`.

## 5. Domain Models (`internal/domain`)

```go
type Webhook struct {
    ID        string    `json:"id"`
    URL       string    `json:"url"`
    Event     string    `json:"event"`
    Secret    string    `json:"secret"`
    Active    bool      `json:"active"`
    CreatedAt time.Time `json:"created_at"`
}

type Delivery struct {
    ID         string    `json:"id"`
    WebhookID  string    `json:"webhook_id"`
    Event      string    `json:"event"`
    Payload    string    `json:"payload"`     // raw JSON string
    StatusCode int       `json:"status_code"` // 0 if request failed
    Success    bool      `json:"success"`     // true if 2xx
    Error      string    `json:"error,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
}
```

Errors:
```go
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
)
```
Wrap validation messages: `fmt.Errorf("%w: url is required", domain.ErrInvalidInput)`.

## 6. Interfaces (`internal/usecase/repository.go`)

```go
type WebhookRepository interface {
    Create(w *domain.Webhook) error
    GetByID(id string) (*domain.Webhook, error)   // returns domain.ErrNotFound
    List() ([]*domain.Webhook, error)
    ListByEvent(event string) ([]*domain.Webhook, error) // active only
    Update(w *domain.Webhook) error               // returns domain.ErrNotFound
    Delete(id string) error                       // returns domain.ErrNotFound
}

type DeliveryRepository interface {
    Save(d *domain.Delivery) error
    ListByWebhook(webhookID string) ([]*domain.Delivery, error)
}

type Sender interface {
    // Send POSTs payload to url, signed with secret. Returns HTTP status code.
    Send(url, secret string, payload []byte) (statusCode int, err error)
}
```

## 7. Use Cases

### 7.1 WebhookUsecase (CRUD)
Constructor: `NewWebhookUsecase(repo WebhookRepository) *WebhookUsecase`

| Method | Behavior |
|---|---|
| `Create(url, event, secret string) (*Webhook, error)` | Validate; generate uuid ID; if secret empty generate 16 random bytes hex; `Active=true`; `CreatedAt=time.Now()`; save |
| `Get(id)` | Delegate to repo |
| `List()` | Delegate to repo |
| `Update(id string, url, event *string, active *bool) (*Webhook, error)` | Load existing; apply non-nil fields; re-validate; save |
| `Delete(id)` | Delegate to repo |

Validation:
- `url`: required, must parse with `url.ParseRequestURI`, scheme `http` or `https`, non-empty host.
- `event`: required, non-empty after `strings.TrimSpace`.
- Violations return `domain.ErrInvalidInput` (wrapped).

### 7.2 DispatchUsecase
Constructor: `NewDispatchUsecase(webhooks WebhookRepository, deliveries DeliveryRepository, sender Sender) *DispatchUsecase`

| Method | Behavior |
|---|---|
| `Dispatch(event string, data json.RawMessage) (DispatchResult, error)` | Validate `event` non-empty and `data` valid JSON. Build payload `{"event":..., "data":..., "timestamp":RFC3339}`. For each active webhook from `ListByEvent`: call `Sender.Send`, then save a `Delivery` (Success = 2xx, StatusCode, Error text if err). A failed delivery MUST NOT stop the loop or fail the whole call. |
| `ListDeliveries(webhookID string) ([]*Delivery, error)` | Return `ErrNotFound` if webhook does not exist; else repo list |

```go
type DispatchResult struct {
    Matched   int `json:"matched"`
    Succeeded int `json:"succeeded"`
    Failed    int `json:"failed"`
}
```
Send sequentially (no goroutines) to keep it simple.

## 8. Infrastructure

### 8.1 Memory repositories
- `map[string]*domain.Webhook` guarded by `sync.RWMutex`.
- `List()` returns items sorted by `CreatedAt` ascending.
- Return copies or be careful not to leak mutable shared pointers.
- Constructors: `NewWebhookRepo()`, `NewDeliveryRepo()`.

### 8.2 HTTP sender (`infra/sender`)
- `http.Client{Timeout: 5 * time.Second}`.
- Request: `POST`, body = payload, headers:
  - `Content-Type: application/json`
  - `X-Signature: sha256=<hex hmac>`
- Close the response body. Return `resp.StatusCode`.
- Constructor: `NewHTTPSender() *HTTPSender`.

### 8.3 Signature helpers (`infra/sender/signature.go`)
```go
func Sign(secret string, payload []byte) string          // returns "sha256=<hex>"
func Verify(secret string, payload []byte, sig string) bool // use hmac.Equal (constant time)
```
`/incoming` handler in `delivery/http` needs `Verify`. To respect dependency rules, the router receives a `verify func(secret string, payload []byte, sig string) bool` injected from `main.go` (do not import `infra` in delivery).

## 9. HTTP API

Base: `http://localhost:8080`. All bodies/responses are JSON (`Content-Type: application/json`).

### 9.1 CRUD

| Method | Path | Request body | Success | Errors |
|---|---|---|---|---|
| POST | `/webhooks` | `{"url","event","secret?"}` | 201 + Webhook | 400 |
| GET | `/webhooks` | – | 200 + `[]Webhook` (empty array, never null) | – |
| GET | `/webhooks/{id}` | – | 200 + Webhook | 404 |
| PUT | `/webhooks/{id}` | `{"url?","event?","active?"}` | 200 + Webhook | 400, 404 |
| DELETE | `/webhooks/{id}` | – | 204 no body | 404 |

### 9.2 Webhook functions

**1. Dispatch** – `POST /events`
Request: `{"event":"order.paid","data":{"order_id":1}}`
Response 200: `{"matched":2,"succeeded":1,"failed":1}`
Errors: 400 for missing event or invalid data.

**2. Receive** – `POST /incoming`
- Reads raw body and header `X-Signature`.
- Verifies the signature against a single secret from env var `INCOMING_SECRET` (default `"demo-secret"`). This keeps the demo simple.
- Valid signature: 200 `{"status":"ok"}`. Log the body.
- Missing/invalid signature: 401 `{"error":"invalid signature"}`.
- To test locally: create a webhook with `url=http://localhost:8080/incoming` and `secret=demo-secret`, then call `POST /events`.

**3. Delivery logs** – `GET /webhooks/{id}/deliveries`
Response 200: `[]Delivery` (newest first, empty array if none). 404 if webhook missing.

### 9.3 Error format and mapping
```json
{"error": "human readable message"}
```
| Error | Status |
|---|---|
| `domain.ErrInvalidInput` | 400 |
| `domain.ErrNotFound` | 404 |
| malformed JSON body | 400 |
| anything else | 500 (generic message, no internals) |

Use `errors.Is` for mapping. Centralize in `response.go`: `writeJSON(w, status, v)` and `writeError(w, err)`.

## 10. Configuration (`cmd/api/main.go`)
| Env var | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | Listen port |
| `INCOMING_SECRET` | `demo-secret` | Secret used by `/incoming` |

`main.go` creates repos, sender, usecases, router, then `http.ListenAndServe`. Log startup address with `log.Printf`.

## 11. Testing Requirements
Use standard `testing` only.
- `usecase`: CRUD validation and not-found, using in-test fake repos. Dispatch with a fake `Sender` (one success, one failing) asserting `DispatchResult` and saved deliveries.
- `infra/sender`: `Sign`/`Verify` round trip; tampered payload fails.
- `delivery/http`: handler tests with `httptest` for POST/GET `/webhooks` and `/incoming` (valid + invalid signature).
- Run: `go test ./...` must pass; `go vet ./...` must be clean.

## 12. Implementation Order
1. `go mod init`, create directories.
2. `domain`.
3. `usecase/repository.go` + `webhook_usecase.go`.
4. `repository/memory`.
5. `delivery/http` CRUD + `main.go`; manually verify with curl.
6. `infra/sender` (signature + HTTP sender).
7. `dispatch_usecase.go`, `/events`, `/webhooks/{id}/deliveries`.
8. `/incoming`.
9. Tests.
10. `README.md` with the curl examples below.

## 12.1 Curl Examples (include in README)
```bash
# Create
curl -X POST localhost:8080/webhooks -d '{"url":"http://localhost:8080/incoming","event":"order.paid","secret":"demo-secret"}'
# List
curl localhost:8080/webhooks
# Trigger event
curl -X POST localhost:8080/events -d '{"event":"order.paid","data":{"order_id":1}}'
# Delivery logs
curl localhost:8080/webhooks/<id>/deliveries
# Delete
curl -X DELETE localhost:8080/webhooks/<id>
```

## 13. Acceptance Criteria
- [ ] All endpoints in section 9 behave exactly as specified.
- [ ] Dependency rules in section 4 hold (verify with imports).
- [ ] The local loop works: create webhook to `/incoming`, trigger event, see a successful delivery log.
- [ ] A bad target URL results in a failed `Delivery` record, and the `/events` call still returns 200.
- [ ] `go build ./...`, `go vet ./...`, and `go test ./...` pass.
- [ ] No dependencies beyond `github.com/google/uuid`.

## 14. Non-Goals / Future Ideas
Retries with backoff, async workers, SQLite/Postgres repository, auth, pagination, per-webhook secret verification on receive.

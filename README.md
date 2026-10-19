# go-webhook

A lightweight Go service for managing and dispatching webhook notifications, built strictly according to [SPEC.md](SPEC.md).

## Features
- **Webhook CRUD**: Register, list, inspect, update, and remove webhook subscriptions.
- **Dispatch**: Broadcast events to active subscribers matching an event name.
- **HMAC Verification**: Sign outgoing payloads with HMAC-SHA256 (`X-Signature` header) and verify incoming deliveries.
- **Delivery Logs**: Maintain an audit log of delivery status codes, errors, and payloads.
- **Clean Architecture**: Strict separation of concerns with domain models, usecases, memory repositories, and HTTP handlers.

## Architecture

```
delivery/http  -->  usecase  -->  domain
repository/memory --> domain (implements usecase interfaces)
infra/sender      --> domain (implements usecase interfaces)
cmd/api           --> composition root (wires components)
```

## Running the Application

### Prerequisites
- Go 1.22+

### Start the Service
```bash
go run ./cmd/api
```
The server will start listening on port `8080` (or the port specified by the `PORT` environment variable).

### Run Tests
```bash
go test ./...
go vet ./...
```

## API Endpoints

### 1. Webhook CRUD
- `POST /webhooks`: Register a new subscriber.
- `GET /webhooks`: List all subscribers.
- `GET /webhooks/{id}`: Get subscriber details by ID.
- `PUT /webhooks/{id}`: Update subscriber URL, event, or active state.
- `DELETE /webhooks/{id}`: Remove subscriber.

### 2. Events & Deliveries
- `POST /events`: Trigger an event broadcast to active subscribers.
- `POST /incoming`: Verify and ingest incoming webhooks (checks `X-Signature`).
- `GET /webhooks/{id}/deliveries`: View delivery history for a webhook (newest first).

## Curl Examples

### Create a Webhook
```bash
curl -X POST localhost:8080/webhooks -d '{"url":"http://localhost:8080/incoming","event":"order.paid","secret":"demo-secret"}'
```

### List Webhooks
```bash
curl localhost:8080/webhooks
```

### Trigger Event
```bash
curl -X POST localhost:8080/events -d '{"event":"order.paid","data":{"order_id":1}}'
```

### View Delivery Logs
```bash
curl localhost:8080/webhooks/<id>/deliveries
```

### Delete Webhook
```bash
curl -X DELETE localhost:8080/webhooks/<id>
```

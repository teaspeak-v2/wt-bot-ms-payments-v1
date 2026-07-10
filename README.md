# WT-Bot Payments Microservice

`wt-bot-ms-payments-v1` manages plans, subscriptions, and payments for the WT-Bot platform.

- CRUD for plans (admin)
- CRUD for subscriptions with ownership checks
- CRUD for payments
- Service-to-service endpoints for subscription status and payment provider callbacks

## Architecture

- **HTTP API**: chi router, JSON-only responses, JWT middleware
- **Service-to-service**: `X-Service-Key` header for provider-facing endpoints
- **Persistence**: Postgres via `pgxpool`
- **Caching**: Redis for subscription caching
- **Migrations**: `golang-migrate` with embedded SQL files
- **Docs**: Swagger/OpenAPI via swag annotations
- **Logging**: structured JSON logs with `slog`

## Quickstart

```bash
make docker-up
make run
```

The API listens on `:8080` by default.

## Environment

Copy `.env.example` to `.env` and adjust values as needed.

## Commands

- `make build` — compile all packages
- `make test` — run unit tests
- `make vet` — run Go vet
- `make lint` — run golangci-lint
- `make swag` — regenerate Swagger docs
- `make migrate-up` / `make migrate-down` — run DB migrations

## API Summary

### Plans
- `GET /api/v1/plans`
- `POST /api/v1/plans` (admin)
- `GET /api/v1/plans/{id}`
- `PATCH /api/v1/plans/{id}` (admin)
- `DELETE /api/v1/plans/{id}` (admin)

### Subscriptions
- `GET /api/v1/subscriptions`
- `POST /api/v1/subscriptions`
- `GET /api/v1/subscriptions/{id}`
- `PATCH /api/v1/subscriptions/{id}`
- `DELETE /api/v1/subscriptions/{id}` (admin)
- `GET /api/v1/subscriptions/{id}/status` (service key)

### Payments
- `GET /api/v1/payments`
- `POST /api/v1/payments`
- `GET /api/v1/payments/{id}`
- `PATCH /api/v1/payments/{id}` (admin or owner)
- `DELETE /api/v1/payments/{id}` (admin)
- `POST /api/v1/payments/{id}/status` (service key)

### Health
- `GET /healthz`
- `GET /readyz`

Swagger UI: `http://localhost:8080/swagger/index.html`

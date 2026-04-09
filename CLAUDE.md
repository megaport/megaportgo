# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
# Build
go build -v ./...

# Run all unit tests
go test -v ./...

# Run a single test
go test -v -run TestPortClientTestSuite/TestBuyPort ./...

# Integration tests (requires MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY)
go test -timeout 20m -integration ./...

# Lint (golangci-lint v2.3.1, config in .golangci.yml)
golangci-lint run

# Format
gofmt -w .
```

## Architecture

This is a **flat, single-package Go SDK** (`package megaport`) for the Megaport API. Everything lives at the top level — no subdirectories.

### Client + Service Pattern

`Client` (`client.go`) is the central struct. It holds HTTP client config, authentication state, and all service fields:

```
Client
├── PortService          (port.go)
├── VXCService           (vxc.go)
├── MCRService           (mcr.go)
├── MVEService           (mve.go)
├── IXService            (ix.go)
├── LocationService      (location.go)
├── PartnerService       (partner.go)
├── ProductService       (product.go)
├── ServiceKeyService    (service_keys.go)
├── UserManagementService (user_management.go)
├── ManagedAccountService (managed_account.go)
├── BillingMarketService (billing_markets.go)
└── NATGatewayService    (nat_gateway.go)
```

Each service follows the same pattern:
1. **Interface** in the service file (e.g., `PortService`)
2. **Implementation** as `*ServiceOp` struct (e.g., `PortServiceOp`) with a `Client` field
3. **Constructor** `NewServiceName(c *Client)` — called during `Client` initialization
4. **Types** in a companion `*_types.go` file (e.g., `port_types.go`)

Shared constants and types (product types, service states, contract terms, port speeds) are in `shared_types.go`.

### Authentication

OAuth2 client credentials flow via `client.Authorize(ctx)`. Three client options:
- `WithCredentials(accessKey, secretKey)` — standard auth
- `WithAccessToken(token, expiry)` — pre-set bearer token
- `WithTokenProvider(tp)` — custom `TokenProvider` interface for external token management (e.g., WASM)

Token endpoints differ per environment (production vs staging vs development).

### Request/Response Flow

`client.NewRequest()` builds HTTP requests with auth headers, `client.Do()` executes and decodes responses. `CheckResponse()` validates status codes. Errors include trace IDs from the `Trace-Id` response header for debugging.

### Logging

Uses `log/slog` with structured JSON logging. The `sloglint` linter enforces `attr-only: true`, `context: all`, and `key-naming-case: snake` — all slog calls must use `slog.Attr` helpers (not key-value pairs) and pass context.

## Test Patterns

**Unit tests** use `testify/suite` with an embedded `ClientTestSuite` that provides `mux` (HTTP multiplexer), `server` (httptest.Server), and `client`. Register mock handlers on `mux` to simulate API responses. Tests run in parallel.

**Integration tests** (`*_integration_test.go`) are gated by the `-integration` flag. They authenticate against the staging API and create/modify real resources.

## Key Constraints

- **Location API v3 only** — v2 methods (`ListLocations`, `GetLocationByID`) are deprecated and non-functional. Always use v3 methods (`ListLocationsV3`, `GetLocationByIDV3`, `FilterLocationsByMarketCode`).
- `megaportgo` is a shared dependency of `megaport-cli` and `terraform-provider-megaport` — changes here affect both consumers.
- Valid contract terms: 1, 12, 24, 36, 48, 60 months.
- Valid MCR port speeds: 1000, 2500, 5000, 10000, 25000, 50000, 100000, 400000 Mbps.

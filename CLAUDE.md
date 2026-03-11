# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the official Go SDK for the Megaport API (`github.com/megaport/megaportgo`). It enables programmatic management of Megaport network services including Ports, MCRs (Megaport Cloud Routers), MVEs (Megaport Virtual Edges), VXCs (Virtual Cross Connects), and IX connections.

## Build and Test Commands

```bash
# Run all unit tests (mock-based)
go test ./...

# Run a single test
go test -run TestPortClientTestSuite/TestBuyPort ./...

# Run integration tests (requires API credentials)
export MEGAPORT_ACCESS_KEY=YOUR_KEY
export MEGAPORT_SECRET_KEY=YOUR_KEY
go test -timeout 20m -integration ./...

# Run linter
golangci-lint run
```

## Architecture

### Single Package Design
All code lives in a single `megaport` package. There are no subpackages - everything is exported from the root module.

### Client Pattern
- `Client` struct in `client.go` is the main entry point
- Created via `megaport.New(httpClient, opts...)` with functional options
- Services are attached to the client: `client.PortService`, `client.VXCService`, etc.
- Authentication uses OAuth2 client credentials flow via `client.Authorize(ctx)`
- Supports custom `TokenProvider` interface for external token management (e.g., WASM)

### Service Interface Pattern
Each service follows a consistent pattern:
1. Interface definition (e.g., `PortService`, `VXCService`)
2. Implementation struct (e.g., `PortServiceOp`) containing a `*Client`
3. Constructor function (e.g., `NewPortService(c *Client)`)
4. Request/Response structs for each operation

### File Organization
- `{service}.go` - Service interface, implementation, and request/response types
- `{service}_types.go` - Additional type definitions (for complex services)
- `{service}_test.go` - Unit tests using httptest mock servers
- `{service}_integration_test.go` - Integration tests against staging API

### Testing
- Unit tests use `github.com/stretchr/testify/suite` with embedded `ClientTestSuite`
- Mock HTTP responses via `httptest.NewServer`
- Integration tests are gated behind `-integration` flag
- Integration tests run against `https://api-staging.megaport.com/`

### Error Handling
Package-level errors defined in `errors.go` (e.g., `ErrLocationNotFound`, `ErrInvalidVLAN`). API errors return `*ErrorResponse` with trace ID for debugging.

## Key APIs

| Service | Purpose |
|---------|---------|
| `PortService` | Physical port management |
| `MCRService` | Cloud router management, prefix filter lists |
| `MCRLookingGlassService` | MCR routing table and BGP session visibility |
| `MVEService` | Virtual edge (SD-WAN) management |
| `VXCService` | Virtual cross-connect management, partner lookups |
| `LocationService` | Data center location queries (use V3 methods) |
| `PartnerService` | Cloud partner port lookups |
| `ProductService` | Generic product operations, resource tags |

## Important: Location API V3 Migration

The Location API v2 is deprecated. Always use V3 methods:
- `ListLocationsV3()` instead of `ListLocations()`
- `GetLocationByIDV3()` instead of `GetLocationByID()`
- `FilterLocationsByMarketCodeV3()` instead of `FilterLocationsByMarketCode()`

V3 returns `LocationV3` structs with helper methods like `HasMCRSupport()`, `GetMCRSpeeds()`.

package megaport

import (
	"os"
	"sync"
	"testing"
)

// ── Integration Test Rate Limiter ─────────────────────────────────────────────

// accTestSemaphore caps concurrent integration test suites that provision real
// infrastructure, to avoid overwhelming staging. Mirrors the terraform
// provider's acceptance-test rate limiter.
const maxConcurrentAccTests = 20

var accTestSemaphore = make(chan struct{}, maxConcurrentAccTests)

// acquireAccTestSlot blocks until a slot is available in the concurrency pool.
// Intended to be called from the top-level TestXxxIntegrationTestSuite func
// before suite.Run — releases the slot automatically via t.Cleanup when the
// suite finishes. No-ops (and skips the caller) when MEGAPORT_ACCESS_KEY is
// unset so `go test ./...` without credentials doesn't block.
func acquireAccTestSlot(t *testing.T) {
	t.Helper()
	if os.Getenv("MEGAPORT_ACCESS_KEY") == "" {
		t.Skip("integration test helper requires MEGAPORT_ACCESS_KEY to be set")
	}
	accTestSemaphore <- struct{}{}
	t.Cleanup(func() { <-accTestSemaphore })
}

// ── Location Claiming Mutexes ─────────────────────────────────────────────────
//
// Parallel suites compete for the same staging capacity. Claiming maps let
// finders hand out a distinct location per test; the claim is released via
// t.Cleanup when the caller's test method completes. Separate mutexes per
// resource type minimise contention and avoid deadlocks.

var (
	mcrClaimedMu        sync.Mutex
	mcrClaimedLocations = map[int]bool{}

	portClaimedMu        sync.Mutex
	portClaimedLocations = map[int]bool{}

	mveClaimedMu        sync.Mutex
	mveClaimedLocations = map[int]bool{}

	natClaimedMu        sync.Mutex
	natClaimedLocations = map[int]bool{}
)

// claimMCRLocation returns true if locID was not yet claimed, marking it
// claimed and registering release via t.Cleanup. Returns false if the
// location was already in use by a parallel test.
func claimMCRLocation(t *testing.T, locID int) bool {
	t.Helper()
	mcrClaimedMu.Lock()
	defer mcrClaimedMu.Unlock()
	if mcrClaimedLocations[locID] {
		return false
	}
	mcrClaimedLocations[locID] = true
	t.Cleanup(func() {
		mcrClaimedMu.Lock()
		delete(mcrClaimedLocations, locID)
		mcrClaimedMu.Unlock()
	})
	return true
}

// claimPortLocation behaves like claimMCRLocation but for Port capacity.
func claimPortLocation(t *testing.T, locID int) bool {
	t.Helper()
	portClaimedMu.Lock()
	defer portClaimedMu.Unlock()
	if portClaimedLocations[locID] {
		return false
	}
	portClaimedLocations[locID] = true
	t.Cleanup(func() {
		portClaimedMu.Lock()
		delete(portClaimedLocations, locID)
		portClaimedMu.Unlock()
	})
	return true
}

// claimMVELocation behaves like claimMCRLocation but for MVE capacity.
func claimMVELocation(t *testing.T, locID int) bool {
	t.Helper()
	mveClaimedMu.Lock()
	defer mveClaimedMu.Unlock()
	if mveClaimedLocations[locID] {
		return false
	}
	mveClaimedLocations[locID] = true
	t.Cleanup(func() {
		mveClaimedMu.Lock()
		delete(mveClaimedLocations, locID)
		mveClaimedMu.Unlock()
	})
	return true
}

// claimNATGatewayLocation behaves like claimMCRLocation but for NAT Gateway.
func claimNATGatewayLocation(t *testing.T, locID int) bool {
	t.Helper()
	natClaimedMu.Lock()
	defer natClaimedMu.Unlock()
	if natClaimedLocations[locID] {
		return false
	}
	natClaimedLocations[locID] = true
	t.Cleanup(func() {
		natClaimedMu.Lock()
		delete(natClaimedLocations, locID)
		natClaimedMu.Unlock()
	})
	return true
}

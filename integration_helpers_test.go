package megaport

import (
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// ── Method-Level Parallelism for Testify Suites ───────────────────────────────
//
// testify's suite.Run drives Test* methods sequentially via reflection and
// stores the active *testing.T on a single mutable field on the suite. Calling
// s.T().Parallel() inside a method races that field with sibling methods, so
// the standard suite shape only permits suite-level parallelism.
//
// runIntegrationMethods works around this by replacing suite.Run with explicit
// per-method dispatch. Each Test* method runs as its own t.Run subtest with
// t.Parallel(), and each subtest uses a fresh suite instance so SetT/SetupSuite
// produce per-method state. Combined with acquireAccTestSlot inside each
// subtest, the 20-slot pool now throttles at method granularity rather than
// suite granularity — long-running methods no longer block short ones behind
// them in their parent suite.

// integrationSuite is the lifecycle surface that runIntegrationMethods drives.
// All *_integration_test.go suite types satisfy this via the embedded testify
// suite.Suite (SetT) and their own SetupSuite definition.
type integrationSuite interface {
	SetT(*testing.T)
	SetupSuite()
}

// runIntegrationMethods dispatches every Test* method on suite type S as a
// parallel subtest. Use it in place of suite.Run(t, new(X)).
func runIntegrationMethods[S any, PS interface {
	*S
	integrationSuite
}](t *testing.T) {
	t.Helper()
	t.Parallel()
	if !*runIntegrationTests {
		return
	}
	var probe S
	typ := reflect.TypeOf(PS(&probe))
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		if !strings.HasPrefix(m.Name, "Test") {
			continue
		}
		t.Run(m.Name, func(t *testing.T) {
			t.Parallel()
			acquireAccTestSlot(t)
			var s S
			ps := PS(&s)
			ps.SetT(t)
			ps.SetupSuite()
			m.Func.Call([]reflect.Value{reflect.ValueOf(ps)})
		})
	}
}

// ── Integration Test Rate Limiter ─────────────────────────────────────────────

// accTestSemaphore caps concurrent integration test suites that provision real
// infrastructure, to avoid overwhelming staging. Mirrors the terraform
// provider's acceptance-test rate limiter.
const maxConcurrentAccTests = 20

var accTestSemaphore = make(chan struct{}, maxConcurrentAccTests)

// acquireAccTestSlot blocks until a slot is available in the concurrency pool.
// runIntegrationMethods calls it from each method-level subtest so the pool
// throttles at method granularity. Releases the slot automatically via
// t.Cleanup when the subtest finishes. No-ops (and skips the caller) when
// MEGAPORT_ACCESS_KEY is unset so `go test ./...` without credentials doesn't
// block.
func acquireAccTestSlot(t *testing.T) {
	t.Helper()
	if os.Getenv("MEGAPORT_ACCESS_KEY") == "" || os.Getenv("MEGAPORT_SECRET_KEY") == "" {
		t.Skip("integration test helper requires MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY to be set")
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

// releaseMCRLocation immediately removes the claim on locID so other parallel
// tests may use that location. It is safe to call even if the location is not
// currently claimed. Use this when a probe fails after a successful claim so
// that the location re-enters the pool within the same test run.
func releaseMCRLocation(locID int) {
	mcrClaimedMu.Lock()
	delete(mcrClaimedLocations, locID)
	mcrClaimedMu.Unlock()
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

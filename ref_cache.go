package megaport

import (
	"context"
	"sync"
	"time"
)

// RefCacheDefaultTTL is the recommended TTL for reference-data caches.
// Reference data such as product matrices, image catalogs, and location
// lists are effectively static during a session but can change between
// deployments, so a long-ish lazy TTL is a reasonable default. Pass it
// explicitly to NewRefCache when no project-specific TTL is required.
const RefCacheDefaultTTL = 15 * time.Minute

// Invalidatable is implemented by reference-data caches so the Client can
// drop them together on auth failure or other invalidation events.
type Invalidatable interface {
	Invalidate()
}

// RefCache is a generic TTL cache for reference data with singleflight-style
// fetch deduplication. Fetches run outside the cache mutex so Invalidate can
// always proceed — important because the Client invalidates registered caches
// from inside Do() on auth failure, and the failing call may itself be a
// cache fetcher.
type RefCache[T any] struct {
	ttl     time.Duration
	fetcher func(ctx context.Context) (T, error)

	mu        sync.Mutex
	value     T
	fetchedAt time.Time
	hasValue  bool
	// gen is bumped every time Invalidate runs. A fetch records the gen
	// it started in and only stores its result if the cache has not been
	// invalidated mid-flight.
	gen uint64
	// inflight coordinates concurrent callers around a single fetch.
	inflight *refCacheCall[T]
}

type refCacheCall[T any] struct {
	done  chan struct{}
	value T
	err   error
}

// NewRefCache returns a RefCache that delegates misses to fetcher. A zero
// ttl means the cached value never expires automatically (it can still be
// dropped via Invalidate). fetcher must not be nil.
func NewRefCache[T any](ttl time.Duration, fetcher func(ctx context.Context) (T, error)) *RefCache[T] {
	if fetcher == nil {
		panic("megaport: NewRefCache requires a non-nil fetcher")
	}
	return &RefCache[T]{ttl: ttl, fetcher: fetcher}
}

// GetOrFetch returns the cached value if still fresh, otherwise invokes the
// fetcher and caches the result. Concurrent callers share a single in-flight
// fetch (singleflight semantics); the fetcher itself runs without the cache
// mutex held so that Invalidate can run concurrently — for example when the
// Client invalidates this cache from inside Do() because the fetcher's HTTP
// call returned 401/403.
//
// For reference types (slices, maps, pointers) the returned value aliases
// the cache's stored state — callers must treat it as immutable. Mutating
// it would corrupt the value subsequent callers receive. Take a defensive
// copy at the call site if you need to mutate.
func (c *RefCache[T]) GetOrFetch(ctx context.Context) (T, error) {
	c.mu.Lock()
	if c.hasValue && (c.ttl == 0 || time.Since(c.fetchedAt) < c.ttl) {
		v := c.value
		c.mu.Unlock()
		return v, nil
	}
	if call := c.inflight; call != nil {
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		case <-call.done:
			return call.value, call.err
		}
	}
	startGen := c.gen
	call := &refCacheCall[T]{done: make(chan struct{})}
	c.inflight = call
	c.mu.Unlock()

	v, err := c.fetcher(ctx)

	c.mu.Lock()
	call.value, call.err = v, err
	// Only detach this call from the cache if it is still the active
	// in-flight pointer. Invalidate may have already cleared c.inflight
	// (so that post-invalidation joiners start a fresh fetch) — leave it
	// alone in that case.
	if c.inflight == call {
		c.inflight = nil
	}
	// Only populate the cache if Invalidate didn't fire while the
	// fetcher was running. If gen advanced, drop the result. Callers
	// that joined this call before Invalidate are already waiting on
	// call.done and receive the in-flight result; that's the standard
	// singleflight semantic. Callers arriving after Invalidate observe
	// c.inflight == nil and start a fresh fetch.
	if err == nil && c.gen == startGen {
		c.value = v
		c.fetchedAt = time.Now()
		c.hasValue = true
	}
	c.mu.Unlock()
	close(call.done)
	return v, err
}

// Invalidate drops the cached value so the next GetOrFetch will re-fetch.
// Safe to call concurrently with an in-flight GetOrFetch — the in-flight
// fetch's result will be discarded rather than stored, and any caller that
// arrives after Invalidate starts a fresh fetch rather than joining the
// existing one. Callers that already joined the in-flight fetch before
// Invalidate fired will still receive that fetch's result.
func (c *RefCache[T]) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	var zero T
	c.value = zero
	c.fetchedAt = time.Time{}
	c.hasValue = false
	c.gen++
	// Detach the in-flight call so post-invalidation callers don't join a
	// fetch that was started against the old state. The fetch goroutine
	// still completes; its result is discarded by the gen check above
	// instead of being written to the cache.
	c.inflight = nil
}

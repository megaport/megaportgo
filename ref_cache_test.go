package megaport

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRefCache_GetOrFetch_CachesValue(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c := NewRefCache(time.Minute, func(ctx context.Context) (int, error) {
		calls.Add(1)
		return 42, nil
	})

	for i := 0; i < 5; i++ {
		v, err := c.GetOrFetch(context.Background())
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if v != 42 {
			t.Fatalf("call %d: got %d, want 42", i, v)
		}
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("fetcher called %d times, want 1", got)
	}
}

func TestRefCache_GetOrFetch_RefetchesAfterTTL(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c := NewRefCache(5*time.Millisecond, func(ctx context.Context) (int, error) {
		calls.Add(1)
		return int(calls.Load()), nil
	})

	if _, err := c.GetOrFetch(context.Background()); err != nil {
		t.Fatalf("first call: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	v, err := c.GetOrFetch(context.Background())
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if v != 2 {
		t.Fatalf("got %d, want 2 (refetch after ttl)", v)
	}
}

func TestRefCache_GetOrFetch_ZeroTTLCachesForever(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c := NewRefCache(0, func(ctx context.Context) (int, error) {
		calls.Add(1)
		return 7, nil
	})

	for i := 0; i < 3; i++ {
		if _, err := c.GetOrFetch(context.Background()); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("fetcher called %d times, want 1", got)
	}
}

func TestRefCache_Invalidate_ForcesRefetch(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c := NewRefCache(time.Hour, func(ctx context.Context) (int, error) {
		return int(calls.Add(1)), nil
	})

	if _, err := c.GetOrFetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	c.Invalidate()
	v, err := c.GetOrFetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v != 2 {
		t.Fatalf("got %d, want 2 (refetch after invalidate)", v)
	}
}

func TestRefCache_GetOrFetch_DoesNotCacheErrors(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	sentinel := errors.New("boom")
	c := NewRefCache(time.Hour, func(ctx context.Context) (int, error) {
		calls.Add(1)
		if calls.Load() == 1 {
			return 0, sentinel
		}
		return 99, nil
	})

	if _, err := c.GetOrFetch(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("first call: got %v, want %v", err, sentinel)
	}
	v, err := c.GetOrFetch(context.Background())
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if v != 99 {
		t.Fatalf("got %d, want 99", v)
	}
}

// TestRefCache_Invalidate_DoesNotDeadlockDuringFetch is a regression test
// for the case where the cache's fetcher triggers an Invalidate (for example
// because the fetcher's HTTP call returns 401 and Client.Do invalidates
// registered caches). The fetcher must run outside the cache mutex so that
// Invalidate can proceed.
func TestRefCache_Invalidate_DoesNotDeadlockDuringFetch(t *testing.T) {
	t.Parallel()
	// Separate declaration + assignment so the fetcher closure can refer
	// to c before NewRefCache returns.
	var c *RefCache[int] //nolint:staticcheck // S1021: see comment above
	c = NewRefCache(time.Hour, func(ctx context.Context) (int, error) {
		// Simulate the inside-fetch invalidate (what Client.Do does on
		// 401/403). If GetOrFetch held the cache lock during the fetch
		// this would deadlock.
		c.Invalidate()
		return 1, nil
	})

	done := make(chan error, 1)
	go func() {
		_, err := c.GetOrFetch(context.Background())
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GetOrFetch deadlocked while Invalidate ran inside the fetcher")
	}
}

// TestRefCache_GetOrFetch_DiscardsResultIfInvalidatedMidFlight ensures that
// when Invalidate runs while a fetch is in progress, the in-flight result
// is dropped (not stored) so the next GetOrFetch refetches.
func TestRefCache_GetOrFetch_DiscardsResultIfInvalidatedMidFlight(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})

	// Separate declaration + assignment so the fetcher closure can refer
	// to c before NewRefCache returns.
	var c *RefCache[int] //nolint:staticcheck // S1021: see comment above
	c = NewRefCache(time.Hour, func(ctx context.Context) (int, error) {
		n := calls.Add(1)
		if n == 1 {
			close(started)
			<-release
		}
		return int(n), nil
	})

	done := make(chan int, 1)
	go func() {
		v, _ := c.GetOrFetch(context.Background())
		done <- v
	}()

	<-started
	c.Invalidate()
	close(release)
	<-done

	// First fetch returned its value to the caller, but the cache should
	// have dropped it because of the mid-flight Invalidate.
	v, err := c.GetOrFetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v != 2 {
		t.Fatalf("got %d, want 2 (cache should have refetched after mid-flight invalidate)", v)
	}
}

// TestRefCache_GetOrFetch_PostInvalidationJoinerRefetches ensures that a
// caller arriving after Invalidate while a fetch is still in flight does
// not join the pre-invalidation fetch; it must start its own fetch so it
// reads state consistent with the new generation.
func TestRefCache_GetOrFetch_PostInvalidationJoinerRefetches(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})

	c := NewRefCache(time.Hour, func(ctx context.Context) (int, error) {
		n := calls.Add(1)
		if n == 1 {
			close(started)
			<-release
		}
		return int(n), nil
	})

	// Start the first fetch and let it block in the fetcher.
	first := make(chan int, 1)
	go func() {
		v, _ := c.GetOrFetch(context.Background())
		first <- v
	}()
	<-started

	// Invalidate while the first fetch is still in flight.
	c.Invalidate()

	// A caller arriving here must NOT join the pre-invalidation fetch.
	// It should start its own and return a fresh value.
	second := make(chan int, 1)
	secondStarted := make(chan struct{})
	go func() {
		close(secondStarted)
		v, _ := c.GetOrFetch(context.Background())
		second <- v
	}()
	<-secondStarted

	// Release the first fetch so it can complete (returns 1, but the
	// gen check drops it from the cache). The second fetch is now free
	// to run — it will see c.inflight == nil and call the fetcher again,
	// observing calls == 2.
	close(release)

	if v := <-first; v != 1 {
		t.Fatalf("first fetch got %d, want 1", v)
	}
	if v := <-second; v != 2 {
		t.Fatalf("post-invalidation joiner got %d, want 2 (fresh fetch)", v)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("fetcher called %d times, want 2", got)
	}
}

// TestNewRefCache_PanicsOnNilFetcher guards the constructor contract: callers
// must provide a non-nil fetcher. Panicking at construction time is far more
// debuggable than a deferred nil-function panic the first time GetOrFetch
// runs.
func TestNewRefCache_PanicsOnNilFetcher(t *testing.T) {
	t.Parallel()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from NewRefCache with nil fetcher")
		}
		msg, ok := r.(string)
		if !ok || msg != "megaport: NewRefCache requires a non-nil fetcher" {
			t.Fatalf("unexpected panic value: %v", r)
		}
	}()
	_ = NewRefCache[int](time.Hour, nil)
}

// TestRefCache_GetOrFetch_RecoversFromFetcherPanic ensures a panicking fetcher
// does not strand c.inflight or leave call.done unclosed. The panic must
// propagate to the caller, and a subsequent GetOrFetch must be able to start
// a fresh fetch rather than blocking forever on the stuck in-flight call.
func TestRefCache_GetOrFetch_RecoversFromFetcherPanic(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c := NewRefCache(time.Hour, func(ctx context.Context) (int, error) {
		n := calls.Add(1)
		if n == 1 {
			panic("boom")
		}
		return 42, nil
	})

	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic to propagate from GetOrFetch")
			}
			if r != "boom" {
				t.Fatalf("unexpected panic value: %v", r)
			}
		}()
		_, _ = c.GetOrFetch(context.Background())
	}()

	done := make(chan int, 1)
	go func() {
		v, err := c.GetOrFetch(context.Background())
		if err != nil {
			t.Errorf("second call: unexpected error: %v", err)
		}
		done <- v
	}()

	select {
	case v := <-done:
		if v != 42 {
			t.Fatalf("second call: got %d, want 42", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("second GetOrFetch deadlocked — panicking fetcher stranded the in-flight slot")
	}
}

func TestRefCache_GetOrFetch_SerializesConcurrentCallers(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	c := NewRefCache(time.Hour, func(ctx context.Context) (int, error) {
		calls.Add(1)
		time.Sleep(20 * time.Millisecond)
		return 1, nil
	})

	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_, _ = c.GetOrFetch(context.Background())
		}()
	}
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("fetcher called %d times under concurrent load, want 1", got)
	}
}

package megaport

import (
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type closeCountingBody struct {
	*strings.Reader
	closes *int32
}

func (b closeCountingBody) Close() error {
	atomic.AddInt32(b.closes, 1)
	return nil
}

// Do must close the response body on its error-return paths, otherwise the
// underlying connection leaks.
func TestDoClosesBodyOnError(t *testing.T) {
	var closes int32

	c, err := New(nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	u, _ := url.Parse("https://example.test")
	c.BaseURL = u
	c.HTTPClient = &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       closeCountingBody{Reader: strings.NewReader(`{"message":"bad request"}`), closes: &closes},
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})}

	req, err := c.NewRequest(ctx, http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if _, err := c.Do(ctx, req, nil); err == nil {
		t.Fatal("expected Do to return an error for a 400 response")
	}
	if got := atomic.LoadInt32(&closes); got < 1 {
		t.Fatalf("response body not closed on error return (Close called %d times)", got)
	}
}

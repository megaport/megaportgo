package megaport

import (
	"bytes"
	"errors"
	"log/slog"
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

type errCloseBody struct {
	*strings.Reader
}

func (errCloseBody) Close() error { return errors.New("close failed") }

// A close error means the connection was already broken, not that the request
// failed. doDiscard backs mutations whose response payload is ignored, so it
// must not report one as a failed mutation.
func TestDoDiscardIgnoresCloseError(t *testing.T) {
	c, err := New(nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	u, _ := url.Parse("https://example.test")
	c.BaseURL = u
	c.HTTPClient = &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       errCloseBody{Reader: strings.NewReader(`{"message":"deleted"}`)},
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})}

	req, err := c.NewRequest(ctx, http.MethodDelete, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if err := c.doDiscard(ctx, req); err != nil {
		t.Fatalf("doDiscard surfaced a close error for a successful request: %v", err)
	}
}

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestDoClosesBodyOnCopyError(t *testing.T) {
	var closes int32

	c, err := New(nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	u, _ := url.Parse("https://example.test")
	c.BaseURL = u
	c.HTTPClient = &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       closeCountingBody{Reader: strings.NewReader(`some body`), closes: &closes},
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})}

	req, err := c.NewRequest(ctx, http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if _, err := c.Do(ctx, req, failingWriter{}); err == nil {
		t.Fatal("expected Do to return an error when io.Copy fails")
	}
	if got := atomic.LoadInt32(&closes); got != 1 {
		t.Fatalf("response body not closed exactly once on error return (Close called %d times)", got)
	}
}

func TestDoClosesBodyOnDecodeError(t *testing.T) {
	var closes int32

	c, err := New(nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	u, _ := url.Parse("https://example.test")
	c.BaseURL = u
	c.HTTPClient = &http.Client{Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       closeCountingBody{Reader: strings.NewReader(`not json`), closes: &closes},
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})}

	req, err := c.NewRequest(ctx, http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	var target struct {
		Message string `json:"message"`
	}
	if _, err := c.Do(ctx, req, &target); err == nil {
		t.Fatal("expected Do to return an error for malformed JSON")
	}
	if got := atomic.LoadInt32(&closes); got != 1 {
		t.Fatalf("response body not closed exactly once on error return (Close called %d times)", got)
	}
}

// With response-body logging on, Do closes the body itself and swaps in a
// replacement reader, so the deferred close lands on the replacement. The
// original body must still be closed exactly once.
func TestDoClosesBodyOnErrorWithResponseLogging(t *testing.T) {
	var closes int32

	logCapture := &bytes.Buffer{}
	c, err := New(nil,
		WithLogResponseBody(),
		WithLogHandler(NewLevelFilterHandler(slog.LevelDebug, slog.NewJSONHandler(logCapture, nil))),
	)
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
	if got := atomic.LoadInt32(&closes); got != 1 {
		t.Fatalf("response body not closed exactly once on error return (Close called %d times)", got)
	}
}

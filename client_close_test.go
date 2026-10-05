package megaport

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// newStubClient returns a client that answers every request with status and body.
func newStubClient(t *testing.T, status int, body io.ReadCloser, opts ...ClientOpt) *Client {
	t.Helper()
	rt := rtFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: body, Header: make(http.Header), Request: r}, nil
	})
	c, err := New(&http.Client{Transport: rt}, append([]ClientOpt{WithBaseURL("https://example.test")}, opts...)...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

type closeCountingBody struct {
	*strings.Reader
	closes *int32
}

func (b closeCountingBody) Close() error {
	atomic.AddInt32(b.closes, 1)
	return nil
}

// Do must close the response body on every error return, because the caller
// gets a nil response and cannot close it.
func TestDoClosesBodyOnError(t *testing.T) {
	var closes int32

	c := newStubClient(t, http.StatusBadRequest, closeCountingBody{Reader: strings.NewReader(`{"message":"bad request"}`), closes: &closes})

	req, err := c.NewRequest(ctx, http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if resp, err := c.Do(ctx, req, nil); err == nil {
		resp.Body.Close()
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

func TestDoDiscardIgnoresCloseError(t *testing.T) {
	c := newStubClient(t, http.StatusOK, errCloseBody{Reader: strings.NewReader(`{"message":"deleted"}`)})

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

	c := newStubClient(t, http.StatusOK, closeCountingBody{Reader: strings.NewReader(`some body`), closes: &closes})

	req, err := c.NewRequest(ctx, http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if resp, err := c.Do(ctx, req, failingWriter{}); err == nil {
		resp.Body.Close()
		t.Fatal("expected Do to return an error when io.Copy fails")
	}
	if got := atomic.LoadInt32(&closes); got != 1 {
		t.Fatalf("response body not closed exactly once on error return (Close called %d times)", got)
	}
}

func TestDoClosesBodyOnDecodeError(t *testing.T) {
	var closes int32

	c := newStubClient(t, http.StatusOK, closeCountingBody{Reader: strings.NewReader(`not json`), closes: &closes})

	req, err := c.NewRequest(ctx, http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	var target struct {
		Message string `json:"message"`
	}
	if resp, err := c.Do(ctx, req, &target); err == nil {
		resp.Body.Close()
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
	c := newStubClient(t, http.StatusBadRequest,
		closeCountingBody{Reader: strings.NewReader(`{"message":"bad request"}`), closes: &closes},
		WithLogResponseBody(),
		WithLogHandler(NewLevelFilterHandler(slog.LevelDebug, slog.NewJSONHandler(logCapture, nil))),
	)

	req, err := c.NewRequest(ctx, http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if resp, err := c.Do(ctx, req, nil); err == nil {
		resp.Body.Close()
		t.Fatal("expected Do to return an error for a 400 response")
	}
	if got := atomic.LoadInt32(&closes); got != 1 {
		t.Fatalf("response body not closed exactly once on error return (Close called %d times)", got)
	}
}

// drainTrackingBody keeps its reader in a named field so io.Copy cannot
// bypass Read through strings.Reader's WriteTo.
type drainTrackingBody struct {
	r       *strings.Reader
	drained bool
	closes  int
}

func (b *drainTrackingBody) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)
	if err == io.EOF {
		b.drained = true
	}
	return n, err
}

func (b *drainTrackingBody) Close() error {
	b.closes++
	return nil
}

func TestDiscardingMethodsDrainAndCloseBody(t *testing.T) {
	tests := []struct {
		name string
		call func(c *Client) (any, error)
		want any
	}{
		{"RestoreProduct", func(c *Client) (any, error) {
			return c.ProductService.RestoreProduct(ctx, "p")
		}, &RestoreProductResponse{}},
		{"ManageProductLock", func(c *Client) (any, error) {
			return c.ProductService.ManageProductLock(ctx, &ManageProductLockRequest{ProductID: "p", ShouldLock: true})
		}, &ManageProductLockResponse{}},
		{"ValidateProductOrder", func(c *Client) (any, error) {
			return nil, c.ProductService.ValidateProductOrder(ctx, map[string]string{})
		}, nil},
		{"UpdateProductResourceTags", func(c *Client) (any, error) {
			return nil, c.ProductService.UpdateProductResourceTags(ctx, "p", &UpdateProductResourceTagsRequest{})
		}, nil},
		{"DeleteMCRPrefixFilterList", func(c *Client) (any, error) {
			return c.MCRService.DeleteMCRPrefixFilterList(ctx, "m", 1)
		}, &DeleteMCRPrefixFilterListResponse{IsDeleted: true}},
		{"ModifyMCRPrefixFilterList", func(c *Client) (any, error) {
			return c.MCRService.ModifyMCRPrefixFilterList(ctx, "m", 1, &MCRPrefixFilterList{})
		}, &ModifyMCRPrefixFilterListResponse{IsUpdated: true}},
		{"UpdateMCRWithAddOn", func(c *Client) (any, error) {
			return nil, c.MCRService.UpdateMCRWithAddOn(ctx, "m", MCRAddOnRequest{AddOn: &MCRAddOnIPsecConfig{TunnelCount: 10}})
		}, nil},
		{"UpdateMCRIPsecAddOn", func(c *Client) (any, error) {
			return nil, c.MCRService.UpdateMCRIPsecAddOn(ctx, "m", "a", 10)
		}, nil},
		{"UpdateServiceKey", func(c *Client) (any, error) {
			return c.ServiceKeyService.UpdateServiceKey(ctx, &UpdateServiceKeyRequest{Key: "k"})
		}, &UpdateServiceKeyResponse{IsUpdated: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &drainTrackingBody{r: strings.NewReader(`{"message":"ok"}`)}
			got, err := tt.call(newStubClient(t, http.StatusOK, body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
			if !body.drained || body.closes != 1 {
				t.Fatalf("body drained=%v closes=%d, want drained and closed once", body.drained, body.closes)
			}
		})
	}
}

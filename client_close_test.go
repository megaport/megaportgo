package megaport

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
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

// drainTrackingBody keeps its reader in a named field so io.Copy cannot
// bypass Read through strings.Reader's WriteTo.
type drainTrackingBody struct {
	r       *strings.Reader
	drained bool
	closes  int
}

func (b *drainTrackingBody) Read(p []byte) (int, error) {
	if b.closes > 0 {
		return 0, errors.New("http: read on closed response body")
	}
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

// Do must close the response body on every error return, because the caller
// gets a nil response and cannot close it.
func TestDoClosesBodyOnError(t *testing.T) {
	body := &drainTrackingBody{r: strings.NewReader(`{"message":"bad request"}`)}
	c := newStubClient(t, http.StatusBadRequest, body)

	req, err := c.NewRequest(ctx, http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if resp, err := c.Do(ctx, req, nil); err == nil {
		resp.Body.Close()
		t.Fatal("expected Do to return an error for a 400 response")
	}
	if body.closes != 1 {
		t.Fatalf("response body not closed exactly once on error return (Close called %d times)", body.closes)
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
	body := &drainTrackingBody{r: strings.NewReader(`some body`)}
	c := newStubClient(t, http.StatusOK, body)

	req, err := c.NewRequest(ctx, http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if resp, err := c.Do(ctx, req, failingWriter{}); err == nil {
		resp.Body.Close()
		t.Fatal("expected Do to return an error when io.Copy fails")
	}
	if body.closes != 1 {
		t.Fatalf("response body not closed exactly once on error return (Close called %d times)", body.closes)
	}
}

func TestDoClosesBodyOnDecodeError(t *testing.T) {
	body := &drainTrackingBody{r: strings.NewReader(`not json`)}
	c := newStubClient(t, http.StatusOK, body)

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
	if body.closes != 1 {
		t.Fatalf("response body not closed exactly once on error return (Close called %d times)", body.closes)
	}
}

// With response-body logging on, Do closes the body itself and swaps in a
// replacement reader, so the deferred close lands on the replacement. The
// original body must still be closed exactly once.
func TestDoClosesBodyOnErrorWithResponseLogging(t *testing.T) {
	body := &drainTrackingBody{r: strings.NewReader(`{"message":"bad request"}`)}
	c := newStubClient(t, http.StatusBadRequest, body, WithLogResponseBody())

	req, err := c.NewRequest(ctx, http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if resp, err := c.Do(ctx, req, nil); err == nil {
		resp.Body.Close()
		t.Fatal("expected Do to return an error for a 400 response")
	}
	if body.closes != 1 {
		t.Fatalf("response body not closed exactly once on error return (Close called %d times)", body.closes)
	}
}

func TestDiscardingMethodsDrainAndCloseBody(t *testing.T) {
	tests := []struct {
		name string
		call func(c *Client) (any, error)
		want any
	}{
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

func TestDeleteNATGatewayDesignDrainsAndClosesBody(t *testing.T) {
	body := &drainTrackingBody{r: strings.NewReader(`{"message":"Nat gateway order item deleted successfully"}`)}
	rt := rtFunc(func(r *http.Request) (*http.Response, error) {
		resp := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Request: r}
		switch r.Method {
		case http.MethodGet:
			resp.Body = io.NopCloser(strings.NewReader(`{"data":{"provisioningStatus":"DESIGN"}}`))
		case http.MethodDelete:
			resp.Body = body
		default:
			return nil, fmt.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		return resp, nil
	})
	c, err := New(&http.Client{Transport: rt}, WithBaseURL("https://example.test"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := c.NATGatewayService.DeleteNATGateway(ctx, "n"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !body.drained || body.closes != 1 {
		t.Fatalf("body drained=%v closes=%d, want drained and closed once", body.drained, body.closes)
	}
}

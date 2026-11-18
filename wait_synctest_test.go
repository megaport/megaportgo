//go:build go1.25

//go:debug asynctimerchan=0

package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeTimeUID = "36b3f68e-2f54-4331-bf94-f8984449365f"

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// newFakeTimeClient returns a client whose API runs in process, so synctest can advance fake time.
// Each product read answers with the next status in reads; an empty status or a read past the end gets a 500.
func newFakeTimeClient(reads ...string) *Client {
	updatePaths := []string{
		"/v2/product/" + PRODUCT_MEGAPORT + "/" + fakeTimeUID,
		"/v2/product/" + PRODUCT_MCR + "/" + fakeTimeUID,
		"/v2/product/" + PRODUCT_MVE + "/" + fakeTimeUID,
		"/v2/product/" + PRODUCT_IX + "/" + fakeTimeUID,
		"/v3/product/" + PRODUCT_VXC + "/" + fakeTimeUID,
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v3/networkdesign/validate":
			fmt.Fprint(w, `{"message":"ok"}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v4/networkdesign/buy":
			var order struct {
				Solutions []struct {
					Ref string `json:"ref"`
				} `json:"solutions"`
			}
			_ = json.NewDecoder(r.Body).Decode(&order)
			solutions := "[]"
			if len(order.Solutions) > 0 {
				solutions = fmt.Sprintf(`[{"uid":%q,"ref":%q}]`, fakeTimeUID, order.Solutions[0].Ref)
			}
			fmt.Fprintf(w, `{"data":[{"technicalServiceUid":%q,"vxcJTechnicalServiceUid":%q}],"solutions":%s}`, fakeTimeUID, fakeTimeUID, solutions)
		case r.Method == http.MethodPut && slices.Contains(updatePaths, r.URL.Path):
			fmt.Fprint(w, `{"data":{}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v2/product/"+fakeTimeUID && len(reads) > 0 && reads[0] != "":
			fmt.Fprintf(w, `{"data":{"provisioningStatus":%q}}`, reads[0])
			reads = reads[1:]
		case r.Method == http.MethodGet && r.URL.Path == "/v1/solutions/"+fakeTimeUID && len(reads) > 0 && reads[0] != "":
			fmt.Fprintf(w, `{"data":{"uid":%q,"provisioningStatus":%q}}`, fakeTimeUID, reads[0])
			reads = reads[1:]
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		handler(rec, r)
		res := rec.Result()
		res.Request = r
		return res, nil
	})
	base, _ := url.Parse("https://api.megaport.test")
	return NewClient(&http.Client{Transport: transport}, base)
}

type waitCase struct {
	name      string
	call      func(ctx context.Context, c *Client) (any, error)
	firstRead time.Duration // fake time of the first status read
	failed    any           // the result when the first status read fails
	ready     any           // the result when the product turns ready on the second read
}

func waitCases() []waitCase {
	const poll = 30 * time.Second
	portOrder := &BuyPortResponse{TechnicalServiceUIDs: []string{fakeTimeUID}}
	mcrOrder := &BuyMCRResponse{TechnicalServiceUID: fakeTimeUID}
	mveOrder := &BuyMVEResponse{TechnicalServiceUID: fakeTimeUID}
	vxcOrder := &BuyVXCResponse{TechnicalServiceUID: fakeTimeUID}
	ixOrder := &BuyIXResponse{TechnicalServiceUID: fakeTimeUID}
	dciOrder := &BuyDCIResponse{SolutionUID: fakeTimeUID}
	return []waitCase{
		{
			name: "BuyDCI",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.SolutionService.BuyDCI(ctx, &BuyDCIRequest{Name: "test-dci", AEndLocationID: 143, BEndLocationID: 145, PortSpeed: 10000, Term: 12, WaitForProvision: true})
			},
			firstRead: poll, failed: dciOrder, ready: dciOrder,
		},
		{
			name: "BuyPort",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.PortService.BuyPort(ctx, &BuyPortRequest{Name: "test-port", Term: 12, PortSpeed: 10000, LocationId: 226, WaitForProvision: true})
			},
			firstRead: poll, failed: portOrder, ready: portOrder,
		},
		{
			name: "BuyMCR",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.MCRService.BuyMCR(ctx, &BuyMCRRequest{Name: "test-mcr", Term: 1, PortSpeed: 1000, LocationID: 1, WaitForProvision: true})
			},
			firstRead: poll, failed: mcrOrder, ready: mcrOrder,
		},
		{
			name: "BuyMVE",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.MVEService.BuyMVE(ctx, &BuyMVERequest{Name: "test-mve", Term: 12, LocationID: 1, WaitForProvision: true})
			},
			firstRead: poll, failed: mveOrder, ready: mveOrder,
		},
		{
			name: "BuyVXC",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.VXCService.BuyVXC(ctx, &BuyVXCRequest{PortUID: "9b1c46c7-1e8d-4035-bf38-1bc60d346d57", VXCName: "test-vxc", RateLimit: 50, Term: 1, WaitForProvision: true})
			},
			firstRead: poll, failed: vxcOrder, ready: vxcOrder,
		},
		{
			name: "BuyIX",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.IXService.BuyIX(ctx, &BuyIXRequest{ProductUID: "9b1c46c7-1e8d-4035-bf38-1bc60d346d57", Name: "test-ix", NetworkServiceType: "Los Angeles IX", ASN: 12345, RateLimit: 500, VLAN: 2001, WaitForProvision: true})
			},
			firstRead: poll, failed: ixOrder, ready: ixOrder,
		},
		{
			name: "ModifyPort",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.PortService.ModifyPort(ctx, &ModifyPortRequest{PortID: fakeTimeUID, Name: "renamed", WaitForUpdate: true})
			},
			firstRead: poll, ready: &ModifyPortResponse{IsUpdated: true},
		},
		{
			name: "ModifyMCR",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.MCRService.ModifyMCR(ctx, &ModifyMCRRequest{MCRID: fakeTimeUID, Name: "renamed", WaitForUpdate: true})
			},
			firstRead: poll, ready: &ModifyMCRResponse{IsUpdated: true},
		},
		{
			name: "ModifyMVE",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.MVEService.ModifyMVE(ctx, &ModifyMVERequest{MVEID: fakeTimeUID, Name: "renamed", WaitForUpdate: true})
			},
			firstRead: poll, ready: &ModifyMVEResponse{MVEUpdated: true},
		},
		{
			name: "UpdateVXC",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.VXCService.UpdateVXC(ctx, fakeTimeUID, &UpdateVXCRequest{Name: PtrTo("renamed"), WaitForUpdate: true})
			},
			firstRead: poll, ready: &VXC{ProvisioningStatus: SERVICE_LIVE},
		},
		{
			name: "UpdateIX",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.IXService.UpdateIX(ctx, fakeTimeUID, &UpdateIXRequest{Name: PtrTo("renamed"), WaitForUpdate: true})
			},
			firstRead: poll, ready: &IX{ProvisioningStatus: SERVICE_LIVE},
		},
		{
			name: "WaitForMCRReady",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.MCRService.WaitForMCRReady(ctx, fakeTimeUID, 0)
			},
			firstRead: 0,
		},
	}
}

// TestWaitStatusReadFails tests that each wait returns the status read error on its first read.
func TestWaitStatusReadFails(t *testing.T) {
	t.Parallel()
	for _, tc := range waitCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				c := newFakeTimeClient("")
				start := time.Now()

				got, err := tc.call(t.Context(), c)

				var apiErr *ErrorResponse
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, http.StatusInternalServerError, apiErr.Response.StatusCode)
				if tc.failed == nil {
					assert.Nil(t, got)
				} else {
					assert.Equal(t, tc.failed, got)
				}
				assert.Equal(t, tc.firstRead, time.Since(start))
			})
		})
	}
}

// TestWaitReadyOnSecondRead tests that each wait returns success when the product turns ready on its second read.
func TestWaitReadyOnSecondRead(t *testing.T) {
	t.Parallel()
	for _, tc := range waitCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				c := newFakeTimeClient("DEPLOYABLE", SERVICE_LIVE)
				start := time.Now()

				got, err := tc.call(t.Context(), c)

				require.NoError(t, err)
				if tc.ready == nil {
					assert.Nil(t, got)
				} else {
					assert.Equal(t, tc.ready, got)
				}
				assert.Equal(t, tc.firstRead+30*time.Second, time.Since(start))
			})
		})
	}
}

// TestWaitForMCRReadyContextEnds tests that WaitForMCRReady returns the context's own error when the context ends the wait.
func TestWaitForMCRReadyContextEnds(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		c := newFakeTimeClient("DEPLOYABLE")
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		err := c.MCRService.WaitForMCRReady(ctx, fakeTimeUID, 0)

		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

// TestWaitUntilReady tests the shared wait loop on its own.
func TestWaitUntilReady(t *testing.T) {
	t.Parallel()
	const what = "Port " + fakeTimeUID + " to provision"
	notReady := func() (bool, error) { return false, nil }

	t.Run("zero timeout waits 5 minutes", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()

			err := waitUntilReady(t.Context(), 0, what, notReady)

			assert.EqualError(t, err, "time expired waiting for "+what)
			assert.Equal(t, 5*time.Minute, time.Since(start))
		})
	})

	t.Run("ready on the second check", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			checks := 0

			err := waitUntilReady(t.Context(), time.Hour, what, func() (bool, error) {
				checks++
				return checks == 2, nil
			})

			require.NoError(t, err)
			assert.Equal(t, time.Minute, time.Since(start))
		})
	})

	t.Run("ready when the last tick lands with the timeout", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			checks := 0

			err := waitUntilReady(t.Context(), time.Minute, what, func() (bool, error) {
				checks++
				return checks == 2, nil
			})

			require.NoError(t, err)
			assert.Equal(t, time.Minute, time.Since(start))
		})
	})

	t.Run("check error is returned as it is", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			errCheck := errors.New("status read failed")
			start := time.Now()

			err := waitUntilReady(t.Context(), time.Hour, what, func() (bool, error) { return false, errCheck })

			assert.Equal(t, errCheck, err)
			assert.Equal(t, errCheck, withContextCause(err))
			assert.Equal(t, 30*time.Second, time.Since(start))
		})
	})

	t.Run("context end exposes the context error only through withContextCause", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
			defer cancel()
			start := time.Now()

			err := waitUntilReady(ctx, time.Hour, what, notReady)

			assert.EqualError(t, err, "context expired waiting for "+what)
			assert.NotErrorIs(t, err, context.DeadlineExceeded)
			assert.EqualError(t, withContextCause(err), "context expired waiting for "+what+": context deadline exceeded")
			assert.ErrorIs(t, withContextCause(err), context.DeadlineExceeded)
			assert.Equal(t, 45*time.Second, time.Since(start))
		})
	})
}

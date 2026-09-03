package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// MCRLookingGlassClientTestSuite tests the MCR Looking Glass Service.
type MCRLookingGlassClientTestSuite struct {
	ClientTestSuite
}

func TestMCRLookingGlassClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MCRLookingGlassClientTestSuite))
}

func (suite *MCRLookingGlassClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url

	// Keep the poll cadence short for every test, so a regression that stops the
	// first poll from returning fails fast instead of running to the package
	// timeout. Tests that need a different cadence set their own values.
	op, ok := suite.client.MCRLookingGlassService.(*MCRLookingGlassServiceOp)
	suite.Require().True(ok)
	op.pollInterval = 5 * time.Millisecond
	op.pollTimeout = 2 * time.Second
}

func (suite *MCRLookingGlassClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// testRouteMCRUID is the MCR the route diagnostics tests query.
const testRouteMCRUID = "36b3f68e-2f54-4331-bf94-f8984449365f"

// serveRouteDiagnostics answers the submit request at routeSuffix with
// operationID, and the operation endpoint with dataBlob. The returned values
// hold the query the client submitted, for the caller to assert after the call.
func (suite *MCRLookingGlassClientTestSuite) serveRouteDiagnostics(routeSuffix, operationID, dataBlob string) *url.Values {
	submitted := &url.Values{}
	base := "/v2/product/mcr2/" + testRouteMCRUID + "/diagnostics/routes"

	suite.mux.HandleFunc(base+routeSuffix, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		*submitted = r.URL.Query()
		fmt.Fprintf(w, `{"message":"Please retrieve the result with the operation id","terms":"","data":%q}`, operationID)
	})

	suite.mux.HandleFunc(base+"/operation", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal(operationID, r.URL.Query().Get("operationId"))
		fmt.Fprintf(w, `{"message":"Diagnostic result retrieved successfully","terms":"","data":%s}`, dataBlob)
	})

	return submitted
}

// TestListIPRoutes tests the ListIPRoutes method.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutes() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID
	operationID := "0f1f3b5a-1f5b-4c0e-9d0a-8f4b1c2d3e4f"

	want := []*LookingGlassIPRoute{
		{
			Prefix:   "10.0.1.0/24",
			Protocol: "BGP",
			Distance: 20,
			Metric:   0,
			NextHop: LookingGlassRouteNextHop{
				IP:  "169.254.0.1",
				VXC: LookingGlassRouteVXCRef{ID: "11111111-2222-4333-8444-555555555555", Name: "Test VXC"},
			},
		},
		{
			Prefix:   "192.168.0.0/16",
			Protocol: "STATIC",
			Distance: 1,
			NextHop: LookingGlassRouteNextHop{
				IP:  "10.0.0.2",
				VXC: LookingGlassRouteVXCRef{ID: "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee", Name: "Static VXC"},
			},
		},
	}

	dataBlob := `[
		{
			"distance": 20,
			"metric": 0,
			"nextHop": {"ip": "169.254.0.1", "vxc": {"id": "11111111-2222-4333-8444-555555555555", "name": "Test VXC"}},
			"prefix": "10.0.1.0/24",
			"protocol": "BGP"
		},
		{
			"distance": 1,
			"nextHop": {"ip": "10.0.0.2", "vxc": {"id": "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee", "name": "Static VXC"}},
			"prefix": "192.168.0.0/16",
			"protocol": "STATIC"
		}
	]`

	submitted := suite.serveRouteDiagnostics("/ip", operationID, dataBlob)

	got, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.NoError(err)
	suite.Equal(want, got)
	suite.Equal("true", submitted.Get("async"))
	suite.Empty(submitted.Get("ip_address"))
}

// TestListBGPRoutes tests the ListBGPRoutes method.
func (suite *MCRLookingGlassClientTestSuite) TestListBGPRoutes() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID
	operationID := "5c4b3a29-1807-4f6e-9d5c-3b2a19087f6e"

	want := []*LookingGlassBGPRoute{
		{
			Prefix:       "10.0.1.0/24",
			ASPath:       "64512 64512",
			Origin:       "incomplete",
			Source:       "169.254.0.1",
			LocalPref:    100,
			MED:          58880,
			Weight:       32768,
			Best:         true,
			External:     true,
			Valid:        true,
			Since:        "2022-12-06T01:47:41+0000",
			Communities:  []string{"64512:1"},
			AdvertisedTo: []string{"168.254.0.5", "168.254.0.9"},
			NextHop: LookingGlassRouteNextHop{
				IP:  "169.254.0.1",
				VXC: LookingGlassRouteVXCRef{ID: "11111111-2222-4333-8444-555555555555", Name: "Test VXC"},
			},
		},
	}

	dataBlob := `[
		{
			"prefix": "10.0.1.0/24",
			"best": true,
			"communities": ["64512:1"],
			"external": true,
			"med": 58880,
			"origin": "incomplete",
			"since": "2022-12-06T01:47:41+0000",
			"source": "169.254.0.1",
			"valid": true,
			"weight": 32768,
			"advertisedTo": ["168.254.0.5", "168.254.0.9"],
			"asPath": "64512 64512",
			"localPref": 100,
			"nextHop": {"ip": "169.254.0.1", "vxc": {"id": "11111111-2222-4333-8444-555555555555", "name": "Test VXC"}}
		}
	]`

	submitted := suite.serveRouteDiagnostics("/bgp", operationID, dataBlob)

	got, err := lgSvc.ListBGPRoutes(ctx, mcrUID)
	suite.NoError(err)
	suite.Equal(want, got)
	suite.Equal("true", submitted.Get("async"))
}

// TestListBGPNeighborRoutes tests that the neighbor request carries the
// direction and the peer IP address.
func (suite *MCRLookingGlassClientTestSuite) TestListBGPNeighborRoutes() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID
	operationID := "7e6d5c4b-3a29-4081-9f7e-5c4b3a291807"

	want := []*LookingGlassBGPRoute{
		{
			Prefix:    "10.0.1.0/24",
			ASPath:    "64512",
			Origin:    "IGP",
			Source:    "169.254.0.1",
			LocalPref: 100,
			Best:      true,
			External:  true,
			Valid:     true,
			NextHop: LookingGlassRouteNextHop{
				IP:  "169.254.0.1",
				VXC: LookingGlassRouteVXCRef{ID: "11111111-2222-4333-8444-555555555555", Name: "Test VXC"},
			},
		},
	}

	dataBlob := `[
		{
			"prefix": "10.0.1.0/24",
			"asPath": "64512",
			"origin": "IGP",
			"source": "169.254.0.1",
			"localPref": 100,
			"best": true,
			"external": true,
			"valid": true,
			"nextHop": {"ip": "169.254.0.1", "vxc": {"id": "11111111-2222-4333-8444-555555555555", "name": "Test VXC"}}
		}
	]`

	submitted := suite.serveRouteDiagnostics("/bgp/neighbor", operationID, dataBlob)

	got, err := lgSvc.ListBGPNeighborRoutes(ctx, &ListBGPNeighborRoutesRequest{
		MCRID:         mcrUID,
		PeerIPAddress: "169.254.0.1",
		Direction:     BGPRouteDirectionReceived,
	})
	suite.NoError(err)
	suite.Equal(want, got)
	suite.Equal("RECEIVED", submitted.Get("direction"))
	suite.Equal("169.254.0.1", submitted.Get("peer_ip_address"))
	suite.Equal("true", submitted.Get("async"))
}

// TestRouteDiagnosticsRequestURLs pins the exact URL of every route request.
// A path or query parameter invented by a later change fails here.
func (suite *MCRLookingGlassClientTestSuite) TestRouteDiagnosticsRequestURLs() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID
	operationID := "8f7e6d5c-4b3a-4192-8081-6d5c4b3a2918"

	var seen []string
	suite.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		seen = append(seen, r.URL.String())
		if r.URL.Query().Get("operationId") != "" {
			fmt.Fprint(w, `{"message":"ok","terms":"","data":[]}`)
			return
		}
		fmt.Fprintf(w, `{"message":"ok","terms":"","data":%q}`, operationID)
	})

	base := "/v2/product/mcr2/" + mcrUID + "/diagnostics/routes"
	pollURL := base + "/operation?operationId=" + operationID

	tests := []struct {
		name    string
		call    func() error
		wantURL string
	}{
		{
			name:    "ip routes",
			call:    func() error { _, err := lgSvc.ListIPRoutes(ctx, mcrUID); return err },
			wantURL: base + "/ip?async=true",
		},
		{
			name: "ip routes with filter",
			call: func() error {
				_, err := lgSvc.ListIPRoutesWithFilter(ctx, &ListIPRoutesRequest{MCRID: mcrUID, IPFilter: "10.0.1.0/24"})
				return err
			},
			wantURL: base + "/ip?async=true&ip_address=10.0.1.0%2F24",
		},
		{
			name:    "bgp routes",
			call:    func() error { _, err := lgSvc.ListBGPRoutes(ctx, mcrUID); return err },
			wantURL: base + "/bgp?async=true",
		},
		{
			name: "bgp routes with filter",
			call: func() error {
				_, err := lgSvc.ListBGPRoutesWithFilter(ctx, &ListBGPRoutesRequest{MCRID: mcrUID, IPFilter: "10.0.1.1"})
				return err
			},
			wantURL: base + "/bgp?async=true&ip_address=10.0.1.1",
		},
		{
			name: "bgp neighbor routes",
			call: func() error {
				_, err := lgSvc.ListBGPNeighborRoutes(ctx, &ListBGPNeighborRoutesRequest{
					MCRID:         mcrUID,
					PeerIPAddress: "169.254.0.1",
					Direction:     BGPRouteDirectionAdvertised,
				})
				return err
			},
			wantURL: base + "/bgp/neighbor?async=true&direction=ADVERTISED&peer_ip_address=169.254.0.1",
		},
	}

	for _, tt := range tests {
		seen = nil
		suite.NoError(tt.call(), tt.name)
		suite.Equal([]string{tt.wantURL, pollURL}, seen, tt.name)
	}
}

// TestListIPRoutesPendingThenComplete covers the pending half of the poll
// contract: the operation endpoint answers 202 until the result is ready.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesPendingThenComplete() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID
	operationID := "9a8f7e6d-5c4b-42a3-9192-7e6d5c4b3a29"

	submitPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/ip", mcrUID)
	suite.mux.HandleFunc(submitPath, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprintf(w, `{"message":"ok","terms":"","data":%q}`, operationID)
	})

	var calls atomic.Int32
	operationPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(operationPath, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		if calls.Add(1) < 3 {
			// The pending response carries no data key.
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"message":"Data result will be available soon","terms":""}`)
			return
		}
		fmt.Fprint(w, `{"message":"Diagnostic result retrieved successfully","terms":"","data":[
			{"distance":20,"metric":0,"prefix":"10.0.1.0/24","protocol":"BGP",
			 "nextHop":{"ip":"169.254.0.1","vxc":{"id":"11111111-2222-4333-8444-555555555555","name":"Test VXC"}}}
		]}`)
	})

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	got, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.NoError(err)
	suite.Len(got, 1)
	suite.Equal("10.0.1.0/24", got[0].Prefix)
	suite.Equal(int32(3), calls.Load())
}

// TestListIPRoutesCallerDeadlineWins pins that a caller deadline is used as
// given. The SDK timeout here is far shorter than the run, so wrapping the
// caller's context in it as well would end the poll early.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesCallerDeadlineWins() {
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID
	operationID := "1c2d3e4f-5a6b-4c7d-8e9f-0a1b2c3d4e5f"

	op, ok := lgSvc.(*MCRLookingGlassServiceOp)
	suite.Require().True(ok)
	op.pollTimeout = 20 * time.Millisecond

	submitPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/ip", mcrUID)
	suite.mux.HandleFunc(submitPath, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprintf(w, `{"message":"ok","terms":"","data":%q}`, operationID)
	})

	start := time.Now()
	operationPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(operationPath, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		if time.Since(start) < 100*time.Millisecond {
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"message":"Data result will be available soon","terms":""}`)
			return
		}
		fmt.Fprint(w, `{"message":"Diagnostic result retrieved successfully","terms":"","data":[
			{"distance":20,"metric":0,"prefix":"10.0.1.0/24","protocol":"BGP",
			 "nextHop":{"ip":"169.254.0.1","vxc":{"id":"11111111-2222-4333-8444-555555555555","name":"Test VXC"}}}
		]}`)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	got, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.NoError(err)
	suite.Len(got, 1)
}

// TestListBGPRoutesEmptyResult covers the other half of the poll contract: a
// 200 with an empty array is a real result, so the poll stops.
func (suite *MCRLookingGlassClientTestSuite) TestListBGPRoutesEmptyResult() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID
	operationID := "a9b8c7d6-e5f4-4312-8291-8f7e6d5c4b3a"

	submitPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/bgp", mcrUID)
	suite.mux.HandleFunc(submitPath, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprintf(w, `{"message":"ok","terms":"","data":%q}`, operationID)
	})

	var calls atomic.Int32
	operationPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(operationPath, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		calls.Add(1)
		fmt.Fprint(w, `{"message":"Diagnostic result retrieved successfully","terms":"","data":[]}`)
	})

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	got, err := lgSvc.ListBGPRoutes(ctx, mcrUID)
	suite.NoError(err)
	suite.NotNil(got)
	suite.Empty(got)
	suite.Equal(int32(1), calls.Load())
}

// TestRouteRequestValidation covers the guards that run before any request.
func (suite *MCRLookingGlassClientTestSuite) TestRouteRequestValidation() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID

	suite.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("validation should reject the request before it is sent", r.URL.String())
	})

	_, err := lgSvc.ListIPRoutesWithFilter(ctx, nil)
	suite.ErrorIs(err, ErrListIPRoutesRequestNil)
	_, err = lgSvc.ListIPRoutes(ctx, "")
	suite.ErrorIs(err, ErrMCRDiagnosticsMCRUIDRequired)

	_, err = lgSvc.ListBGPRoutesWithFilter(ctx, nil)
	suite.ErrorIs(err, ErrListBGPRoutesRequestNil)
	_, err = lgSvc.ListBGPRoutes(ctx, "")
	suite.ErrorIs(err, ErrMCRDiagnosticsMCRUIDRequired)

	_, err = lgSvc.ListBGPNeighborRoutes(ctx, nil)
	suite.ErrorIs(err, ErrListBGPNeighborRoutesRequestNil)
	_, err = lgSvc.ListBGPNeighborRoutes(ctx, &ListBGPNeighborRoutesRequest{
		PeerIPAddress: "169.254.0.1",
		Direction:     BGPRouteDirectionReceived,
	})
	suite.ErrorIs(err, ErrMCRDiagnosticsMCRUIDRequired)
	_, err = lgSvc.ListBGPNeighborRoutes(ctx, &ListBGPNeighborRoutesRequest{
		MCRID:     mcrUID,
		Direction: BGPRouteDirectionReceived,
	})
	suite.ErrorIs(err, ErrMCRDiagnosticsPeerIPRequired)
	_, err = lgSvc.ListBGPNeighborRoutes(ctx, &ListBGPNeighborRoutesRequest{
		MCRID:         mcrUID,
		PeerIPAddress: "169.254.0.1",
	})
	suite.ErrorIs(err, ErrMCRDiagnosticsDirectionInvalid)
	_, err = lgSvc.ListBGPNeighborRoutes(ctx, &ListBGPNeighborRoutesRequest{
		MCRID:         mcrUID,
		PeerIPAddress: "169.254.0.1",
		Direction:     "sideways",
	})
	suite.ErrorIs(err, ErrMCRDiagnosticsDirectionInvalid)
}

// TestListIPRoutesError tests that an API error reaches the caller.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesError() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID

	submitPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/ip", mcrUID)
	suite.mux.HandleFunc(submitPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message": "MCR not found", "data": ""}`)
	})

	_, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.Error(err)
}

// TestListIPRoutesMissingOperationID tests that a submit response with no
// operation ID is reported instead of polled.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesMissingOperationID() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID

	submitPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/ip", mcrUID)
	suite.mux.HandleFunc(submitPath, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, `{"message":"ok","terms":"","data":""}`)
	})

	_, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.ErrorIs(err, ErrMCRDiagnosticsOperationEmpty)
}

// TestListIPRoutesPollTimeout tests that a never-ready operation stops at the
// SDK-managed timeout instead of polling forever.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesPollTimeout() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID
	operationID := "b8c7d6e5-f4a3-4213-9182-9f8e7d6c5b4a"

	op, ok := lgSvc.(*MCRLookingGlassServiceOp)
	suite.Require().True(ok)
	op.pollTimeout = 50 * time.Millisecond

	submitPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/ip", mcrUID)
	suite.mux.HandleFunc(submitPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"ok","terms":"","data":%q}`, operationID)
	})

	operationPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(operationPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"message":"Data result will be available soon","terms":""}`)
	})

	_, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.ErrorIs(err, ErrMCRDiagnosticsTimeout)
}

// TestListIPRoutesPollError tests that an API error partway through the poll
// reaches the caller and stops the loop instead of polling to the timeout.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesPollError() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID
	operationID := "1f2e3d4c-5b6a-4798-8071-2e3d4c5b6a79"

	submitPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/ip", mcrUID)
	suite.mux.HandleFunc(submitPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"ok","terms":"","data":%q}`, operationID)
	})

	var calls atomic.Int32
	operationPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(operationPath, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"message":"Data result will be available soon","terms":""}`)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"message":"diagnostics backend unavailable"}`)
	})

	_, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.Error(err)
	suite.NotErrorIs(err, ErrMCRDiagnosticsTimeout)
	suite.Equal(int32(2), calls.Load())
}

// TestListIPRoutesCallerCancelled tests that cancelling the caller's context
// mid-poll reports the caller's error, not the SDK timeout sentinel.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesCallerCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := testRouteMCRUID
	operationID := "2a3b4c5d-6e7f-4081-9203-4c5d6e7f8091"

	submitPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/ip", mcrUID)
	suite.mux.HandleFunc(submitPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"ok","terms":"","data":%q}`, operationID)
	})

	operationPath := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(operationPath, func(w http.ResponseWriter, r *http.Request) {
		cancel()
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"message":"Data result will be available soon","terms":""}`)
	})

	_, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.ErrorIs(err, context.Canceled)
	suite.NotErrorIs(err, ErrMCRDiagnosticsTimeout)
}

// TestRouteDiagnosticsEscapesMCRUID tests that an MCR UID needing escaping is
// escaped in the path rather than changing the URL the API sees.
func (suite *MCRLookingGlassClientTestSuite) TestRouteDiagnosticsEscapesMCRUID() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "not a uid#frag"
	operationID := "3b4c5d6e-7f80-4192-8314-5d6e7f809142"

	var seen []string
	suite.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.String())
		if r.URL.Query().Get("operationId") != "" {
			fmt.Fprint(w, `{"message":"ok","terms":"","data":[]}`)
			return
		}
		fmt.Fprintf(w, `{"message":"ok","terms":"","data":%q}`, operationID)
	})

	_, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.NoError(err)

	base := "/v2/product/mcr2/not%20a%20uid%23frag/diagnostics/routes"
	suite.Equal([]string{base + "/ip?async=true", base + "/operation?operationId=" + operationID}, seen)
}

// TestRouteDiagnosticsEscapesOperationID pins that the operation ID the API
// hands back is escaped before it goes into the poll URL.
func (suite *MCRLookingGlassClientTestSuite) TestRouteDiagnosticsEscapesOperationID() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	operationID := "op id&async=false"

	var pollURL string
	base := "/v2/product/mcr2/" + testRouteMCRUID + "/diagnostics/routes"
	suite.mux.HandleFunc(base+"/ip", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"ok","terms":"","data":%q}`, operationID)
	})
	suite.mux.HandleFunc(base+"/operation", func(w http.ResponseWriter, r *http.Request) {
		pollURL = r.URL.String()
		suite.Equal(operationID, r.URL.Query().Get("operationId"))
		fmt.Fprint(w, `{"message":"ok","terms":"","data":[]}`)
	})

	_, err := lgSvc.ListIPRoutes(ctx, testRouteMCRUID)
	suite.NoError(err)
	suite.Equal(base+"/operation?operationId=op+id%26async%3Dfalse", pollURL)
}

// TestDiagnosticsPollDefaults pins the fallback every caller gets when nothing
// overrides it. Only the tests set these fields, so nothing else covers the
// fallback. A zero interval would panic time.NewTicker, and an interval that
// does not fit inside the timeout would allow only one poll.
func (suite *MCRLookingGlassClientTestSuite) TestDiagnosticsPollDefaults() {
	svc := &MCRLookingGlassServiceOp{}

	suite.Equal(mcrDiagnosticsPollInterval, svc.diagnosticsPollInterval())
	suite.Equal(mcrDiagnosticsPollTimeout, svc.diagnosticsPollTimeout())
	suite.Greater(svc.diagnosticsPollInterval(), time.Duration(0))
	suite.Greater(svc.diagnosticsPollTimeout(), time.Duration(0))
	suite.Less(svc.diagnosticsPollInterval(), svc.diagnosticsPollTimeout())
}

// TestRouteDiagnosticsPollsImmediately pins the poll before the ticker starts.
// The interval here is longer than the suite poll timeout, so an already-ready
// result can only arrive if that first poll runs.
func (suite *MCRLookingGlassClientTestSuite) TestRouteDiagnosticsPollsImmediately() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService

	op, ok := lgSvc.(*MCRLookingGlassServiceOp)
	suite.Require().True(ok)
	op.pollInterval = time.Hour

	suite.serveRouteDiagnostics("/ip", "4c5d6e7f-8091-42a3-b4c5-d6e7f8091a2b", `[]`)

	_, err := lgSvc.ListIPRoutes(ctx, testRouteMCRUID)
	suite.NoError(err)
}

// TestListIPRoutesSubmitNotJSON covers the submit leg of the proxy error page
// that TestRouteOperationUnexpected200Shapes covers on the poll leg.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesSubmitNotJSON() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService

	base := "/v2/product/mcr2/" + testRouteMCRUID + "/diagnostics/routes"
	suite.mux.HandleFunc(base+"/ip", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<html>gateway error</html>")
	})

	_, err := lgSvc.ListIPRoutes(ctx, testRouteMCRUID)
	suite.Error(err)
	suite.NotErrorIs(err, ErrMCRDiagnosticsOperationEmpty)
}

// TestPingMCR tests the PingMCR method happy path.
func (suite *MCRLookingGlassClientTestSuite) TestPingMCR() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
		"message": "Ping operation submitted",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": "op-id-ping-123"
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/ping", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal("8.8.8.8", r.URL.Query().Get("destination_address"))
		fmt.Fprint(w, jblob)
	})

	operationID, err := lgSvc.PingMCR(ctx, &MCRPingRequest{
		MCRID:              mcrUID,
		DestinationAddress: "8.8.8.8",
	})
	suite.NoError(err)
	suite.Equal("op-id-ping-123", operationID)
}

// TestTracerouteMCR tests the TracerouteMCR method happy path.
func (suite *MCRLookingGlassClientTestSuite) TestTracerouteMCR() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
		"message": "Traceroute operation submitted",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": "op-id-traceroute-456"
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/traceroute", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal("1.1.1.1", r.URL.Query().Get("destination_address"))
		fmt.Fprint(w, jblob)
	})

	operationID, err := lgSvc.TracerouteMCR(ctx, &MCRTracerouteRequest{
		MCRID:              mcrUID,
		DestinationAddress: "1.1.1.1",
	})
	suite.NoError(err)
	suite.Equal("op-id-traceroute-456", operationID)
}

// TestGetMCRPingResult tests the GetMCRPingResult method with a complete result.
func (suite *MCRLookingGlassClientTestSuite) TestGetMCRPingResult() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	operationID := "op-id-ping-123"

	jblob := `{
		"message": "Operation complete",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"rawOutput": "PING 8.8.8.8: 56 data bytes",
			"statistics": {
				"duplicates": 0,
				"errors": 0,
				"packetLossPct": 0.0,
				"packetsReceived": 3,
				"packetsTransmitted": 3,
				"rttAvgMs": 1.5,
				"rttMaxMs": 2.0,
				"rttMdevMs": 0.2,
				"rttMinMs": 1.2,
				"totalTimeMs": 5.0
			}
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal(operationID, r.URL.Query().Get("operationId"))
		fmt.Fprint(w, jblob)
	})

	result, err := lgSvc.GetMCRPingResult(ctx, mcrUID, operationID)
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("PING 8.8.8.8: 56 data bytes", result.RawOutput)
	suite.NotNil(result.Statistics)
	suite.Equal(3, result.Statistics.PacketsReceived)
	suite.Equal(3, result.Statistics.PacketsTransmitted)
	suite.Equal(0.0, result.Statistics.PacketLossPct)
	suite.Equal(1.5, result.Statistics.RTTAvgMs)
}

// TestGetMCRTracerouteResult tests the GetMCRTracerouteResult method with a complete result.
func (suite *MCRLookingGlassClientTestSuite) TestGetMCRTracerouteResult() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	operationID := "op-id-traceroute-456"

	jblob := `{
		"message": "Operation complete",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"rawOutput": "traceroute to 1.1.1.1",
			"hops": [
				{
					"hop": "1",
					"probes": [
						{"host": "192.168.1.1", "rttMs": 0.5},
						{"host": "192.168.1.1", "rttMs": 0.4}
					]
				},
				{
					"hop": "2",
					"probes": [
						{"host": "1.1.1.1", "rttMs": 1.2}
					]
				}
			]
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal(operationID, r.URL.Query().Get("operationId"))
		fmt.Fprint(w, jblob)
	})

	result, err := lgSvc.GetMCRTracerouteResult(ctx, mcrUID, operationID)
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("traceroute to 1.1.1.1", result.RawOutput)
	suite.Len(result.Hops, 2)
	suite.Equal("1", result.Hops[0].Hop)
	suite.Len(result.Hops[0].Probes, 2)
	suite.Equal("192.168.1.1", result.Hops[0].Probes[0].Host)
	suite.Equal(0.5, result.Hops[0].Probes[0].RTTMs)
	suite.Equal("2", result.Hops[1].Hop)
	suite.Equal("1.1.1.1", result.Hops[1].Probes[0].Host)
}

// TestPingMCRValidation tests that PingMCR returns an error when DestinationAddress is empty.
func (suite *MCRLookingGlassClientTestSuite) TestPingMCRValidation() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService

	_, err := lgSvc.PingMCR(ctx, &MCRPingRequest{
		MCRID: "36b3f68e-2f54-4331-bf94-f8984449365f",
	})
	suite.ErrorIs(err, ErrMCRPingDestinationRequired)
}

// TestTracerouteMCRValidation tests that TracerouteMCR returns an error when DestinationAddress is empty.
func (suite *MCRLookingGlassClientTestSuite) TestTracerouteMCRValidation() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService

	_, err := lgSvc.TracerouteMCR(ctx, &MCRTracerouteRequest{
		MCRID: "36b3f68e-2f54-4331-bf94-f8984449365f",
	})
	suite.ErrorIs(err, ErrMCRTracerouteDestinationRequired)
}

// TestPingMCRPacketCountValidation tests out-of-range packet_count values.
func (suite *MCRLookingGlassClientTestSuite) TestPingMCRPacketCountValidation() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	tooLow := int32(0)
	_, err := lgSvc.PingMCR(ctx, &MCRPingRequest{
		MCRID:              mcrUID,
		DestinationAddress: "8.8.8.8",
		PacketCount:        &tooLow,
	})
	suite.ErrorIs(err, ErrMCRPingPacketCountOutOfRange)

	tooHigh := int32(61)
	_, err = lgSvc.PingMCR(ctx, &MCRPingRequest{
		MCRID:              mcrUID,
		DestinationAddress: "8.8.8.8",
		PacketCount:        &tooHigh,
	})
	suite.ErrorIs(err, ErrMCRPingPacketCountOutOfRange)
}

// TestPingMCRPacketSizeValidation tests out-of-range packet_size values.
func (suite *MCRLookingGlassClientTestSuite) TestPingMCRPacketSizeValidation() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	tooLow := int32(0)
	_, err := lgSvc.PingMCR(ctx, &MCRPingRequest{
		MCRID:              mcrUID,
		DestinationAddress: "8.8.8.8",
		PacketSize:         &tooLow,
	})
	suite.ErrorIs(err, ErrMCRPingPacketSizeOutOfRange)

	tooHigh := int32(9187)
	_, err = lgSvc.PingMCR(ctx, &MCRPingRequest{
		MCRID:              mcrUID,
		DestinationAddress: "8.8.8.8",
		PacketSize:         &tooHigh,
	})
	suite.ErrorIs(err, ErrMCRPingPacketSizeOutOfRange)
}

// TestWaitForMCRPingSuccess tests WaitForMCRPing when the result is ready on the first poll.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRPingSuccess() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	operationID := "op-id-ping-123"

	jblob := `{
		"message": "Operation complete",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"rawOutput": "PING 8.8.8.8: 56 data bytes",
			"statistics": {
				"duplicates": 0,
				"errors": 0,
				"packetLossPct": 0.0,
				"packetsReceived": 3,
				"packetsTransmitted": 3,
				"rttAvgMs": 1.5,
				"rttMaxMs": 2.0,
				"rttMdevMs": 0.2,
				"rttMinMs": 1.2,
				"totalTimeMs": 5.0
			}
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal(operationID, r.URL.Query().Get("operationId"))
		fmt.Fprint(w, jblob)
	})

	// Use a timeout context to bound the total wait time and prevent the test from hanging.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := lgSvc.WaitForMCRPing(ctx, mcrUID, operationID)
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("PING 8.8.8.8: 56 data bytes", result.RawOutput)
	suite.Equal(3, result.Statistics.PacketsReceived)
}

// TestWaitForMCRTracerouteSuccess tests WaitForMCRTraceroute when the result is ready on the first poll.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRTracerouteSuccess() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	operationID := "op-id-traceroute-456"

	jblob := `{
		"message": "Operation complete",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"rawOutput": "traceroute to 1.1.1.1",
			"hops": [
				{
					"hop": "1",
					"probes": [{"host": "192.168.1.1", "rttMs": 0.5}]
				}
			]
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal(operationID, r.URL.Query().Get("operationId"))
		fmt.Fprint(w, jblob)
	})

	// Use a timeout context to bound the total wait time and prevent the test from hanging.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := lgSvc.WaitForMCRTraceroute(ctx, mcrUID, operationID)
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("traceroute to 1.1.1.1", result.RawOutput)
	suite.Len(result.Hops, 1)
}

// TestWaitForMCRPingValidation tests that WaitForMCRPing returns sentinel errors for invalid inputs.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRPingValidation() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService

	_, err := lgSvc.WaitForMCRPing(ctx, "", "op-id")
	suite.ErrorIs(err, ErrMCRDiagnosticsMCRUIDRequired)

	_, err = lgSvc.WaitForMCRPing(ctx, "36b3f68e-2f54-4331-bf94-f8984449365f", "")
	suite.ErrorIs(err, ErrMCRDiagnosticsOperationEmpty)
}

// TestWaitForMCRTracerouteValidation tests that WaitForMCRTraceroute returns sentinel errors for invalid inputs.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRTracerouteValidation() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService

	_, err := lgSvc.WaitForMCRTraceroute(ctx, "", "op-id")
	suite.ErrorIs(err, ErrMCRDiagnosticsMCRUIDRequired)

	_, err = lgSvc.WaitForMCRTraceroute(ctx, "36b3f68e-2f54-4331-bf94-f8984449365f", "")
	suite.ErrorIs(err, ErrMCRDiagnosticsOperationEmpty)
}

// TestPingMCREmptyMCRID tests that PingMCR returns ErrMCRDiagnosticsMCRUIDRequired when MCRID is empty.
func (suite *MCRLookingGlassClientTestSuite) TestPingMCREmptyMCRID() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService

	_, err := lgSvc.PingMCR(ctx, &MCRPingRequest{
		DestinationAddress: "8.8.8.8",
	})
	suite.ErrorIs(err, ErrMCRDiagnosticsMCRUIDRequired)
}

// TestTracerouteMCREmptyMCRID tests that TracerouteMCR returns ErrMCRDiagnosticsMCRUIDRequired when MCRID is empty.
func (suite *MCRLookingGlassClientTestSuite) TestTracerouteMCREmptyMCRID() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService

	_, err := lgSvc.TracerouteMCR(ctx, &MCRTracerouteRequest{
		DestinationAddress: "1.1.1.1",
	})
	suite.ErrorIs(err, ErrMCRDiagnosticsMCRUIDRequired)
}

// TestPingMCREmptyOperationID tests that PingMCR returns ErrMCRDiagnosticsOperationEmpty when the API returns an empty operation ID.
func (suite *MCRLookingGlassClientTestSuite) TestPingMCREmptyOperationID() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/ping", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, `{"message":"ok","terms":"","data":""}`)
	})

	_, err := lgSvc.PingMCR(ctx, &MCRPingRequest{
		MCRID:              mcrUID,
		DestinationAddress: "8.8.8.8",
	})
	suite.ErrorIs(err, ErrMCRDiagnosticsOperationEmpty)
}

// TestTracerouteMCREmptyOperationID tests that TracerouteMCR returns ErrMCRDiagnosticsOperationEmpty when the API returns an empty operation ID.
func (suite *MCRLookingGlassClientTestSuite) TestTracerouteMCREmptyOperationID() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/traceroute", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, `{"message":"ok","terms":"","data":""}`)
	})

	_, err := lgSvc.TracerouteMCR(ctx, &MCRTracerouteRequest{
		MCRID:              mcrUID,
		DestinationAddress: "1.1.1.1",
	})
	suite.ErrorIs(err, ErrMCRDiagnosticsOperationEmpty)
}

// TestWaitForMCRPingPending tests WaitForMCRPing when the first poll returns pending (nil) and the second returns the result.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRPingPending() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	operationID := "op-id-ping-pending"

	pendingBlob := `{"message":"pending","terms":"","data":null}`
	doneBlob := `{
		"message": "Operation complete",
		"terms": "",
		"data": {
			"rawOutput": "PING 8.8.8.8: 56 data bytes",
			"statistics": {
				"duplicates": 0, "errors": 0, "packetLossPct": 0.0,
				"packetsReceived": 3, "packetsTransmitted": 3,
				"rttAvgMs": 1.5, "rttMaxMs": 2.0, "rttMdevMs": 0.2,
				"rttMinMs": 1.2, "totalTimeMs": 5.0
			}
		}
	}`

	var calls atomic.Int32
	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal(operationID, r.URL.Query().Get("operationId"))
		calls.Add(1)
		if calls.Load() == 1 {
			fmt.Fprint(w, pendingBlob)
		} else {
			fmt.Fprint(w, doneBlob)
		}
	})

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result, err := lgSvc.WaitForMCRPing(ctx, mcrUID, operationID)
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("PING 8.8.8.8: 56 data bytes", result.RawOutput)
	suite.GreaterOrEqual(calls.Load(), int32(2))
}

// TestWaitForMCRTraceroutePending tests WaitForMCRTraceroute when the first poll returns pending and the second returns the result.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRTraceroutePending() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	operationID := "op-id-traceroute-pending"

	pendingBlob := `{"message":"pending","terms":"","data":null}`
	doneBlob := `{
		"message": "Operation complete",
		"terms": "",
		"data": {
			"rawOutput": "traceroute to 1.1.1.1",
			"hops": [{"hop": "1", "probes": [{"host": "192.168.1.1", "rttMs": 0.5}]}]
		}
	}`

	var calls atomic.Int32
	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal(operationID, r.URL.Query().Get("operationId"))
		calls.Add(1)
		if calls.Load() == 1 {
			fmt.Fprint(w, pendingBlob)
		} else {
			fmt.Fprint(w, doneBlob)
		}
	})

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result, err := lgSvc.WaitForMCRTraceroute(ctx, mcrUID, operationID)
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("traceroute to 1.1.1.1", result.RawOutput)
	suite.GreaterOrEqual(calls.Load(), int32(2))
}

// TestWaitForMCRPingContextCancellation tests that WaitForMCRPing respects context cancellation.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRPingContextCancellation() {
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := lgSvc.WaitForMCRPing(ctx, mcrUID, "op-id")
	suite.Error(err)
}

// TestWaitForMCRTracerouteContextCancellation tests that WaitForMCRTraceroute respects context cancellation.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRTracerouteContextCancellation() {
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := lgSvc.WaitForMCRTraceroute(ctx, mcrUID, "op-id")
	suite.Error(err)
}

// TestWaitForMCRPingTimeoutDuringRequest verifies that when the SDK-managed
// timeout fires while a poll request is in flight, the wrapped context error
// from the HTTP client is mapped to ErrMCRDiagnosticsTimeout.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRPingTimeoutDuringRequest() {
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	operationID := "op-id-ping-timeout"

	op, ok := lgSvc.(*MCRLookingGlassServiceOp)
	suite.Require().True(ok)
	op.pollTimeout = 20 * time.Millisecond

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // block until the SDK timeout cancels the request
	})

	// No caller deadline, so the SDK-managed timeout applies.
	_, err := lgSvc.WaitForMCRPing(context.Background(), mcrUID, operationID)
	suite.ErrorIs(err, ErrMCRDiagnosticsTimeout)
}

// TestWaitForMCRTracerouteTimeoutDuringRequest verifies the same in-flight
// timeout mapping for WaitForMCRTraceroute.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRTracerouteTimeoutDuringRequest() {
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	operationID := "op-id-traceroute-timeout"

	op, ok := lgSvc.(*MCRLookingGlassServiceOp)
	suite.Require().True(ok)
	op.pollTimeout = 20 * time.Millisecond

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // block until the SDK timeout cancels the request
	})

	// No caller deadline, so the SDK-managed timeout applies.
	_, err := lgSvc.WaitForMCRTraceroute(context.Background(), mcrUID, operationID)
	suite.ErrorIs(err, ErrMCRDiagnosticsTimeout)
}

// TestWaitForMCRPingCallerDeadlineDuringRequest verifies that a caller-provided
// deadline firing mid-request surfaces as the caller's context error, not the
// SDK timeout sentinel.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForMCRPingCallerDeadlineDuringRequest() {
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	operationID := "op-id-ping-caller-deadline"

	path := fmt.Sprintf("/v2/product/mcr2/%s/diagnostics/routes/operation", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // block until the caller deadline cancels the request
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := lgSvc.WaitForMCRPing(ctx, mcrUID, operationID)
	suite.ErrorIs(err, context.DeadlineExceeded)
	suite.False(errors.Is(err, ErrMCRDiagnosticsTimeout))
}

// Go's JSON decoder matches keys case-insensitively, so a decode test stays
// green when a tag's case drifts from the API schema. Marshaling catches it.
// best, external, and valid carry distinct values here, so a tag swapped
// between them fails rather than round-tripping.
func (suite *MCRLookingGlassClientTestSuite) TestRouteJSONTagCase() {
	ipJSON, err := json.Marshal(&LookingGlassIPRoute{
		Prefix:   "10.0.1.0/24",
		Protocol: "bgp",
		Distance: 20,
		Metric:   100,
		NextHop: LookingGlassRouteNextHop{
			IP:  "10.0.0.1",
			VXC: LookingGlassRouteVXCRef{ID: "vxc-uid", Name: "vxc-name"},
		},
	})
	suite.Require().NoError(err)
	for _, want := range []string{
		`"prefix":`, `"protocol":`, `"distance":`, `"metric":`, `"nextHop":`,
		`"ip":`, `"vxc":`, `"id":`, `"name":`,
	} {
		suite.Contains(string(ipJSON), want)
	}

	bgpJSON, err := json.Marshal(&LookingGlassBGPRoute{
		Prefix:       "10.0.1.0/24",
		ASPath:       "65000 65001",
		Origin:       "IGP",
		Source:       "external",
		Since:        "2026-01-01T00:00:00Z",
		Best:         true,
		External:     false,
		Valid:        true,
		Communities:  []string{"65000:100"},
		AdvertisedTo: []string{"10.0.0.2"},
		NextHop:      LookingGlassRouteNextHop{IP: "10.0.0.1"},
	})
	suite.Require().NoError(err)
	for _, want := range []string{
		`"best":true`, `"external":false`, `"valid":true`,
		`"prefix":`, `"asPath":`, `"origin":`, `"source":`, `"localPref":`,
		`"med":`, `"weight":`, `"best":`, `"external":`, `"valid":`,
		`"since":`, `"communities":`, `"advertisedTo":`, `"nextHop":`,
	} {
		suite.Contains(string(bgpJSON), want)
	}
}

// TestRouteEncodingKeepsDocumentedZeros tests that a route re-encodes the zero
// values the spec documents as real. distance is the deliberate exception: the
// spec sets minimum 1, so its zero can only mean absent.
func (suite *MCRLookingGlassClientTestSuite) TestRouteEncodingKeepsDocumentedZeros() {
	ipJSON, err := json.Marshal(&LookingGlassIPRoute{Prefix: "10.0.1.0/24", Protocol: "connected"})
	suite.Require().NoError(err)
	suite.Contains(string(ipJSON), `"metric":0`)
	suite.NotContains(string(ipJSON), "distance")

	bgpJSON, err := json.Marshal(&LookingGlassBGPRoute{Prefix: "10.0.1.0/24", Origin: "IGP"})
	suite.Require().NoError(err)
	for _, want := range []string{`"med":0`, `"localPref":0`, `"weight":0`, `"best":false`, `"external":false`, `"valid":false`} {
		suite.Contains(string(bgpJSON), want)
	}
}

// TestRouteOperationUnexpected200Shapes pins how the poll reads two 200 bodies
// the spec does not describe.
func (suite *MCRLookingGlassClientTestSuite) TestRouteOperationUnexpected200Shapes() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	operationID := "5c6d7e8f-9012-4345-8678-9012345678ab"

	body := ""
	base := "/v2/product/mcr2/" + testRouteMCRUID + "/diagnostics/routes"
	suite.mux.HandleFunc(base+"/ip", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"ok","terms":"","data":%q}`, operationID)
	})
	suite.mux.HandleFunc(base+"/operation", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	})

	// A 200 with no data key. Only a 202 means the result is not ready, so this
	// reads as an empty routing table rather than another poll.
	body = `{"message":"Data result will be available soon","terms":""}`
	routes, err := lgSvc.ListIPRoutes(ctx, testRouteMCRUID)
	suite.NoError(err)
	suite.NotNil(routes)
	suite.Empty(routes)

	// A 200 that is not JSON, which is what a proxy error page looks like.
	body = "<html>gateway error</html>"
	_, err = lgSvc.ListIPRoutes(ctx, testRouteMCRUID)
	suite.Error(err)
	suite.NotErrorIs(err, ErrMCRDiagnosticsTimeout)
}

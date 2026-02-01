package megaport

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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
}

func (suite *MCRLookingGlassClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// TestListIPRoutes tests the ListIPRoutes method.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutes() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	want := []*LookingGlassIPRoute{
		{
			Prefix:    "10.0.0.0/24",
			NextHop:   "192.168.1.1",
			Protocol:  RouteProtocolBGP,
			Metric:    PtrTo(100),
			LocalPref: PtrTo(100),
			ASPath:    []int{65001, 65002},
			Interface: "vxc-1234",
			VXCId:     PtrTo(1234),
			VXCName:   "Test VXC",
			Best:      PtrTo(true),
		},
		{
			Prefix:    "172.16.0.0/16",
			NextHop:   "0.0.0.0",
			Protocol:  RouteProtocolConnected,
			Interface: "eth0",
		},
	}

	jblob := `{
		"message": "Routes retrieved successfully",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"prefix": "10.0.0.0/24",
				"nextHop": "192.168.1.1",
				"protocol": "BGP",
				"metric": 100,
				"localPref": 100,
				"asPath": [65001, 65002],
				"interface": "vxc-1234",
				"vxcId": 1234,
				"vxcName": "Test VXC",
				"best": true
			},
			{
				"prefix": "172.16.0.0/16",
				"nextHop": "0.0.0.0",
				"protocol": "CONNECTED",
				"interface": "eth0"
			}
		]
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestListIPRoutesWithFilter tests the ListIPRoutesWithFilter method with protocol filter.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesWithFilter() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	want := []*LookingGlassIPRoute{
		{
			Prefix:    "10.0.0.0/24",
			NextHop:   "192.168.1.1",
			Protocol:  RouteProtocolBGP,
			LocalPref: PtrTo(100),
			ASPath:    []int{65001},
			Best:      PtrTo(true),
		},
	}

	jblob := `{
		"message": "Routes retrieved successfully",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"prefix": "10.0.0.0/24",
				"nextHop": "192.168.1.1",
				"protocol": "BGP",
				"localPref": 100,
				"asPath": [65001],
				"best": true
			}
		]
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal("BGP", r.URL.Query().Get("protocol"))
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.ListIPRoutesWithFilter(ctx, &ListIPRoutesRequest{
		MCRID:    mcrUID,
		Protocol: RouteProtocolBGP,
	})
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestListBGPRoutes tests the ListBGPRoutes method.
func (suite *MCRLookingGlassClientTestSuite) TestListBGPRoutes() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	want := []*LookingGlassBGPRoute{
		{
			Prefix:      "10.0.0.0/24",
			NextHop:     "192.168.1.1",
			ASPath:      []int{65001, 65002, 65003},
			LocalPref:   PtrTo(100),
			MED:         PtrTo(50),
			Origin:      "IGP",
			Communities: []string{"65001:100", "65001:200"},
			Valid:       true,
			Best:        true,
			NeighborIP:  "192.168.1.1",
			NeighborASN: PtrTo(65001),
			VXCId:       PtrTo(1234),
			VXCName:     "Test VXC",
		},
	}

	jblob := `{
		"message": "BGP routes retrieved successfully",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"prefix": "10.0.0.0/24",
				"nextHop": "192.168.1.1",
				"asPath": [65001, 65002, 65003],
				"localPref": 100,
				"med": 50,
				"origin": "IGP",
				"communities": ["65001:100", "65001:200"],
				"valid": true,
				"best": true,
				"neighborIp": "192.168.1.1",
				"neighborAsn": 65001,
				"vxcId": 1234,
				"vxcName": "Test VXC"
			}
		]
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgp", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.ListBGPRoutes(ctx, mcrUID)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestListBGPSessions tests the ListBGPSessions method.
func (suite *MCRLookingGlassClientTestSuite) TestListBGPSessions() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	want := []*LookingGlassBGPSession{
		{
			SessionID:       "session-1",
			NeighborAddress: "192.168.1.1",
			NeighborASN:     65001,
			LocalASN:        65000,
			Status:          BGPSessionStatusUp,
			Uptime:          PtrTo(86400),
			PrefixesIn:      PtrTo(100),
			PrefixesOut:     PtrTo(50),
			VXCId:           1234,
			VXCName:         "Test VXC",
			Description:     "AWS Direct Connect",
		},
		{
			SessionID:       "session-2",
			NeighborAddress: "192.168.2.1",
			NeighborASN:     65002,
			LocalASN:        65000,
			Status:          BGPSessionStatusDown,
			VXCId:           5678,
			VXCName:         "Test VXC 2",
			Description:     "Azure ExpressRoute",
		},
	}

	jblob := `{
		"message": "BGP sessions retrieved successfully",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"sessionId": "session-1",
				"neighborAddress": "192.168.1.1",
				"neighborAsn": 65001,
				"localAsn": 65000,
				"status": "UP",
				"uptime": 86400,
				"prefixesIn": 100,
				"prefixesOut": 50,
				"vxcId": 1234,
				"vxcName": "Test VXC",
				"description": "AWS Direct Connect"
			},
			{
				"sessionId": "session-2",
				"neighborAddress": "192.168.2.1",
				"neighborAsn": 65002,
				"localAsn": 65000,
				"status": "DOWN",
				"vxcId": 5678,
				"vxcName": "Test VXC 2",
				"description": "Azure ExpressRoute"
			}
		]
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.ListBGPSessions(ctx, mcrUID)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestListBGPNeighborRoutes tests the ListBGPNeighborRoutes method.
func (suite *MCRLookingGlassClientTestSuite) TestListBGPNeighborRoutes() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	sessionID := "session-1"

	want := []*LookingGlassBGPNeighborRoute{
		{
			Prefix:      "10.0.0.0/24",
			NextHop:     "192.168.1.1",
			ASPath:      []int{65001},
			LocalPref:   PtrTo(100),
			MED:         PtrTo(0),
			Origin:      "IGP",
			Communities: []string{"65001:100"},
			Valid:       true,
			Best:        true,
		},
		{
			Prefix:    "10.0.1.0/24",
			NextHop:   "192.168.1.1",
			ASPath:    []int{65001, 65002},
			LocalPref: PtrTo(100),
			Origin:    "EGP",
			Valid:     true,
			Best:      false,
		},
	}

	jblob := `{
		"message": "BGP neighbor routes retrieved successfully",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"prefix": "10.0.0.0/24",
				"nextHop": "192.168.1.1",
				"asPath": [65001],
				"localPref": 100,
				"med": 0,
				"origin": "IGP",
				"communities": ["65001:100"],
				"valid": true,
				"best": true
			},
			{
				"prefix": "10.0.1.0/24",
				"nextHop": "192.168.1.1",
				"asPath": [65001, 65002],
				"localPref": 100,
				"origin": "EGP",
				"valid": true,
				"best": false
			}
		]
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/%s/received", mcrUID, sessionID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.ListBGPNeighborRoutes(ctx, &ListBGPNeighborRoutesRequest{
		MCRID:     mcrUID,
		SessionID: sessionID,
		Direction: LookingGlassRouteDirectionReceived,
	})
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestListBGPNeighborRoutesAdvertised tests the ListBGPNeighborRoutes method for advertised routes.
func (suite *MCRLookingGlassClientTestSuite) TestListBGPNeighborRoutesAdvertised() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	sessionID := "session-1"

	want := []*LookingGlassBGPNeighborRoute{
		{
			Prefix:  "172.16.0.0/24",
			NextHop: "0.0.0.0",
			ASPath:  []int{65000},
			Origin:  "IGP",
			Valid:   true,
			Best:    true,
		},
	}

	jblob := `{
		"message": "BGP neighbor routes retrieved successfully",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"prefix": "172.16.0.0/24",
				"nextHop": "0.0.0.0",
				"asPath": [65000],
				"origin": "IGP",
				"valid": true,
				"best": true
			}
		]
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/%s/advertised", mcrUID, sessionID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.ListBGPNeighborRoutes(ctx, &ListBGPNeighborRoutesRequest{
		MCRID:     mcrUID,
		SessionID: sessionID,
		Direction: LookingGlassRouteDirectionAdvertised,
	})
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestListIPRoutesAsync tests the ListIPRoutesAsync method.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesAsync() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	want := &LookingGlassAsyncJob{
		JobID:  "job-12345",
		Status: LookingGlassAsyncStatusPending,
	}

	jblob := `{
		"message": "Async job created",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"jobId": "job-12345",
			"status": "PENDING"
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal("true", r.URL.Query().Get("async"))
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.ListIPRoutesAsync(ctx, mcrUID)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestGetAsyncIPRoutes tests the GetAsyncIPRoutes method.
func (suite *MCRLookingGlassClientTestSuite) TestGetAsyncIPRoutes() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	jobID := "job-12345"

	want := &AsyncIPRoutesData{
		JobID:  jobID,
		Status: LookingGlassAsyncStatusComplete,
		Routes: []*LookingGlassIPRoute{
			{
				Prefix:   "10.0.0.0/24",
				NextHop:  "192.168.1.1",
				Protocol: RouteProtocolBGP,
			},
		},
	}

	jblob := `{
		"message": "Async job complete",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"jobId": "job-12345",
			"status": "COMPLETE",
			"routes": [
				{
					"prefix": "10.0.0.0/24",
					"nextHop": "192.168.1.1",
					"protocol": "BGP"
				}
			]
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes/async/%s", mcrUID, jobID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.GetAsyncIPRoutes(ctx, mcrUID, jobID)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestListBGPNeighborRoutesAsync tests the ListBGPNeighborRoutesAsync method.
func (suite *MCRLookingGlassClientTestSuite) TestListBGPNeighborRoutesAsync() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	sessionID := "session-1"

	want := &LookingGlassAsyncJob{
		JobID:  "job-67890",
		Status: LookingGlassAsyncStatusPending,
	}

	jblob := `{
		"message": "Async job created",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"jobId": "job-67890",
			"status": "PENDING"
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/%s/received", mcrUID, sessionID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal("true", r.URL.Query().Get("async"))
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.ListBGPNeighborRoutesAsync(ctx, &ListBGPNeighborRoutesRequest{
		MCRID:     mcrUID,
		SessionID: sessionID,
		Direction: LookingGlassRouteDirectionReceived,
	})
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestGetAsyncBGPNeighborRoutes tests the GetAsyncBGPNeighborRoutes method.
func (suite *MCRLookingGlassClientTestSuite) TestGetAsyncBGPNeighborRoutes() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	jobID := "job-67890"

	want := &AsyncBGPNeighborRoutesData{
		JobID:  jobID,
		Status: LookingGlassAsyncStatusComplete,
		Routes: []*LookingGlassBGPNeighborRoute{
			{
				Prefix:  "10.0.0.0/24",
				NextHop: "192.168.1.1",
				ASPath:  []int{65001},
				Valid:   true,
				Best:    true,
			},
		},
	}

	jblob := `{
		"message": "Async job complete",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"jobId": "job-67890",
			"status": "COMPLETE",
			"routes": [
				{
					"prefix": "10.0.0.0/24",
					"nextHop": "192.168.1.1",
					"asPath": [65001],
					"valid": true,
					"best": true
				}
			]
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/async/%s", mcrUID, jobID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.GetAsyncBGPNeighborRoutes(ctx, mcrUID, jobID)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestListIPRoutesError tests error handling for ListIPRoutes.
func (suite *MCRLookingGlassClientTestSuite) TestListIPRoutesError() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "invalid-mcr-uid"

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message": "MCR not found", "data": ""}`)
	})

	_, err := lgSvc.ListIPRoutes(ctx, mcrUID)
	suite.Error(err)
}

// TestListBGPSessionsEmpty tests ListBGPSessions with empty results.
func (suite *MCRLookingGlassClientTestSuite) TestListBGPSessionsEmpty() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	jblob := `{
		"message": "No BGP sessions found",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": []
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.ListBGPSessions(ctx, mcrUID)
	suite.NoError(err)
	suite.Empty(got)
}

// TestListBGPRoutesWithFilter tests the ListBGPRoutesWithFilter method with IP filter.
func (suite *MCRLookingGlassClientTestSuite) TestListBGPRoutesWithFilter() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"

	want := []*LookingGlassBGPRoute{
		{
			Prefix:     "10.0.0.0/24",
			NextHop:    "192.168.1.1",
			ASPath:     []int{65001},
			LocalPref:  PtrTo(100),
			Origin:     "IGP",
			Valid:      true,
			Best:       true,
			NeighborIP: "192.168.1.1",
		},
	}

	jblob := `{
		"message": "BGP routes retrieved successfully",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"prefix": "10.0.0.0/24",
				"nextHop": "192.168.1.1",
				"asPath": [65001],
				"localPref": 100,
				"origin": "IGP",
				"valid": true,
				"best": true,
				"neighborIp": "192.168.1.1"
			}
		]
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgp", mcrUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal("10.0.0.0/24", r.URL.Query().Get("ip"))
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.ListBGPRoutesWithFilter(ctx, &ListBGPRoutesRequest{
		MCRID:    mcrUID,
		IPFilter: "10.0.0.0/24",
	})
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestWaitForAsyncIPRoutesComplete tests WaitForAsyncIPRoutes when job completes immediately.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForAsyncIPRoutesComplete() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	jobID := "job-12345"

	want := []*LookingGlassIPRoute{
		{
			Prefix:   "10.0.0.0/24",
			NextHop:  "192.168.1.1",
			Protocol: RouteProtocolBGP,
		},
	}

	jblob := `{
		"message": "Async job complete",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"jobId": "job-12345",
			"status": "COMPLETE",
			"routes": [
				{
					"prefix": "10.0.0.0/24",
					"nextHop": "192.168.1.1",
					"protocol": "BGP"
				}
			]
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes/async/%s", mcrUID, jobID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.WaitForAsyncIPRoutes(ctx, mcrUID, jobID, 10*time.Second)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestWaitForAsyncIPRoutesFailed tests WaitForAsyncIPRoutes when job fails.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForAsyncIPRoutesFailed() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	jobID := "job-12345"

	jblob := `{
		"message": "Async job failed",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"jobId": "job-12345",
			"status": "FAILED"
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/routes/async/%s", mcrUID, jobID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	_, err := lgSvc.WaitForAsyncIPRoutes(ctx, mcrUID, jobID, 10*time.Second)
	suite.Error(err)
	suite.Contains(err.Error(), "failed")
}

// TestWaitForAsyncBGPNeighborRoutesComplete tests WaitForAsyncBGPNeighborRoutes when job completes.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForAsyncBGPNeighborRoutesComplete() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	jobID := "job-67890"

	want := []*LookingGlassBGPNeighborRoute{
		{
			Prefix:  "10.0.0.0/24",
			NextHop: "192.168.1.1",
			ASPath:  []int{65001},
			Valid:   true,
			Best:    true,
		},
	}

	jblob := `{
		"message": "Async job complete",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"jobId": "job-67890",
			"status": "COMPLETE",
			"routes": [
				{
					"prefix": "10.0.0.0/24",
					"nextHop": "192.168.1.1",
					"asPath": [65001],
					"valid": true,
					"best": true
				}
			]
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/async/%s", mcrUID, jobID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	got, err := lgSvc.WaitForAsyncBGPNeighborRoutes(ctx, mcrUID, jobID, 10*time.Second)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestWaitForAsyncBGPNeighborRoutesFailed tests WaitForAsyncBGPNeighborRoutes when job fails.
func (suite *MCRLookingGlassClientTestSuite) TestWaitForAsyncBGPNeighborRoutesFailed() {
	ctx := context.Background()
	lgSvc := suite.client.MCRLookingGlassService
	mcrUID := "36b3f68e-2f54-4331-bf94-f8984449365f"
	jobID := "job-67890"

	jblob := `{
		"message": "Async job failed",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"jobId": "job-67890",
			"status": "FAILED"
		}
	}`

	path := fmt.Sprintf("/v2/product/mcr2/%s/lookingGlass/bgpSessions/async/%s", mcrUID, jobID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	_, err := lgSvc.WaitForAsyncBGPNeighborRoutes(ctx, mcrUID, jobID, 10*time.Second)
	suite.Error(err)
	suite.Contains(err.Error(), "failed")
}

package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Global time variables used throughout the test suite
var (
	startDate = &Time{GetTime(1706104800000)} // Jan 24, 2024
	endDate   = &Time{GetTime(1737727200000)} // May 24, 2025
	liveDate  = &Time{GetTime(1737728200000)} // May 24, 2025 + 16 minutes
)

// IXClientTestSuite tests the IX service.
type IXClientTestSuite struct {
	ClientTestSuite
}

func TestIXClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(IXClientTestSuite))
}

func (suite *IXClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *IXClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// TestBuyIX tests the BuyIX method.
func (suite *IXClientTestSuite) TestBuyIX() {
	portProductUid := "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	ixProductUid := "4017d2c3-1537-408e-9f0a-fdc72b6cce39"
	req := &BuyIXRequest{
		ProductUID:         portProductUid,
		Name:               "My LA IX",
		NetworkServiceType: "Los Angeles IX",
		ASN:                12345,
		MACAddress:         "00:11:22:33:44:55",
		RateLimit:          500,
		VLAN:               2001,
		Shutdown:           false,
		PromoCode:          "promox3mnthfree2",
	}

	// Expected order format
	ixOrder := IXOrder{
		ProductUID: portProductUid,
		AssociatedIXs: []AssociatedIXOrder{
			{
				ProductName:        "My LA IX",
				NetworkServiceType: "Los Angeles IX",
				ASN:                12345,
				MACAddress:         "00:11:22:33:44:55",
				RateLimit:          500,
				VLAN:               2001,
				Shutdown:           false,
				PromoCode:          "promox3mnthfree2",
			},
		},
	}

	ixSvc := suite.client.IXService
	ctx := context.Background()

	// Mock the validation endpoint
	validatePath := "/v3/networkdesign/validate"
	suite.mux.HandleFunc(validatePath, func(w http.ResponseWriter, r *http.Request) {
		v := new([]IXOrder)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		orders := *v
		wantOrder := ixOrder
		gotOrder := orders[0]
		suite.testMethod(r, http.MethodPost)
		suite.Equal(wantOrder.ProductUID, gotOrder.ProductUID)
		suite.Equal(wantOrder.AssociatedIXs[0].ProductName, gotOrder.AssociatedIXs[0].ProductName)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "[]") // Respond with an empty JSON array
	})

	// Mock the buy endpoint
	buyPath := "/v3/networkdesign/buy"
	jblob := fmt.Sprintf(`{
        "message": "IX [%s] created.",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": [
            {
                "technicalServiceUid": "%s"
            }
        ]
    }`, ixProductUid, ixProductUid)
	suite.mux.HandleFunc(buyPath, func(w http.ResponseWriter, r *http.Request) {
		v := new([]IXOrder)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		orders := *v
		wantOrder := ixOrder
		gotOrder := orders[0]
		suite.testMethod(r, http.MethodPost)
		suite.Equal(wantOrder.ProductUID, gotOrder.ProductUID)
		suite.Equal(wantOrder.AssociatedIXs[0].ProductName, gotOrder.AssociatedIXs[0].ProductName)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	want := &BuyIXResponse{
		TechnicalServiceUID: ixProductUid,
	}
	got, err := ixSvc.BuyIX(ctx, req)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestGetIX tests the GetIX method.
func (suite *IXClientTestSuite) TestGetIX() {
	ctx := context.Background()
	ixSvc := suite.client.IXService

	ixUid := "4017d2c3-1537-408e-9f0a-fdc72b6cce39"

	wantIX := &IX{
		ProductID:  235400,
		ProductUID: ixUid,
		LocationID: 67,
		LocationDetail: IXLocationDetail{
			Name:    "Equinix DC4",
			City:    "Ashburn",
			Metro:   "Ashburn",
			Country: "USA",
		},
		Term:               1,
		LocationUID:        "",
		ProductName:        "My LA IX",
		ProvisioningStatus: "CONFIGURED",
		RateLimit:          500,
		VLAN:               2001,
		MACAddress:         "00:11:22:33:44:55",
		ASN:                12345,
		NetworkServiceType: "Los Angeles IX",
		CreateDate:         startDate,
		DeployDate:         liveDate,
		PublicGraph:        true,
		UsageAlgorithm:     "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
		Resources: IXResources{
			Interface: IXInterface{
				Demarcation:  "Equinix DC4\\n21691 Filigree Court\\nDC4:01:041260:0000\\nType: Single-mode Fibre Pair\\nPorts: PP:0000:1231967, ports 25 Tx+25 Rx\\nTermination: LC connector\\n",
				LOATemplate:  "megaport",
				Media:        "LR",
				PortSpeed:    10000,
				ResourceName: "interface",
				ResourceType: "interface",
				Up:           1,
				Shutdown:     false,
			},
			BGPConnections: []IXBGPConnection{
				{
					ASN:               12345,
					CustomerASN:       12345,
					CustomerIPAddress: "206.53.172.30/24",
					ISPASN:            64220,
					ISPIPAddress:      "206.53.172.1",
					IXPeerPolicy:      "open",
					MaxPrefixes:       1000,
					ResourceName:      "rs1_ipv4_bgp_connection",
					ResourceType:      "bgp_connection",
				},
			},
			IPAddresses: []IXIPAddress{
				{
					Address:      "206.53.172.30/24",
					ResourceName: "ipv4_address",
					ResourceType: "ip_address",
					Version:      4,
					ReverseDNS:   "as12345.los-angeles.megaport.com.",
				},
			},
			VPLSInterface: IXVPLSInterface{
				MACAddress:    "00:11:22:33:44:55",
				RateLimitMbps: 500,
				ResourceName:  "vpls_interface",
				ResourceType:  "vpls_interface",
				VLAN:          2001,
				Shutdown:      false,
			},
		},
	}

	createDateMillis := startDate.UnixMilli()
	deployDateMillis := liveDate.UnixMilli()

	jblob := fmt.Sprintf(`{
        "message": "Found Product %s",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": {
            "productId": 235400,
            "productUid": "%s",
            "locationId": 67,
            "locationDetail": {
                "name": "Equinix DC4",
                "city": "Ashburn",
                "metro": "Ashburn",
                "country": "USA"
            },
            "term": 1,
            "locationUid": "",
            "productName": "My LA IX",
            "provisioningStatus": "CONFIGURED",
            "rateLimit": 500,
            "promoCode": null,
            "deployDate": %d,
            "secondaryName": "",
            "attributeTags": {},
            "vlan": 2001,
            "macAddress": "00:11:22:33:44:55",
            "ixPeerMacro": null,
            "asn": 12345,
            "networkServiceType": "Los Angeles IX",
            "createDate": %d,
            "publicGraph": true,
            "usageAlgorithm": "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
            "resources": {
                "interface": {
                    "demarcation": "Equinix DC4\\n21691 Filigree Court\\nDC4:01:041260:0000\\nType: Single-mode Fibre Pair\\nPorts: PP:0000:1231967, ports 25 Tx+25 Rx\\nTermination: LC connector\\n",
                    "loa_template": "megaport",
                    "media": "LR",
                    "port_speed": 10000,
                    "resource_name": "interface",
                    "resource_type": "interface",
                    "up": 1,
                    "shutdown": false
                },
                "bgp_connection": [
                    {
                        "asn": 12345,
                        "customer_asn": 12345,
                        "customer_ip_address": "206.53.172.30/24",
                        "isp_asn": 64220,
                        "isp_ip_address": "206.53.172.1",
                        "ix_peer_policy": "open",
                        "max_prefixes": 1000,
                        "resource_name": "rs1_ipv4_bgp_connection",
                        "resource_type": "bgp_connection"
                    }
                ],
                "ip_address": [
                    {
                        "address": "206.53.172.30/24",
                        "resource_name": "ipv4_address",
                        "resource_type": "ip_address",
                        "version": 4,
                        "reverse_dns": "as12345.los-angeles.megaport.com."
                    }
                ],
                "vpls_interface": {
                    "mac_address": "00:11:22:33:44:55",
                    "rate_limit_mbps": 500,
                    "resource_name": "vpls_interface",
                    "resource_type": "vpls_interface",
                    "vlan": 2001,
                    "shutdown": false
                }
            }
        }
    }`, ixUid, ixUid, deployDateMillis, createDateMillis)

	path := "/v2/product/" + ixUid
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	gotIX, err := ixSvc.GetIX(ctx, ixUid)
	suite.NoError(err)
	suite.Equal(wantIX.ProductUID, gotIX.ProductUID)
	suite.Equal(wantIX.CreateDate, gotIX.CreateDate)
	suite.Equal(wantIX.DeployDate, gotIX.DeployDate)
	suite.Equal(wantIX.LocationDetail, gotIX.LocationDetail)
	suite.Equal(wantIX.LocationID, gotIX.LocationID)
	suite.Equal(wantIX.LocationUID, gotIX.LocationUID)
	suite.Equal(wantIX.MACAddress, gotIX.MACAddress)
	suite.Equal(wantIX.NetworkServiceType, gotIX.NetworkServiceType)
	suite.Equal(wantIX.ProductID, gotIX.ProductID)
	suite.Equal(wantIX.ProductName, gotIX.ProductName)
	suite.Equal(wantIX.ProvisioningStatus, gotIX.ProvisioningStatus)
	suite.Equal(wantIX.PublicGraph, gotIX.PublicGraph)
	suite.Equal(wantIX.RateLimit, gotIX.RateLimit)
	suite.Equal(wantIX.Term, gotIX.Term)
	suite.Equal(wantIX.UsageAlgorithm, gotIX.UsageAlgorithm)
	suite.Equal(wantIX.VLAN, gotIX.VLAN)
	suite.Equal(wantIX.ASN, gotIX.ASN)
	suite.Equal(wantIX.Resources, gotIX.Resources)
}

// TestUpdateIX tests the UpdateIX method.
func (suite *IXClientTestSuite) TestUpdateIX() {
	ctx := context.Background()
	ixSvc := suite.client.IXService

	ixUid := "4017d2c3-1537-408e-9f0a-fdc72b6cce39"

	updateName := "Updated IX Name"
	newRateLimit := 1000
	newVlan := 2002
	newMacAddress := "00:22:33:44:55:66"

	updateReq := &UpdateIXRequest{
		Name:       &updateName,
		RateLimit:  &newRateLimit,
		VLAN:       &newVlan,
		MACAddress: &newMacAddress,
	}

	jblob := fmt.Sprintf(`{
        "message": "Product [%s] has been updated",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": {
            "productId": 235400,
            "productUid": "%s",
            "locationId": 67,
            "locationDetail": {
                "name": "Equinix DC4",
                "city": "Ashburn",
                "metro": "Ashburn",
                "country": "USA"
            },
            "term": 0,
            "locationUid": "",
            "productName": "Updated IX Name",
            "provisioningStatus": "CONFIGURED",
            "rateLimit": 1000,
            "promoCode": null,
            "deployDate": %d,
            "secondaryName": "",
            "attributeTags": {},
            "vlan": 2002,
            "macAddress": "00:22:33:44:55:66",
            "ixPeerMacro": null,
            "asn": 12345,
            "networkServiceType": "Los Angeles IX",
            "createDate": %d,
            "publicGraph": true,
            "usageAlgorithm": "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC"
        }
    }`, ixUid, ixUid, endDate.UnixMilli(), startDate.UnixMilli())

	update := &IXUpdate{
		Name:       updateName,
		RateLimit:  &newRateLimit,
		VLAN:       &newVlan,
		MACAddress: newMacAddress,
	}

	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_IX, ixUid)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(IXUpdate)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			suite.FailNowf("could not decode json", "could not decode json %v", err)
		}
		suite.testMethod(r, http.MethodPut)
		fmt.Fprint(w, jblob)
		suite.Equal(update.Name, v.Name)
		suite.Equal(update.RateLimit, v.RateLimit)
		suite.Equal(update.VLAN, v.VLAN)
		suite.Equal(update.MACAddress, v.MACAddress)
	})

	wantIX := &IX{
		ProductID:  235400,
		ProductUID: ixUid,
		LocationID: 67,
		LocationDetail: IXLocationDetail{
			Name:    "Equinix DC4",
			City:    "Ashburn",
			Metro:   "Ashburn",
			Country: "USA",
		},
		Term:               0,
		LocationUID:        "",
		ProductName:        updateName,
		ProvisioningStatus: "CONFIGURED",
		RateLimit:          newRateLimit,
		VLAN:               newVlan,
		MACAddress:         newMacAddress,
		ASN:                12345,
		NetworkServiceType: "Los Angeles IX",
		CreateDate:         startDate,
		DeployDate:         endDate,
		PublicGraph:        true,
		UsageAlgorithm:     "POST_PAID_HOURLY_SPEED_LONG_HAUL_VXC",
	}

	gotIX, err := ixSvc.UpdateIX(ctx, ixUid, updateReq)
	suite.NoError(err)
	suite.Equal(wantIX.ProductUID, gotIX.ProductUID)
	suite.Equal(wantIX.ProductName, gotIX.ProductName)
	suite.Equal(wantIX.RateLimit, gotIX.RateLimit)
	suite.Equal(wantIX.VLAN, gotIX.VLAN)
	suite.Equal(wantIX.MACAddress, gotIX.MACAddress)
	suite.Equal(wantIX.DeployDate, gotIX.DeployDate)
	suite.Equal(wantIX.CreateDate, gotIX.CreateDate)
	suite.Equal(wantIX.LocationDetail, gotIX.LocationDetail)
	suite.Equal(wantIX.LocationID, gotIX.LocationID)
	suite.Equal(wantIX.LocationUID, gotIX.LocationUID)
	suite.Equal(wantIX.NetworkServiceType, gotIX.NetworkServiceType)
	suite.Equal(wantIX.ProductID, gotIX.ProductID)
	suite.Equal(wantIX.ProvisioningStatus, gotIX.ProvisioningStatus)
	suite.Equal(wantIX.PublicGraph, gotIX.PublicGraph)
	suite.Equal(wantIX.Term, gotIX.Term)
	suite.Equal(wantIX.UsageAlgorithm, gotIX.UsageAlgorithm)
	suite.Equal(wantIX.Resources, gotIX.Resources)
	suite.Equal(wantIX.ASN, gotIX.ASN)
}

// TestListIXs tests the ListIXs method with various filters
func (suite *IXClientTestSuite) TestListIXs() {
	// Define mock response for products list API with associated IXs
	productsResponse := `{
        "message": "Found 3 Products",
        "terms": "This data is subject to the Acceptable Use Policy",
        "data": [
            {
                "productId": 8001,
                "productUid": "port-test-ix-001",
                "productName": "Test IX Port 1",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "companyUid": "company-test-123",
                "companyName": "Test Company",
                "associatedIxs": [
                    {
                        "productId": 9001,
                        "productUid": "ix-test-123",
                        "productName": "Test IX 1",
                        "productType": "IX",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 1000,
                        "locationId": 42,
                        "vlan": 100,
                        "asn": 65001,
                        "networkServiceType": "Los Angeles IX",
                        "locationDetail": {
                            "name": "Test Location LA",
                            "city": "Los Angeles",
                            "metro": "Los Angeles",
                            "country": "USA"
                        }
                    },
                    {
                        "productId": 9002,
                        "productUid": "ix-test-456",
                        "productName": "Test EXAMPLE IX",
                        "productType": "IX",
                        "provisioningStatus": "CONFIGURED",
                        "rateLimit": 500,
                        "locationId": 42,
                        "vlan": 200,
                        "asn": 65002,
                        "networkServiceType": "Chicago IX", 
                        "locationDetail": {
                            "name": "Test Location CHI",
                            "city": "Chicago",
                            "metro": "Chicago",
                            "country": "USA"
                        }
                    }
                ],
                "associatedVxcs": []
            },
            {
                "productId": 8002,
                "productUid": "port-test-ix-002",
                "productName": "Test IX Port 2",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "companyUid": "company-test-123",
                "companyName": "Test Company",
                "associatedIxs": [
                    {
                        "productId": 9002,
                        "productUid": "ix-test-456",
                        "productName": "Test EXAMPLE IX",
                        "productType": "IX",
                        "provisioningStatus": "CONFIGURED",
                        "rateLimit": 500,
                        "locationId": 42,
                        "vlan": 200,
                        "asn": 65002,
                        "networkServiceType": "Chicago IX",
                        "locationDetail": {
                            "name": "Test Location CHI",
                            "city": "Chicago",
                            "metro": "Chicago",
                            "country": "USA"
                        }
                    }
                ],
                "associatedVxcs": []
            },
            {
                "productId": 8003,
                "productUid": "port-test-ix-003",
                "productName": "Test IX Port 3",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "companyUid": "company-test-123",
                "companyName": "Test Company",
                "associatedIxs": [
                    {
                        "productId": 9003,
                        "productUid": "ix-test-789",
                        "productName": "Inactive Test IX",
                        "productType": "IX",
                        "provisioningStatus": "CANCELLED",
                        "rateLimit": 2000,
                        "locationId": 44,
                        "vlan": 300,
                        "asn": 65003,
                        "networkServiceType": "Sydney IX",
                        "locationDetail": {
                            "name": "Test Location SYD",
                            "city": "Sydney",
                            "metro": "Sydney",
                            "country": "Australia"
                        }
                    }
                ],
                "associatedVxcs": []
            }
        ]
    }`

	// Set up mock handler for products list API
	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(productsResponse))
		if err != nil {
			suite.FailNowf("could not write response", "could not write response %v", err)
		}
	})

	// Test cases for ListIXs with different filters
	tests := []struct {
		name           string
		request        *ListIXsRequest
		expectedCount  int
		expectedIXUIDs []string
	}{
		{
			name:           "List all active IXs (default behavior)",
			request:        &ListIXsRequest{},
			expectedCount:  2,
			expectedIXUIDs: []string{"ix-test-123", "ix-test-456"},
		},
		{
			name:           "List all IXs including inactive",
			request:        &ListIXsRequest{IncludeInactive: true},
			expectedCount:  3,
			expectedIXUIDs: []string{"ix-test-123", "ix-test-456", "ix-test-789"},
		},
		{
			name:           "Filter by exact name",
			request:        &ListIXsRequest{Name: "Test IX 1"},
			expectedCount:  1,
			expectedIXUIDs: []string{"ix-test-123"},
		},
		{
			name:           "Filter by name contains",
			request:        &ListIXsRequest{NameContains: "EXAMPLE"},
			expectedCount:  1,
			expectedIXUIDs: []string{"ix-test-456"},
		},
		{
			name:           "Filter by status",
			request:        &ListIXsRequest{Status: []string{"CONFIGURED"}},
			expectedCount:  1,
			expectedIXUIDs: []string{"ix-test-456"},
		},
		{
			name:           "Filter by ASN",
			request:        &ListIXsRequest{ASN: 65001},
			expectedCount:  1,
			expectedIXUIDs: []string{"ix-test-123"},
		},
		{
			name:           "Filter by VLAN",
			request:        &ListIXsRequest{VLAN: 200},
			expectedCount:  1,
			expectedIXUIDs: []string{"ix-test-456"},
		},
		{
			name:           "Filter by network service type",
			request:        &ListIXsRequest{NetworkServiceType: "Chicago IX"},
			expectedCount:  1,
			expectedIXUIDs: []string{"ix-test-456"},
		},
		{
			name:           "Filter by location ID",
			request:        &ListIXsRequest{LocationID: 42},
			expectedCount:  2,
			expectedIXUIDs: []string{"ix-test-123", "ix-test-456"},
		},
		{
			name:           "Filter by rate limit",
			request:        &ListIXsRequest{RateLimit: 1000},
			expectedCount:  1,
			expectedIXUIDs: []string{"ix-test-123"},
		},
		{
			name:           "Multiple filters (LocationID and NetworkServiceType)",
			request:        &ListIXsRequest{LocationID: 42, NetworkServiceType: "Chicago IX"},
			expectedCount:  1,
			expectedIXUIDs: []string{"ix-test-456"},
		},
		{
			name:           "Filter with no matches",
			request:        &ListIXsRequest{Name: "Non-existent IX"},
			expectedCount:  0,
			expectedIXUIDs: []string{},
		},
	}

	// Run all test cases
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			// Execute the test
			result, err := suite.client.IXService.ListIXs(context.Background(), tc.request)

			// Assert success
			suite.NoError(err, "Expected no error")
			suite.Equal(tc.expectedCount, len(result), "Expected %d IXs, got %d", tc.expectedCount, len(result))

			// Verify the correct IXs were returned
			var actualUIDs []string
			for _, ix := range result {
				actualUIDs = append(actualUIDs, ix.ProductUID)
			}

			// Verify each expected UID is in the result
			for _, expectedUID := range tc.expectedIXUIDs {
				suite.Contains(actualUIDs, expectedUID, "Expected IX with UID %s in results", expectedUID)
			}
		})
	}
}

// TestListIXsDeduplication tests that duplicate IXs are properly deduplicated
func (suite *IXClientTestSuite) TestListIXsDeduplication() {
	// Define mock response with duplicated IX
	productsResponse := `{
        "message": "Found 3 Products",
        "terms": "This data is subject to the Acceptable Use Policy",
        "data": [
            {
                "productId": 8101,
                "productUid": "port-test-ix-101",
                "productName": "Port With Duplicate IX 1",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "associatedIxs": [
                    {
                        "productId": 9101,
                        "productUid": "ix-test-duplicate",
                        "productName": "Duplicated IX",
                        "productType": "IX",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 1000,
                        "locationId": 50,
                        "vlan": 400,
                        "asn": 65101,
                        "networkServiceType": "Tokyo IX",
                        "locationDetail": {
                            "name": "Test Location TYO",
                            "city": "Tokyo",
                            "metro": "Tokyo",
                            "country": "Japan"
                        }
                    }
                ]
            },
            {
                "productId": 8102,
                "productUid": "port-test-ix-102",
                "productName": "Port With Duplicate IX 2",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "associatedIxs": [
                    {
                        "productId": 9101,
                        "productUid": "ix-test-duplicate",
                        "productName": "Duplicated IX",
                        "productType": "IX",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 1000,
                        "locationId": 50,
                        "vlan": 400,
                        "asn": 65101,
                        "networkServiceType": "Tokyo IX",
                        "locationDetail": {
                            "name": "Test Location TYO",
                            "city": "Tokyo",
                            "metro": "Tokyo",
                            "country": "Japan"
                        }
                    }
                ]
            },
            {
                "productId": 8103,
                "productUid": "port-test-ix-103",
                "productName": "Port With Duplicate IX 3",
                "productType": "MEGAPORT",
                "provisioningStatus": "LIVE",
                "associatedIxs": [
                    {
                        "productId": 9101,
                        "productUid": "ix-test-duplicate",
                        "productName": "Duplicated IX",
                        "productType": "IX",
                        "provisioningStatus": "LIVE",
                        "rateLimit": 1000,
                        "locationId": 50,
                        "vlan": 400,
                        "asn": 65101,
                        "networkServiceType": "Tokyo IX",
                        "locationDetail": {
                            "name": "Test Location TYO",
                            "city": "Tokyo",
                            "metro": "Tokyo",
                            "country": "Japan"
                        }
                    }
                ]
            }
        ]
    }`

	// Set up mock handler
	suite.mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(productsResponse))
		if err != nil {
			suite.FailNowf("could not write response", "could not write response %v", err)
		}
	})

	// Execute the test
	result, err := suite.client.IXService.ListIXs(context.Background(), nil)

	// Assert success and deduplication
	suite.NoError(err, "Expected no error")
	suite.Equal(1, len(result), "Expected exactly 1 IX after deduplication")

	if len(result) == 1 {
		suite.Equal("ix-test-duplicate", result[0].ProductUID, "Expected the duplicated IX")
		suite.Equal("Duplicated IX", result[0].ProductName, "Expected correct IX name")
		suite.Equal(65101, result[0].ASN, "Expected correct IX ASN")
	}
}

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

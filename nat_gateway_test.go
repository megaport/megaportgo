package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// NATGatewayClientTestSuite tests the NAT Gateway Service Client.
type NATGatewayClientTestSuite struct {
	ClientTestSuite
}

func TestNATGatewayClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(NATGatewayClientTestSuite))
}

func (suite *NATGatewayClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *NATGatewayClientTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *NATGatewayClientTestSuite) TestCreateNATGateway() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	jblob := `{
		"message": "Data returned successfully",
		"terms": "This data is subject to the Acceptable Use Policy",
		"data": {
			"adminLocked": false,
			"autoRenewTerm": true,
			"config": {
				"asn": 64512,
				"bgpShutdownDefault": false,
				"diversityZone": "red",
				"sessionCount": 100
			},
			"contractEndDate": "2024-10-01T14:34:56Z",
			"createDate": "2023-10-01T14:34:56Z",
			"createdBy": "user-name",
			"locationId": 123456,
			"locked": false,
			"orderApprovalStatus": "PENDING",
			"productName": "NAT Gateway",
			"productUid": "e900d0d5-1030-4e29-b2d8-816ad4263190",
			"promoCode": "PROMO123",
			"provisioningStatus": "DESIGN",
			"resourceTags": [{"key": "env", "value": "test"}],
			"serviceLevelReference": "SLR-1",
			"speed": 1000,
			"term": 1
		}
	}`

	suite.mux.HandleFunc("/v3/products/nat_gateways", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	gw, err := natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		AutoRenewTerm: true,
		Config: NATGatewayNetworkConfig{
			ASN:                64512,
			BGPShutdownDefault: false,
			DiversityZone:      "red",
			SessionCount:       100,
		},
		LocationID:            123456,
		ProductName:           "NAT Gateway",
		PromoCode:             "PROMO123",
		ResourceTags:          []ResourceTag{{Key: "env", Value: "test"}},
		ServiceLevelReference: "SLR-1",
		Speed:                 1000,
		Term:                  1,
	})
	suite.NoError(err)
	suite.Equal("e900d0d5-1030-4e29-b2d8-816ad4263190", gw.ProductUID)
	suite.Equal("NAT Gateway", gw.ProductName)
	suite.Equal(1000, gw.Speed)
	suite.Equal(1, gw.Term)
	suite.Equal(123456, gw.LocationID)
	suite.True(gw.AutoRenewTerm)
	suite.Equal(64512, gw.Config.ASN)
	suite.False(gw.Config.BGPShutdownDefault)
	suite.Equal("red", gw.Config.DiversityZone)
	suite.Equal(100, gw.Config.SessionCount)
	suite.Equal("DESIGN", gw.ProvisioningStatus)
	suite.Equal("PENDING", gw.OrderApprovalStatus)
	suite.Len(gw.ResourceTags, 1)
	suite.Equal("env", gw.ResourceTags[0].Key)
	suite.Equal("test", gw.ResourceTags[0].Value)
}

func (suite *NATGatewayClientTestSuite) TestCreateNATGatewayValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	// Missing ProductName
	_, err := natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		LocationID: 1, Speed: 1000, Term: 1,
	})
	suite.ErrorIs(err, ErrNATGatewayProductNameRequired)

	// Invalid LocationID
	_, err = natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		ProductName: "test", LocationID: 0, Speed: 1000, Term: 1,
	})
	suite.ErrorIs(err, ErrNATGatewayLocationIDRequired)

	// Invalid Speed
	_, err = natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		ProductName: "test", LocationID: 1, Speed: 0, Term: 1,
	})
	suite.ErrorIs(err, ErrNATGatewaySpeedRequired)

	// Invalid Term
	_, err = natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		ProductName: "test", LocationID: 1, Speed: 1000, Term: 5,
	})
	suite.ErrorIs(err, ErrNATGatewayInvalidTerm)
}

func (suite *NATGatewayClientTestSuite) TestListNATGateways() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	jblob := `{
		"message": "Data returned successfully",
		"terms": "This data is subject to the Acceptable Use Policy",
		"data": [
			{
				"adminLocked": false,
				"autoRenewTerm": true,
				"config": {"asn": 64512, "bgpShutdownDefault": false, "diversityZone": "red", "sessionCount": 100},
				"locationId": 123456,
				"productName": "NAT Gateway 1",
				"productUid": "uid-1",
				"provisioningStatus": "LIVE",
				"speed": 1000,
				"term": 12
			},
			{
				"adminLocked": false,
				"autoRenewTerm": false,
				"config": {"asn": 64513, "bgpShutdownDefault": true, "diversityZone": "blue", "sessionCount": 200},
				"locationId": 789012,
				"productName": "NAT Gateway 2",
				"productUid": "uid-2",
				"provisioningStatus": "DESIGN",
				"speed": 2500,
				"term": 24
			}
		]
	}`

	suite.mux.HandleFunc("/v3/products/nat_gateways", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	gateways, err := natSvc.ListNATGateways(ctx)
	suite.NoError(err)
	suite.Len(gateways, 2)

	suite.Equal("uid-1", gateways[0].ProductUID)
	suite.Equal("NAT Gateway 1", gateways[0].ProductName)
	suite.Equal(1000, gateways[0].Speed)
	suite.Equal("LIVE", gateways[0].ProvisioningStatus)

	suite.Equal("uid-2", gateways[1].ProductUID)
	suite.Equal("NAT Gateway 2", gateways[1].ProductName)
	suite.Equal(2500, gateways[1].Speed)
	suite.True(gateways[1].Config.BGPShutdownDefault)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewaysEmpty() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	jblob := `{
		"message": "Data returned successfully",
		"terms": "This data is subject to the Acceptable Use Policy",
		"data": []
	}`

	suite.mux.HandleFunc("/v3/products/nat_gateways", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	gateways, err := natSvc.ListNATGateways(ctx)
	suite.NoError(err)
	suite.Empty(gateways)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGateway() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "e900d0d5-1030-4e29-b2d8-816ad4263190"

	jblob := `{
		"message": "Data returned successfully",
		"terms": "This data is subject to the Acceptable Use Policy",
		"data": {
			"adminLocked": false,
			"autoRenewTerm": true,
			"config": {"asn": 64512, "bgpShutdownDefault": false, "diversityZone": "red", "sessionCount": 100},
			"contractEndDate": "2024-10-01T14:34:56Z",
			"createDate": "2023-10-01T14:34:56Z",
			"createdBy": "user-name",
			"locationId": 123456,
			"locked": false,
			"orderApprovalStatus": "APPROVED",
			"productName": "NAT Gateway",
			"productUid": "e900d0d5-1030-4e29-b2d8-816ad4263190",
			"provisioningStatus": "LIVE",
			"speed": 1000,
			"term": 12
		}
	}`

	path := fmt.Sprintf("/v3/products/nat_gateways/%s", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	gw, err := natSvc.GetNATGateway(ctx, productUID)
	suite.NoError(err)
	suite.Equal(productUID, gw.ProductUID)
	suite.Equal("NAT Gateway", gw.ProductName)
	suite.Equal("LIVE", gw.ProvisioningStatus)
	suite.Equal("APPROVED", gw.OrderApprovalStatus)
	suite.Equal(12, gw.Term)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	_, err := natSvc.GetNATGateway(ctx, "")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)
}

func (suite *NATGatewayClientTestSuite) TestUpdateNATGateway() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "e900d0d5-1030-4e29-b2d8-816ad4263190"

	jblob := `{
		"message": "Data returned successfully",
		"terms": "This data is subject to the Acceptable Use Policy",
		"data": {
			"adminLocked": false,
			"autoRenewTerm": false,
			"config": {"asn": 64512, "bgpShutdownDefault": true, "diversityZone": "blue", "sessionCount": 200},
			"locationId": 123456,
			"productName": "Updated NAT Gateway",
			"productUid": "e900d0d5-1030-4e29-b2d8-816ad4263190",
			"provisioningStatus": "LIVE",
			"speed": 1000,
			"term": 24
		}
	}`

	path := fmt.Sprintf("/v3/products/nat_gateways/%s", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPut, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	gw, err := natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductUID:    productUID,
		AutoRenewTerm: false,
		Config: NATGatewayNetworkConfig{
			ASN:                64512,
			BGPShutdownDefault: true,
			DiversityZone:      "blue",
			SessionCount:       200,
		},
		LocationID:  123456,
		ProductName: "Updated NAT Gateway",
		Speed:       1000,
		Term:        24,
	})
	suite.NoError(err)
	suite.Equal("Updated NAT Gateway", gw.ProductName)
	suite.Equal(24, gw.Term)
	suite.True(gw.Config.BGPShutdownDefault)
	suite.Equal("blue", gw.Config.DiversityZone)
	suite.Equal(200, gw.Config.SessionCount)
}

func (suite *NATGatewayClientTestSuite) TestUpdateNATGatewayValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	// Missing ProductUID
	_, err := natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductName: "test", LocationID: 1, Speed: 1000, Term: 1,
	})
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	// Missing ProductName
	_, err = natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductUID: "uid", LocationID: 1, Speed: 1000, Term: 1,
	})
	suite.ErrorIs(err, ErrNATGatewayProductNameRequired)

	// Invalid Term
	_, err = natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductUID: "uid", ProductName: "test", LocationID: 1, Speed: 1000, Term: 7,
	})
	suite.ErrorIs(err, ErrNATGatewayInvalidTerm)
}

func (suite *NATGatewayClientTestSuite) TestDeleteNATGatewayDesignState() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "e900d0d5-1030-4e29-b2d8-816ad4263190"

	// Pre-flight GET — DeleteNATGateway inspects ProvisioningStatus to pick
	// the right endpoint. DESIGN-state gateways use the nat-gateway-specific
	// DELETE path.
	getPath := fmt.Sprintf("/v3/products/nat_gateways/%s", productUID)
	var designDeleteCalled atomic.Bool
	suite.mux.HandleFunc(getPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, `{"message":"","terms":"","data":{"productUid":"%s","provisioningStatus":"DESIGN"}}`, productUID)
		case http.MethodDelete:
			designDeleteCalled.Store(true)
			fmt.Fprint(w, `{"message":"Nat gateway order item deleted successfully","terms":""}`)
		default:
			suite.FailNowf("unexpected method", "unexpected %s on %s", r.Method, getPath)
		}
	})

	err := natSvc.DeleteNATGateway(ctx, productUID)
	suite.NoError(err)
	suite.True(designDeleteCalled.Load(), "expected DESIGN-state delete to hit /v3/products/nat_gateways/{uid}")
}

func (suite *NATGatewayClientTestSuite) TestDeleteNATGatewayProvisioned() {
	// DeleteNATGateway must route every non-DESIGN status through the
	// generic product cancel endpoint. Covering each provisioned status
	// explicitly guards against accidental regressions where the routing
	// check is narrowed to a single status (e.g. STATUS_LIVE only).
	cases := []struct {
		name   string
		status string
	}{
		{name: "deployable", status: "DEPLOYABLE"},
		{name: "configured", status: SERVICE_CONFIGURED},
		{name: "live", status: SERVICE_LIVE},
	}

	for i, tc := range cases {
		i, tc := i, tc // capture per-iteration copies for go <1.22 loopclosure safety
		suite.Run(tc.name, func() {
			// Each sub-test uses a distinct UID so the handler paths do not
			// collide on the shared mux, avoiding the server leak that would
			// occur from calling suite.SetupTest() per sub-test.
			ctx := context.Background()
			natSvc := suite.client.NATGatewayService
			productUID := fmt.Sprintf("c1a2b3c4-d5e6-7890-1234-%012d", i+1)

			getPath := fmt.Sprintf("/v3/products/nat_gateways/%s", productUID)
			suite.mux.HandleFunc(getPath, func(w http.ResponseWriter, r *http.Request) {
				suite.Equal(http.MethodGet, r.Method, "provisioned gateway must not hit DESIGN DELETE endpoint")
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"message":"","terms":"","data":{"productUid":"%s","provisioningStatus":"%s"}}`, productUID, tc.status)
			})

			cancelPath := "/v3/product/" + productUID + "/action/CANCEL_NOW"
			var cancelCalled atomic.Bool
			suite.mux.HandleFunc(cancelPath, func(w http.ResponseWriter, r *http.Request) {
				suite.Equal(http.MethodPost, r.Method)
				cancelCalled.Store(true)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"message":"Action [CANCEL_NOW Service %s] has been done.","terms":""}`, productUID)
			})

			err := natSvc.DeleteNATGateway(ctx, productUID)
			suite.NoError(err)
			suite.True(cancelCalled.Load(), "expected provisioned delete to hit /v3/product/{uid}/action/CANCEL_NOW")
		})
	}
}

func (suite *NATGatewayClientTestSuite) TestDeleteNATGatewayValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	err := natSvc.DeleteNATGateway(ctx, "")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewaySessions() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	jblob := `{
		"message": "Success",
		"terms": "https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"sessionCount": [1000, 2000, 4000],
				"speedMbps": 100
			},
			{
				"sessionCount": [8000, 16000],
				"speedMbps": 1000
			}
		]
	}`

	suite.mux.HandleFunc("/v3/products/nat_gateways/sessions", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	sessions, err := natSvc.ListNATGatewaySessions(ctx)
	suite.NoError(err)
	suite.Len(sessions, 2)

	suite.Equal(100, sessions[0].SpeedMbps)
	suite.Equal([]int{1000, 2000, 4000}, sessions[0].SessionCount)

	suite.Equal(1000, sessions[1].SpeedMbps)
	suite.Equal([]int{8000, 16000}, sessions[1].SessionCount)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewaySessionsEmpty() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	jblob := `{
		"message": "Success",
		"terms": "https://www.megaport.com/legal/acceptable-use-policy",
		"data": []
	}`

	suite.mux.HandleFunc("/v3/products/nat_gateways/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	sessions, err := natSvc.ListNATGatewaySessions(ctx)
	suite.NoError(err)
	suite.Empty(sessions)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayTelemetry() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	jblob := `{
		"serviceUid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"type": "BITS",
		"timeFrame": {"from": 1608516536000, "to": 1608603936000},
		"data": [
			{
				"type": "BITS",
				"subtype": "IN",
				"samples": [[1608516536000, 125.5], [1608517536000, 130.2]],
				"unit": {"name": "Mbps", "fullName": "Megabits per second"}
			}
		]
	}`

	path := fmt.Sprintf("/v3/products/nat_gateways/%s/telemetry", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		suite.Equal("7", r.URL.Query().Get("days"))
		suite.Equal([]string{"BITS"}, r.URL.Query()["type"])
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	resp, err := natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: productUID,
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](7),
	})
	suite.NoError(err)
	suite.Equal(productUID, resp.ServiceUID)
	suite.Equal("BITS", resp.Type)
	suite.Equal(int64(1608516536000), resp.TimeFrame.From)
	suite.Equal(int64(1608603936000), resp.TimeFrame.To)
	suite.Len(resp.Data, 1)
	suite.Equal("BITS", resp.Data[0].Type)
	suite.Equal("IN", resp.Data[0].Subtype)
	suite.Len(resp.Data[0].Samples, 2)
	suite.Equal(int64(1608516536000), resp.Data[0].Samples[0].Timestamp)
	suite.Equal(125.5, resp.Data[0].Samples[0].Value)
	suite.Equal(int64(1608517536000), resp.Data[0].Samples[1].Timestamp)
	suite.Equal(130.2, resp.Data[0].Samples[1].Value)
	suite.Equal("Mbps", resp.Data[0].Unit.Name)
	suite.Equal("Megabits per second", resp.Data[0].Unit.FullName)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayTelemetryWithFromTo() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	jblob := `{
		"serviceUid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"type": "BITS",
		"timeFrame": {"from": 1608516536000, "to": 1608603936000},
		"data": []
	}`

	path := fmt.Sprintf("/v3/products/nat_gateways/%s/telemetry", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		suite.Equal("1608516536000", r.URL.Query().Get("from"))
		suite.Equal("1608603936000", r.URL.Query().Get("to"))
		suite.Equal([]string{"BITS", "PACKETS"}, r.URL.Query()["type"])
		suite.Empty(r.URL.Query().Get("days"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	fromTime := time.UnixMilli(1608516536000)
	toTime := time.UnixMilli(1608603936000)
	resp, err := natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: productUID,
		Types:      []string{"BITS", "PACKETS"},
		From:       &fromTime,
		To:         &toTime,
	})
	suite.NoError(err)
	suite.Equal(productUID, resp.ServiceUID)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayTelemetryValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	// Missing ProductUID
	_, err := natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		Types: []string{"BITS"},
		Days:  PtrTo[int32](7),
	})
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	// Missing Types
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Days:       PtrTo[int32](7),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryTypesRequired)

	// Days and From/To mutually exclusive
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](7),
		From:       PtrTo(time.UnixMilli(1608516536000)),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryTimeExclusive)

	// Days out of range (too low)
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](0),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryDaysOutOfRange)

	// Days out of range (too high)
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](181),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryDaysOutOfRange)

	// Only From without To
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		From:       PtrTo(time.UnixMilli(1608516536000)),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryFromToIncomplete)

	// Only To without From
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		To:         PtrTo(time.UnixMilli(1608603936000)),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryFromToIncomplete)
}

func (suite *NATGatewayClientTestSuite) TestValidateNATGatewayOrder() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "11111111-2222-3333-4444-555555555555"

	called := false
	suite.mux.HandleFunc("/v3/networkdesign/validate", func(w http.ResponseWriter, r *http.Request) {
		called = true
		suite.Equal(http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		suite.NoError(err)
		var payload []map[string]string
		suite.NoError(json.Unmarshal(body, &payload))
		suite.Equal([]map[string]string{{"productUid": productUID}}, payload)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"message":"Validation passed",
			"terms":"",
			"data":[{
				"productUid":%q,
				"productType":"NAT_GATEWAY",
				"string":"Sydney",
				"price":{
					"monthlyRate":600,
					"mbpsRate":0.6,
					"currency":"AUD",
					"productType":"NAT_GATEWAY",
					"monthlyRackRate":600
				}
			}]
		}`, productUID)
	})

	result, err := natSvc.ValidateNATGatewayOrder(ctx, productUID)
	suite.NoError(err)
	suite.True(called)
	suite.Equal(productUID, result.ProductUID)
	suite.Equal("NAT_GATEWAY", result.ProductType)
	suite.Equal("Sydney", result.Metro)
	suite.Equal(float64(600), result.Price.MonthlyRate)
	suite.Equal("AUD", result.Price.Currency)
}

func (suite *NATGatewayClientTestSuite) TestValidateNATGatewayOrder_EmptyData() {
	ctx := context.Background()
	productUID := "11111111-2222-3333-4444-555555555555"

	suite.mux.HandleFunc("/v3/networkdesign/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":[]}`)
	})

	_, err := suite.client.NATGatewayService.ValidateNATGatewayOrder(ctx, productUID)
	suite.ErrorIs(err, ErrNATGatewayOrderResponseEmpty)
}

func (suite *NATGatewayClientTestSuite) TestValidateNATGatewayOrder_MissingUID() {
	_, err := suite.client.NATGatewayService.ValidateNATGatewayOrder(context.Background(), "")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)
}

func (suite *NATGatewayClientTestSuite) TestBuyNATGateway() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "11111111-2222-3333-4444-555555555555"

	called := false
	suite.mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		called = true
		suite.Equal(http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		suite.NoError(err)
		var payload []map[string]string
		suite.NoError(json.Unmarshal(body, &payload))
		suite.Equal([]map[string]string{{"productUid": productUID}}, payload)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"message":"NAT_GATEWAY created",
			"terms":"",
			"data":[{
				"uid":%q,
				"name":"gw-name",
				"serviceName":"gw-name",
				"productType":"NAT_GATEWAY",
				"provisioningStatus":"DEPLOYABLE",
				"rateLimit":1000,
				"aLocationId":10,
				"contractTermMonths":1,
				"createDate":1776431685787
			}]
		}`, productUID)
	})

	result, err := natSvc.BuyNATGateway(ctx, productUID)
	suite.NoError(err)
	suite.True(called)
	suite.Equal(productUID, result.ProductUID)
	suite.Equal("DEPLOYABLE", result.ProvisioningStatus)
	suite.Equal(1000, result.RateLimit)
	suite.Equal(10, result.LocationID)
	suite.Equal(1, result.ContractTermMonths)
}

func (suite *NATGatewayClientTestSuite) TestBuyNATGateway_EmptyData() {
	ctx := context.Background()
	productUID := "11111111-2222-3333-4444-555555555555"

	suite.mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":[]}`)
	})

	_, err := suite.client.NATGatewayService.BuyNATGateway(ctx, productUID)
	suite.ErrorIs(err, ErrNATGatewayOrderResponseEmpty)
}

func (suite *NATGatewayClientTestSuite) TestBuyNATGateway_MissingUID() {
	_, err := suite.client.NATGatewayService.BuyNATGateway(context.Background(), "")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)
}

// --- Packet filters -------------------------------------------------------

func (suite *NATGatewayClientTestSuite) TestListNATGatewayPacketFilters() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "11111111-2222-3333-4444-555555555555"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/packet_filter_summaries", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":[{"id":1,"description":"first"},{"id":2,"description":"second"}]}`)
	})

	summaries, err := natSvc.ListNATGatewayPacketFilters(ctx, productUID)
	suite.NoError(err)
	suite.Len(summaries, 2)
	suite.Equal(1, summaries[0].ID)
	suite.Equal("first", summaries[0].Description)
	suite.Equal(2, summaries[1].ID)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewayPacketFiltersValidation() {
	_, err := suite.client.NATGatewayService.ListNATGatewayPacketFilters(context.Background(), "")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)
}

func (suite *NATGatewayClientTestSuite) TestCreateNATGatewayPacketFilter() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-create"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/packet_filters", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPost, r.Method)
		body, err := io.ReadAll(r.Body)
		suite.Require().NoError(err)
		var got NATGatewayPacketFilterRequest
		suite.Require().NoError(json.Unmarshal(body, &got))
		suite.Equal("permit-https", got.Description)
		suite.Len(got.Entries, 1)
		suite.Equal(PacketFilterActionPermit, got.Entries[0].Action)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":{"id":42,"description":"permit-https","entries":[{"action":"permit","sourceAddress":"0.0.0.0/0","destinationAddress":"10.0.0.0/24","destinationPorts":"443","ipProtocol":6}]}}`)
	})

	filter, err := natSvc.CreateNATGatewayPacketFilter(ctx, productUID, &NATGatewayPacketFilterRequest{
		Description: "permit-https",
		Entries: []NATGatewayPacketFilterEntry{
			{Action: PacketFilterActionPermit, SourceAddress: "0.0.0.0/0", DestinationAddress: "10.0.0.0/24", DestinationPorts: "443", IPProtocol: 6},
		},
	})
	suite.NoError(err)
	suite.Equal(42, filter.ID)
	suite.Equal("permit-https", filter.Description)
	suite.Len(filter.Entries, 1)
	suite.Equal(6, filter.Entries[0].IPProtocol)
}

func (suite *NATGatewayClientTestSuite) TestCreateNATGatewayPacketFilterValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	_, err := natSvc.CreateNATGatewayPacketFilter(ctx, "", &NATGatewayPacketFilterRequest{Description: "x", Entries: []NATGatewayPacketFilterEntry{{Action: "permit"}}})
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	_, err = natSvc.CreateNATGatewayPacketFilter(ctx, "uid", &NATGatewayPacketFilterRequest{Entries: []NATGatewayPacketFilterEntry{{Action: "permit"}}})
	suite.ErrorIs(err, ErrNATGatewayPacketFilterDescriptionEmpty)

	_, err = natSvc.CreateNATGatewayPacketFilter(ctx, "uid", &NATGatewayPacketFilterRequest{Description: "x"})
	suite.ErrorIs(err, ErrNATGatewayPacketFilterEntriesEmpty)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayPacketFilter() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-get"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/packet_filters/7", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":{"id":7,"description":"d","entries":[{"action":"deny","sourceAddress":"1.1.1.1/32","destinationAddress":"2.2.2.2/32"}]}}`)
	})

	filter, err := natSvc.GetNATGatewayPacketFilter(ctx, productUID, 7)
	suite.NoError(err)
	suite.Equal(7, filter.ID)
	suite.Equal(PacketFilterActionDeny, filter.Entries[0].Action)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayPacketFilterValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	_, err := natSvc.GetNATGatewayPacketFilter(ctx, "", 1)
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	_, err = natSvc.GetNATGatewayPacketFilter(ctx, "uid", 0)
	suite.ErrorIs(err, ErrNATGatewayPacketFilterIDRequired)
}

func (suite *NATGatewayClientTestSuite) TestUpdateNATGatewayPacketFilter() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-upd"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/packet_filters/12", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPut, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":{"id":12,"description":"updated","entries":[{"action":"permit","sourceAddress":"0.0.0.0/0","destinationAddress":"0.0.0.0/0"}]}}`)
	})

	filter, err := natSvc.UpdateNATGatewayPacketFilter(ctx, productUID, 12, &NATGatewayPacketFilterRequest{
		Description: "updated",
		Entries:     []NATGatewayPacketFilterEntry{{Action: PacketFilterActionPermit, SourceAddress: "0.0.0.0/0", DestinationAddress: "0.0.0.0/0"}},
	})
	suite.NoError(err)
	suite.Equal("updated", filter.Description)
}

func (suite *NATGatewayClientTestSuite) TestDeleteNATGatewayPacketFilter() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-del"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/packet_filters/9", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodDelete, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"deleted","terms":""}`)
	})

	err := natSvc.DeleteNATGatewayPacketFilter(ctx, productUID, 9)
	suite.NoError(err)
}

func (suite *NATGatewayClientTestSuite) TestDeleteNATGatewayPacketFilterValidation() {
	err := suite.client.NATGatewayService.DeleteNATGatewayPacketFilter(context.Background(), "", 1)
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	err = suite.client.NATGatewayService.DeleteNATGatewayPacketFilter(context.Background(), "uid", 0)
	suite.ErrorIs(err, ErrNATGatewayPacketFilterIDRequired)
}

// --- Prefix lists ---------------------------------------------------------

func (suite *NATGatewayClientTestSuite) TestListNATGatewayPrefixLists() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-list"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/prefix_list_summaries", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":[{"id":1,"description":"v4-list","addressFamily":"IPv4"}]}`)
	})

	summaries, err := natSvc.ListNATGatewayPrefixLists(ctx, productUID)
	suite.NoError(err)
	suite.Len(summaries, 1)
	suite.Equal(AddressFamilyIPv4, summaries[0].AddressFamily)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewayPrefixListsValidation() {
	_, err := suite.client.NATGatewayService.ListNATGatewayPrefixLists(context.Background(), "")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)
}

func (suite *NATGatewayClientTestSuite) TestCreateNATGatewayPrefixList() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-pl-create"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/prefix_lists", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPost, r.Method)
		body, err := io.ReadAll(r.Body)
		suite.Require().NoError(err)
		// Verify request body uses string ge/le on the wire.
		var got map[string]interface{}
		suite.Require().NoError(json.Unmarshal(body, &got))
		entries, ok := got["entries"].([]interface{})
		suite.Require().True(ok)
		suite.Require().Len(entries, 1)
		entry, ok := entries[0].(map[string]interface{})
		suite.Require().True(ok)
		suite.Equal("24", entry["ge"])
		suite.Equal("32", entry["le"])
		w.Header().Set("Content-Type", "application/json")
		// API returns ge/le as strings.
		fmt.Fprint(w, `{"message":"ok","terms":"","data":{"id":11,"description":"private","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":"24","le":"32"}]}}`)
	})

	pl, err := natSvc.CreateNATGatewayPrefixList(ctx, productUID, &NATGatewayPrefixList{
		Description:   "private",
		AddressFamily: AddressFamilyIPv4,
		Entries: []NATGatewayPrefixListEntry{
			{Action: PrefixListActionPermit, Prefix: "10.0.0.0/8", Ge: 24, Le: 32},
		},
	})
	suite.NoError(err)
	suite.Equal(11, pl.ID)
	suite.Equal(24, pl.Entries[0].Ge)
	suite.Equal(32, pl.Entries[0].Le)
}

func (suite *NATGatewayClientTestSuite) TestCreateNATGatewayPrefixListValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	_, err := natSvc.CreateNATGatewayPrefixList(ctx, "", &NATGatewayPrefixList{Description: "x", AddressFamily: "IPv4", Entries: []NATGatewayPrefixListEntry{{Action: "permit", Prefix: "0.0.0.0/0"}}})
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	_, err = natSvc.CreateNATGatewayPrefixList(ctx, "uid", &NATGatewayPrefixList{AddressFamily: "IPv4", Entries: []NATGatewayPrefixListEntry{{Action: "permit"}}})
	suite.ErrorIs(err, ErrNATGatewayPrefixListDescriptionEmpty)

	_, err = natSvc.CreateNATGatewayPrefixList(ctx, "uid", &NATGatewayPrefixList{Description: "x", Entries: []NATGatewayPrefixListEntry{{Action: "permit"}}})
	suite.ErrorIs(err, ErrNATGatewayPrefixListAddressFamilyEmpty)

	_, err = natSvc.CreateNATGatewayPrefixList(ctx, "uid", &NATGatewayPrefixList{Description: "x", AddressFamily: "IPv4"})
	suite.ErrorIs(err, ErrNATGatewayPrefixListEntriesEmpty)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayPrefixList() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-pl-get"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/prefix_lists/3", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		// Entry omits ge/le entirely (zero-value case).
		fmt.Fprint(w, `{"message":"ok","terms":"","data":{"id":3,"description":"v6","addressFamily":"IPv6","entries":[{"action":"deny","prefix":"::/0"}]}}`)
	})

	pl, err := natSvc.GetNATGatewayPrefixList(ctx, productUID, 3)
	suite.NoError(err)
	suite.Equal(AddressFamilyIPv6, pl.AddressFamily)
	suite.Equal(0, pl.Entries[0].Ge)
	suite.Equal(0, pl.Entries[0].Le)
	suite.Equal(PrefixListActionDeny, pl.Entries[0].Action)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayPrefixListInvalidGe() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-pl-bad"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/prefix_lists/4", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":{"id":4,"description":"x","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8","ge":"not-a-number"}]}}`)
	})

	_, err := natSvc.GetNATGatewayPrefixList(ctx, productUID, 4)
	suite.Error(err)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayPrefixListValidation() {
	_, err := suite.client.NATGatewayService.GetNATGatewayPrefixList(context.Background(), "", 1)
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	_, err = suite.client.NATGatewayService.GetNATGatewayPrefixList(context.Background(), "uid", 0)
	suite.ErrorIs(err, ErrNATGatewayPrefixListIDRequired)
}

func (suite *NATGatewayClientTestSuite) TestUpdateNATGatewayPrefixList() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-pl-upd"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/prefix_lists/5", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPut, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":{"id":5,"description":"updated","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"172.16.0.0/12"}]}}`)
	})

	pl, err := natSvc.UpdateNATGatewayPrefixList(ctx, productUID, 5, &NATGatewayPrefixList{
		Description:   "updated",
		AddressFamily: AddressFamilyIPv4,
		Entries:       []NATGatewayPrefixListEntry{{Action: PrefixListActionPermit, Prefix: "172.16.0.0/12"}},
	})
	suite.NoError(err)
	suite.Equal("updated", pl.Description)
}

func (suite *NATGatewayClientTestSuite) TestDeleteNATGatewayPrefixList() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-pl-del"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/prefix_lists/8", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodDelete, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"deleted","terms":""}`)
	})

	err := natSvc.DeleteNATGatewayPrefixList(ctx, productUID, 8)
	suite.NoError(err)
}

func (suite *NATGatewayClientTestSuite) TestDeleteNATGatewayPrefixListValidation() {
	err := suite.client.NATGatewayService.DeleteNATGatewayPrefixList(context.Background(), "", 1)
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	err = suite.client.NATGatewayService.DeleteNATGatewayPrefixList(context.Background(), "uid", 0)
	suite.ErrorIs(err, ErrNATGatewayPrefixListIDRequired)
}

// --- Diagnostics ----------------------------------------------------------

func (suite *NATGatewayClientTestSuite) TestListNATGatewayIPRoutesAsync() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-diag-ip"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/diagnostics/routes/ip", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		suite.Equal("10.0.0.1", r.URL.Query().Get("ip_address"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":"op-uuid-123"}`)
	})

	opID, err := natSvc.ListNATGatewayIPRoutesAsync(ctx, productUID, "10.0.0.1")
	suite.NoError(err)
	suite.Equal("op-uuid-123", opID)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewayBGPRoutesAsync() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-diag-bgp"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/diagnostics/routes/bgp", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		// No filter when ipAddress is empty.
		suite.Empty(r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":"op-uuid-bgp"}`)
	})

	opID, err := natSvc.ListNATGatewayBGPRoutesAsync(ctx, productUID, "")
	suite.NoError(err)
	suite.Equal("op-uuid-bgp", opID)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewayBGPNeighborRoutesAsyncValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	_, err := natSvc.ListNATGatewayBGPNeighborRoutesAsync(ctx, &NATGatewayBGPNeighborRoutesRequest{})
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	_, err = natSvc.ListNATGatewayBGPNeighborRoutesAsync(ctx, &NATGatewayBGPNeighborRoutesRequest{ProductUID: "uid"})
	suite.ErrorIs(err, ErrNATGatewayDiagnosticsPeerIPRequired)

	_, err = natSvc.ListNATGatewayBGPNeighborRoutesAsync(ctx, &NATGatewayBGPNeighborRoutesRequest{ProductUID: "uid", PeerIPAddress: "10.0.0.2"})
	suite.ErrorIs(err, ErrNATGatewayDiagnosticsDirectionInvalid)

	_, err = natSvc.ListNATGatewayBGPNeighborRoutesAsync(ctx, &NATGatewayBGPNeighborRoutesRequest{ProductUID: "uid", PeerIPAddress: "10.0.0.2", Direction: "INVALID"})
	suite.ErrorIs(err, ErrNATGatewayDiagnosticsDirectionInvalid)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewayBGPNeighborRoutesAsync() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-diag-nbr"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/diagnostics/routes/bgp/neighbor", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		suite.Equal(BGPRouteDirectionReceived, r.URL.Query().Get("direction"))
		suite.Equal("10.0.0.2", r.URL.Query().Get("peer_ip_address"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":"op-nbr"}`)
	})

	opID, err := natSvc.ListNATGatewayBGPNeighborRoutesAsync(ctx, &NATGatewayBGPNeighborRoutesRequest{
		ProductUID:    productUID,
		PeerIPAddress: "10.0.0.2",
		Direction:     BGPRouteDirectionReceived,
	})
	suite.NoError(err)
	suite.Equal("op-nbr", opID)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayDiagnosticsRoutesIP() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-op-ip"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/diagnostics/routes/operation", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal("op-1", r.URL.Query().Get("operationId"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":[
			{"prefix":"10.0.0.0/24","protocol":"STATIC","nextHop":{"ip":"10.0.0.1","vxc":{"id":"vxc-1","name":"vxc-name"}}}
		]}`)
	})

	routes, err := natSvc.GetNATGatewayDiagnosticsRoutes(ctx, productUID, "op-1")
	suite.NoError(err)
	suite.Len(routes, 1)
	suite.Require().NotNil(routes[0].IP)
	suite.Nil(routes[0].BGP)
	suite.Equal("10.0.0.0/24", routes[0].IP.Prefix)
	suite.Equal("STATIC", routes[0].IP.Protocol)
	suite.Equal("vxc-1", routes[0].IP.NextHop.VXC.ID)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayDiagnosticsRoutesBGP() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-op-bgp"

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/diagnostics/routes/operation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":[
			{"prefix":"192.168.0.0/16","asPath":"65000 65001","origin":"IGP","localPref":100,"best":true,"nextHop":{"ip":"10.0.0.2","vxc":{"id":"vxc-2","name":"v2"}}}
		]}`)
	})

	routes, err := natSvc.GetNATGatewayDiagnosticsRoutes(ctx, productUID, "op-2")
	suite.NoError(err)
	suite.Len(routes, 1)
	suite.Nil(routes[0].IP)
	suite.Require().NotNil(routes[0].BGP)
	suite.Equal("192.168.0.0/16", routes[0].BGP.Prefix)
	suite.Equal("65000 65001", routes[0].BGP.ASPath)
	suite.Equal(100, routes[0].BGP.LocalPref)
	suite.True(routes[0].BGP.Best)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayDiagnosticsRoutesValidation() {
	_, err := suite.client.NATGatewayService.GetNATGatewayDiagnosticsRoutes(context.Background(), "", "op")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	_, err = suite.client.NATGatewayService.GetNATGatewayDiagnosticsRoutes(context.Background(), "uid", "")
	suite.ErrorIs(err, ErrNATGatewayDiagnosticsOperationEmpty)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewayIPRoutesPolling() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "uid-poll"

	var opCalls atomic.Int32

	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/diagnostics/routes/ip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":"op-poll"}`)
	})
	suite.mux.HandleFunc("/v3/products/nat_gateways/"+productUID+"/diagnostics/routes/operation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// First call returns empty (still processing); subsequent calls return data.
		if opCalls.Add(1) == 1 {
			fmt.Fprint(w, `{"message":"ok","terms":"","data":[]}`)
			return
		}
		fmt.Fprint(w, `{"message":"ok","terms":"","data":[
			{"prefix":"10.0.0.0/24","protocol":"STATIC","nextHop":{"ip":"10.0.0.1","vxc":{"id":"vxc-1","name":"v1"}}},
			{"prefix":"192.168.0.0/16","asPath":"65000","origin":"IGP","best":true,"nextHop":{"ip":"10.0.0.2","vxc":{"id":"vxc-2","name":"v2"}}}
		]}`)
	})

	// Bypass the long polling defaults by polling directly via the async + Get methods,
	// so this test stays fast. The poll timeout/interval are package-level constants
	// and not worth plumbing through a setter just for testing.
	opID, err := natSvc.ListNATGatewayIPRoutesAsync(ctx, productUID, "")
	suite.Require().NoError(err)
	suite.Equal("op-poll", opID)

	// Drain the empty response then the populated one.
	routes, err := natSvc.GetNATGatewayDiagnosticsRoutes(ctx, productUID, opID)
	suite.Require().NoError(err)
	suite.Empty(routes)

	routes, err = natSvc.GetNATGatewayDiagnosticsRoutes(ctx, productUID, opID)
	suite.Require().NoError(err)
	suite.Len(routes, 2)

	// Discriminator: one IP, one BGP.
	var ipCount, bgpCount int
	for _, r := range routes {
		if r.IP != nil {
			ipCount++
		}
		if r.BGP != nil {
			bgpCount++
		}
	}
	suite.Equal(1, ipCount)
	suite.Equal(1, bgpCount)
}

// --- Prefix list round-trip ----------------------------------------------

func (suite *NATGatewayClientTestSuite) TestPrefixListGeLeRoundTrip() {
	pl := &NATGatewayPrefixList{
		Description:   "rt",
		AddressFamily: AddressFamilyIPv4,
		Entries: []NATGatewayPrefixListEntry{
			{Action: PrefixListActionPermit, Prefix: "10.0.0.0/8", Ge: 24, Le: 32},
			{Action: PrefixListActionDeny, Prefix: "0.0.0.0/0"}, // ge/le omitted
		},
	}

	api := pl.toAPI()
	suite.Equal("24", api.Entries[0].Ge)
	suite.Equal("32", api.Entries[0].Le)
	suite.Equal("", api.Entries[1].Ge)
	suite.Equal("", api.Entries[1].Le)

	back, err := api.toPrefixList()
	suite.Require().NoError(err)
	suite.Equal(24, back.Entries[0].Ge)
	suite.Equal(32, back.Entries[0].Le)
	suite.Equal(0, back.Entries[1].Ge)
	suite.Equal(0, back.Entries[1].Le)
}
